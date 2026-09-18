package protocol

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sublink/cache"

	regexp "github.com/dlclark/regexp2/v2/compat"
)

// EncodeLoon converts supported node links to native Loon proxy lines and
// merges them into a complete Loon profile template.
func EncodeLoon(urls []Urls, config OutputConfig) (string, error) {
	proxyLines := make([]string, 0, len(urls))
	proxyNames := make([]string, 0, len(urls))
	for _, item := range urls {
		proxy, err := LinkToProxy(item, config)
		if err != nil {
			log.Printf("Loon 节点转换失败: %v", err)
			continue
		}
		line, err := buildLoonProxyLine(proxy, config)
		if err != nil {
			log.Printf("Loon 不支持节点 %q: %v", proxy.Name, err)
			continue
		}
		if strings.TrimSpace(line) == "" || strings.TrimSpace(proxy.Name) == "" {
			continue
		}
		proxyLines = append(proxyLines, line)
		proxyNames = append(proxyNames, proxy.Name)
	}
	return DecodeLoon(proxyLines, proxyNames, config.Loon)
}

func buildLoonProxyLine(proxy Proxy, config OutputConfig) (string, error) {
	proxy.Server = replaceLoonHost(proxy.Server, config)
	switch strings.ToLower(strings.TrimSpace(proxy.Type)) {
	case "ss":
		return buildLoonSSLine(proxy)
	case "ssr":
		return buildLoonSSRLine(proxy)
	case "vmess":
		return buildLoonVMessLine(proxy)
	case "vless":
		return buildLoonVLESSLine(proxy)
	case "trojan":
		return buildLoonTrojanLine(proxy)
	case "anytls":
		return buildLoonAnyTLSLine(proxy)
	case "hysteria2":
		return buildLoonHysteria2Line(proxy)
	case "http", "https":
		return buildLoonHTTPLine(proxy)
	case "socks5":
		return buildLoonSocks5Line(proxy)
	case "wireguard":
		return buildLoonWireGuardLine(proxy)
	default:
		return "", fmt.Errorf("unsupported protocol %s", proxy.Type)
	}
}

func replaceLoonHost(server string, config OutputConfig) string {
	if config.ReplaceServerWithHost && len(config.HostMap) > 0 {
		if ip, ok := config.HostMap[server]; ok {
			return ip
		}
	}
	return server
}

func buildLoonSSLine(proxy Proxy) (string, error) {
	if proxy.Cipher == "" {
		return "", fmt.Errorf("missing cipher")
	}
	line := fmt.Sprintf("%s=shadowsocks,%s,%d,%s,%s", proxy.Name, proxy.Server, proxy.Port.Int(), proxy.Cipher, quoteLoon(proxy.Password))
	if proxy.Plugin == "obfs" || proxy.Plugin == "obfs-local" || proxy.Plugin == "simple-obfs" {
		line = appendLoonString(line, "obfs-name", loonMapString(proxy.Plugin_opts, "mode", "obfs"), false)
		line = appendLoonString(line, "obfs-host", loonMapString(proxy.Plugin_opts, "host", "obfs-host"), false)
		line = appendLoonString(line, "obfs-uri", loonMapString(proxy.Plugin_opts, "path"), false)
	}
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonSSRLine(proxy Proxy) (string, error) {
	line := fmt.Sprintf("%s=shadowsocksr,%s,%d,%s,%s", proxy.Name, proxy.Server, proxy.Port.Int(), proxy.Cipher, quoteLoon(proxy.Password))
	line = appendLoonString(line, "protocol", proxy.Protocol, false)
	line = appendLoonString(line, "protocol-param", proxy.ProtocolParam, false)
	line = appendLoonString(line, "obfs", proxy.Obfs, false)
	line = appendLoonString(line, "obfs-param", proxy.ObfsParam, false)
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonVMessLine(proxy Proxy) (string, error) {
	security := strings.TrimSpace(proxy.Cipher)
	if security == "" {
		security = "auto"
	}
	line := fmt.Sprintf("%s=vmess,%s,%d,%s,%s", proxy.Name, proxy.Server, proxy.Port.Int(), security, quoteLoon(proxy.Uuid))
	var err error
	line, err = appendLoonV2RayTransport(line, proxy)
	if err != nil {
		return "", err
	}
	line = appendLoonBoolValue(line, "over-tls", proxy.Tls)
	line = appendLoonTLS(line, proxy)
	alterID := strings.TrimSpace(proxy.AlterId)
	if alterID == "" {
		alterID = "0"
	}
	line += ",alterId=" + alterID
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonVLESSLine(proxy Proxy) (string, error) {
	if proxy.Encryption != "" && proxy.Encryption != "none" {
		return "", fmt.Errorf("VLESS encryption %s is not supported", proxy.Encryption)
	}
	if proxy.Flow != "" && proxy.Flow != "xtls-rprx-vision" {
		return "", fmt.Errorf("VLESS flow %s is not supported", proxy.Flow)
	}
	line := fmt.Sprintf("%s=vless,%s,%d,%s", proxy.Name, proxy.Server, proxy.Port.Int(), quoteLoon(proxy.Uuid))
	var err error
	line, err = appendLoonV2RayTransport(line, proxy)
	if err != nil {
		return "", err
	}
	line = appendLoonBoolValue(line, "over-tls", proxy.Tls || len(proxy.Reality_opts) > 0 || proxy.Flow != "")
	line = appendLoonString(line, "flow", proxy.Flow, false)
	line = appendLoonTLS(line, proxy)
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonTrojanLine(proxy Proxy) (string, error) {
	line := fmt.Sprintf("%s=trojan,%s,%d,%s", proxy.Name, proxy.Server, proxy.Port.Int(), quoteLoon(proxy.Password))
	if proxy.Network != "" && proxy.Network != "tcp" {
		switch proxy.Network {
		case "ws":
			line += ",transport=ws"
			line = appendLoonString(line, "path", loonMapString(proxy.Ws_opts, "path"), false)
			line = appendLoonString(line, "host", loonNestedMapString(proxy.Ws_opts, "headers", "Host"), false)
		case "http", "h2":
			line += ",transport=http"
			line = appendLoonString(line, "path", loonFirstMapString(proxy.Http_opts, "path"), false)
			line = appendLoonString(line, "host", loonNestedFirstMapString(proxy.Http_opts, "headers", "Host"), false)
		default:
			return "", fmt.Errorf("Trojan network %s is not supported", proxy.Network)
		}
	}
	line = appendLoonTLS(line, proxy)
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonAnyTLSLine(proxy Proxy) (string, error) {
	if proxy.Network != "" && proxy.Network != "tcp" {
		return "", fmt.Errorf("AnyTLS network %s is not supported", proxy.Network)
	}
	line := fmt.Sprintf("%s=anytls,%s,%d,%s", proxy.Name, proxy.Server, proxy.Port.Int(), quoteLoon(proxy.Password))
	if proxy.AnyTLSIdleTimeout > 0 {
		line += fmt.Sprintf(",idle-session-timeout=%d", proxy.AnyTLSIdleTimeout)
	}
	line = appendLoonTLS(line, proxy)
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonHysteria2Line(proxy Proxy) (string, error) {
	port := proxy.Port.Int()
	if port <= 0 {
		port = firstLoonPort(proxy.Ports)
	}
	if port <= 0 {
		port = 443
	}
	line := fmt.Sprintf("%s=Hysteria2,%s,%d,%s", proxy.Name, proxy.Server, port, quoteLoon(proxy.Password))
	if strings.TrimSpace(proxy.Ports) != "" {
		line += ",server-ports=" + quoteLoon(strings.ReplaceAll(proxy.Ports, "/", ","))
	}
	line = appendLoonString(line, "sni", firstNonEmpty(proxy.Sni, proxy.Servername), false)
	line = appendLoonBoolValue(line, "skip-cert-verify", proxy.Skip_cert_verify)
	line = appendLoonString(line, "tls-cert-sha256", proxy.Fingerprint, false)
	if proxy.Obfs == "salamander" && proxy.Obfs_password != "" {
		line = appendLoonString(line, "salamander-password", proxy.Obfs_password, false)
	}
	if proxy.Down > 0 {
		line += fmt.Sprintf(",download-bandwidth=%d", proxy.Down)
	}
	line = appendLoonBool(line, "fast-open", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonHTTPLine(proxy Proxy) (string, error) {
	typeName := "http"
	if proxy.Tls || strings.EqualFold(proxy.Type, "https") {
		typeName = "https"
	}
	line := fmt.Sprintf("%s=%s,%s,%d", proxy.Name, typeName, proxy.Server, proxy.Port.Int())
	if proxy.Username != "" {
		line += "," + proxy.Username
		if proxy.Password != "" {
			line += "," + quoteLoon(proxy.Password)
		}
	}
	line = appendLoonString(line, "sni", firstNonEmpty(proxy.Sni, proxy.Servername), false)
	line = appendLoonBoolValue(line, "skip-cert-verify", proxy.Skip_cert_verify)
	line = appendLoonBool(line, "tfo", proxy.Tfo)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonSocks5Line(proxy Proxy) (string, error) {
	line := fmt.Sprintf("%s=socks5,%s,%d", proxy.Name, proxy.Server, proxy.Port.Int())
	if proxy.Username != "" {
		line += "," + proxy.Username
		if proxy.Password != "" {
			line += "," + quoteLoon(proxy.Password)
		}
	}
	line = appendLoonBoolValue(line, "over-tls", proxy.Tls)
	line = appendLoonString(line, "sni", firstNonEmpty(proxy.Sni, proxy.Servername), false)
	line = appendLoonBoolValue(line, "skip-cert-verify", proxy.Skip_cert_verify)
	line = appendLoonBool(line, "tfo", proxy.Tfo)
	line = appendLoonBool(line, "udp", proxy.Udp)
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func buildLoonWireGuardLine(proxy Proxy) (string, error) {
	if proxy.Private_key == "" || proxy.Public_key == "" {
		return "", fmt.Errorf("missing WireGuard key")
	}
	line := proxy.Name + "=wireguard"
	line = appendLoonString(line, "interface-ip", proxy.Ip, false)
	line = appendLoonString(line, "interface-ipV6", proxy.Ipv6, false)
	line = appendLoonString(line, "private-key", proxy.Private_key, true)
	if proxy.Mtu > 0 {
		line += fmt.Sprintf(",mtu=%d", proxy.Mtu)
	}
	allowed := "0.0.0.0/0,::/0"
	if len(proxy.Allowed_ips) > 0 {
		allowed = strings.Join(proxy.Allowed_ips, ",")
	}
	peer := fmt.Sprintf("{public-key=%s,allowed-ips=%s,endpoint=%s:%d", quoteLoon(proxy.Public_key), quoteLoon(allowed), proxy.Server, proxy.Port.Int())
	if len(proxy.Reserved) > 0 {
		parts := make([]string, 0, len(proxy.Reserved))
		for _, value := range proxy.Reserved {
			parts = append(parts, strconv.Itoa(value))
		}
		peer += ",reserved=[" + strings.Join(parts, ",") + "]"
	}
	if proxy.Pre_shared_key != "" {
		peer += ",preshared-key=" + quoteLoon(proxy.Pre_shared_key)
	}
	line += ",peers=[" + peer + "}]"
	return appendLoonIPMode(line, proxy.Ip_version), nil
}

func appendLoonV2RayTransport(line string, proxy Proxy) (string, error) {
	network := strings.TrimSpace(proxy.Network)
	if network == "" || network == "tcp" {
		return line + ",transport=tcp", nil
	}
	switch network {
	case "ws":
		line += ",transport=ws"
		line = appendLoonString(line, "path", loonMapString(proxy.Ws_opts, "path"), false)
		line = appendLoonString(line, "host", loonNestedMapString(proxy.Ws_opts, "headers", "Host"), false)
		return line, nil
	case "http", "h2":
		line += ",transport=http"
		line = appendLoonString(line, "path", loonFirstMapString(proxy.Http_opts, "path"), false)
		line = appendLoonString(line, "host", loonNestedFirstMapString(proxy.Http_opts, "headers", "Host"), false)
		return line, nil
	default:
		return "", fmt.Errorf("network %s is not supported", network)
	}
}

func appendLoonTLS(line string, proxy Proxy) string {
	line = appendLoonBoolValue(line, "skip-cert-verify", proxy.Skip_cert_verify)
	if profile := loonTLSProfile(proxy.Client_fingerprint); profile != "" {
		line = appendLoonString(line, "tls-profile", profile, false)
	}
	if len(proxy.Alpn) > 0 {
		line = appendLoonString(line, "alpn", strings.Join(proxy.Alpn, ","), true)
	}
	if len(proxy.Reality_opts) > 0 {
		line = appendLoonString(line, "sni", firstNonEmpty(proxy.Sni, proxy.Servername), false)
		line = appendLoonString(line, "public-key", loonMapString(proxy.Reality_opts, "public-key"), true)
		line = appendLoonString(line, "short-id", loonMapString(proxy.Reality_opts, "short-id"), false)
		return line
	}
	line = appendLoonString(line, "sni", firstNonEmpty(proxy.Sni, proxy.Servername), false)
	line = appendLoonString(line, "tls-cert-sha256", proxy.Fingerprint, false)
	return line
}

func loonTLSProfile(fingerprint string) string {
	switch strings.ToLower(strings.TrimSpace(fingerprint)) {
	case "chrome":
		return "chrome147"
	case "ios", "safari":
		return "ios26"
	case "default", "chrome147", "ios18", "ios26":
		return strings.ToLower(strings.TrimSpace(fingerprint))
	default:
		return ""
	}
}

func appendLoonBool(line, key string, enabled bool) string {
	if !enabled {
		return line
	}
	return line + "," + key + "=true"
}

func appendLoonBoolValue(line, key string, value bool) string {
	return line + "," + key + "=" + strconv.FormatBool(value)
}

func appendLoonString(line, key, value string, quote bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return line
	}
	if quote {
		value = quoteLoon(value)
	}
	return line + "," + key + "=" + value
}

func appendLoonIPMode(line, value string) string {
	modes := map[string]string{
		"dual": "dual", "ipv4": "v4-only", "ipv6": "v6-only",
		"ipv4-prefer": "prefer-v4", "ipv6-prefer": "prefer-v6",
	}
	value = strings.TrimSpace(value)
	if mapped := modes[value]; mapped != "" {
		value = mapped
	}
	return appendLoonString(line, "ip-mode", value, false)
}

func quoteLoon(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstLoonPort(value string) int {
	value = strings.TrimSpace(strings.ReplaceAll(value, "/", ","))
	if value == "" {
		return 0
	}
	first := strings.Split(value, ",")[0]
	first = strings.Split(first, "-")[0]
	port, _ := strconv.Atoi(strings.TrimSpace(first))
	return port
}

func loonMapString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if text := strings.TrimSpace(fmt.Sprint(value)); text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func loonNestedMapString(values map[string]any, parent, key string) string {
	child, ok := values[parent].(map[string]any)
	if !ok {
		return ""
	}
	return loonMapString(child, key)
}

func loonFirstMapString(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case []string:
		if len(typed) > 0 {
			return typed[0]
		}
	case []any:
		if len(typed) > 0 {
			return fmt.Sprint(typed[0])
		}
	default:
		return fmt.Sprint(value)
	}
	return ""
}

func loonNestedFirstMapString(values map[string]any, parent, key string) string {
	child, ok := values[parent].(map[string]any)
	if !ok {
		return ""
	}
	return loonFirstMapString(child, key)
}

type loonFilter struct {
	mode       string
	pattern    string
	compiled   *regexp.Regexp
	compileErr error
}

// DecodeLoon merges generated proxy lines into a complete Loon profile.
// Remote subscription/filter sections are consumed server-side and removed to
// avoid recursively requesting the same SublinkPro Loon endpoint.
func DecodeLoon(proxyLines, proxyNames []string, file string) (string, error) {
	template, err := loadLoonTemplate(file)
	if err != nil {
		return "", err
	}
	filters := parseLoonFilters(template)
	lines := strings.Split(template, "\n")
	result := make([]string, 0, len(lines)+len(proxyLines))
	currentSection := ""
	proxySectionFound := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isLoonSection(trimmed) {
			currentSection = trimmed
			switch currentSection {
			case "[Remote Proxy]", "[Remote Filter]":
				continue
			case "[Proxy]":
				proxySectionFound = true
				result = append(result, line)
				result = append(result, proxyLines...)
				continue
			default:
				result = append(result, line)
				continue
			}
		}

		switch currentSection {
		case "[Remote Proxy]", "[Remote Filter]", "[Proxy]":
			continue
		case "[Proxy Group]":
			if strings.Contains(line, "=") && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				line = expandLoonProxyGroup(line, proxyNames, filters)
			}
		}
		result = append(result, line)
	}

	if !proxySectionFound {
		result = append(result, "", "[Proxy]")
		result = append(result, proxyLines...)
	}
	return strings.TrimSpace(strings.Join(result, "\n")) + "\n", nil
}

func loadLoonTemplate(file string) (string, error) {
	if strings.TrimSpace(file) == "" {
		return "", fmt.Errorf("loon template is not configured")
	}
	if strings.Contains(file, "://") {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, file, nil)
		if err != nil {
			return "", err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
	filename := filepath.Base(file)
	if cached, ok := cache.GetTemplateContent(filename); ok {
		return cached, nil
	}
	body, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	cache.SetTemplateContent(filename, string(body))
	return string(body), nil
}

func parseLoonFilters(template string) map[string]loonFilter {
	filters := map[string]loonFilter{}
	currentSection := ""
	for _, line := range strings.Split(template, "\n") {
		trimmed := strings.TrimSpace(line)
		if isLoonSection(trimmed) {
			currentSection = trimmed
			continue
		}
		if currentSection != "[Remote Filter]" || trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		rhs := strings.TrimSpace(parts[1])
		mode := strings.ToLower(strings.TrimSpace(strings.SplitN(rhs, ",", 2)[0]))
		pattern := loonFilterKey(rhs)
		filter := loonFilter{mode: mode, pattern: pattern}
		if mode == "nameregex" && pattern != "" {
			filter.compiled, filter.compileErr = regexp.Compile(pattern)
		}
		filters[name] = filter
	}
	return filters
}

func loonFilterKey(value string) string {
	lower := strings.ToLower(value)
	idx := strings.Index(lower, "filterkey")
	if idx < 0 {
		return ""
	}
	value = value[idx+len("filterkey"):]
	if equals := strings.Index(value, "="); equals >= 0 {
		value = value[equals+1:]
	}
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func expandLoonProxyGroup(line string, proxyNames []string, filters map[string]loonFilter) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return line
	}
	groupName := strings.TrimSpace(parts[0])
	tokens := splitLoonCSV(parts[1])
	if len(tokens) == 0 {
		return line
	}
	groupType := strings.TrimSpace(tokens[0])
	expanded := []string{groupType}
	seen := map[string]bool{}
	proxyCount := 0
	for _, token := range tokens[1:] {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "=") {
			expanded = append(expanded, trimmed)
			continue
		}
		var candidates []string
		switch {
		case trimmed == "__ALL_PROXIES__":
			candidates = proxyNames
		case filters[trimmed].mode != "":
			candidates = matchLoonFilter(filters[trimmed], proxyNames)
		default:
			candidates = []string{trimmed}
		}
		for _, candidate := range candidates {
			if candidate == "" || seen[candidate] {
				continue
			}
			seen[candidate] = true
			expanded = append(expanded, candidate)
			proxyCount++
		}
	}
	if proxyCount == 0 && groupType != "ssid" {
		expanded = append(expanded, "DIRECT")
	}
	return groupName + " = " + strings.Join(expanded, ",")
}

func matchLoonFilter(filter loonFilter, names []string) []string {
	matches := make([]string, 0)
	for _, name := range names {
		matched := false
		switch filter.mode {
		case "nameregex":
			if filter.compileErr == nil && filter.compiled != nil {
				matched = filter.compiled.MatchString(name)
			}
		case "namekeyword":
			matched = strings.Contains(strings.ToLower(name), strings.ToLower(filter.pattern))
		case "nodeselect":
			// UI selections cannot be reconstructed after remote subscriptions are
			// removed. Expanding to all generated local nodes is safer than silently
			// routing an empty selection through DIRECT.
			matched = true
		}
		if matched {
			matches = append(matches, name)
		}
	}
	return matches
}

func splitLoonCSV(value string) []string {
	var tokens []string
	var current strings.Builder
	quote := rune(0)
	depth := 0
	for _, r := range value {
		switch {
		case quote != 0:
			current.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
			current.WriteRune(r)
		case r == '[' || r == '{' || r == '(':
			depth++
			current.WriteRune(r)
		case r == ']' || r == '}' || r == ')':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		case r == ',' && depth == 0:
			tokens = append(tokens, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	tokens = append(tokens, strings.TrimSpace(current.String()))
	return tokens
}

func isLoonSection(value string) bool {
	return strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")
}
