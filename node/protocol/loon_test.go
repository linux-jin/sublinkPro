package protocol

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildLoonProxyLineSupportedProtocols(t *testing.T) {
	tests := []struct {
		name  string
		proxy Proxy
		want  string
	}{
		{name: "ss", proxy: Proxy{Name: "SS", Type: "ss", Server: "ss.example", Port: 443, Cipher: "aes-128-gcm", Password: "secret", Udp: true}, want: `SS=shadowsocks,ss.example,443,aes-128-gcm,"secret",udp=true`},
		{name: "ssr", proxy: Proxy{Name: "SSR", Type: "ssr", Server: "ssr.example", Port: 443, Cipher: "aes-256-cfb", Password: "secret", Protocol: "auth_sha1_v4", Obfs: "tls1.2_ticket_auth"}, want: `SSR=shadowsocksr,ssr.example,443,aes-256-cfb,"secret",protocol=auth_sha1_v4,obfs=tls1.2_ticket_auth`},
		{name: "vmess", proxy: Proxy{Name: "VMess", Type: "vmess", Server: "vmess.example", Port: 443, Uuid: "uuid", Network: "ws", Tls: true, Ws_opts: map[string]any{"path": "/ws", "headers": map[string]any{"Host": "cdn.example"}}, Sni: "tls.example"}, want: `VMess=vmess,vmess.example,443,auto,"uuid",transport=ws,path=/ws,host=cdn.example,over-tls=true`},
		{name: "vless", proxy: Proxy{Name: "VLESS", Type: "vless", Server: "vless.example", Port: 443, Uuid: "uuid", Network: "tcp", Tls: true}, want: `VLESS=vless,vless.example,443,"uuid",transport=tcp,over-tls=true`},
		{name: "trojan", proxy: Proxy{Name: "Trojan", Type: "trojan", Server: "trojan.example", Port: 443, Password: "secret", Network: "tcp", Sni: "tls.example"}, want: `Trojan=trojan,trojan.example,443,"secret",skip-cert-verify=false,sni=tls.example`},
		{name: "anytls", proxy: Proxy{Name: "AnyTLS", Type: "anytls", Server: "anytls.example", Port: 443, Password: "secret", Network: "tcp"}, want: `AnyTLS=anytls,anytls.example,443,"secret"`},
		{name: "hysteria2", proxy: Proxy{Name: "HY2", Type: "hysteria2", Server: "hy2.example", Port: 443, Password: "secret", Ports: "443-445"}, want: `HY2=Hysteria2,hy2.example,443,"secret",server-ports="443-445"`},
		{name: "http", proxy: Proxy{Name: "HTTP", Type: "http", Server: "http.example", Port: 8080, Username: "user", Password: "pass"}, want: `HTTP=http,http.example,8080,user,"pass"`},
		{name: "socks5", proxy: Proxy{Name: "SOCKS", Type: "socks5", Server: "socks.example", Port: 1080, Username: "user", Password: "pass", Udp: true}, want: `SOCKS=socks5,socks.example,1080,user,"pass",over-tls=false,skip-cert-verify=false,udp=true`},
		{name: "wireguard", proxy: Proxy{Name: "WG", Type: "wireguard", Server: "wg.example", Port: 51820, Private_key: "private", Public_key: "public", Ip: "10.0.0.2/32"}, want: `WG=wireguard,interface-ip=10.0.0.2/32,private-key="private",peers=[{public-key="public",allowed-ips="0.0.0.0/0,::/0",endpoint=wg.example:51820}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildLoonProxyLine(tt.proxy, OutputConfig{})
			if err != nil {
				t.Fatalf("buildLoonProxyLine() error = %v", err)
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("buildLoonProxyLine() = %q, want substring %q", got, tt.want)
			}
		})
	}
}

func TestBuildLoonProxyLineRejectsUnsupportedProtocol(t *testing.T) {
	if _, err := buildLoonProxyLine(Proxy{Name: "Unsupported", Type: "tuic"}, OutputConfig{}); err == nil {
		t.Fatal("expected unsupported protocol error")
	}
}

func TestDecodeLoonReplacesNodeSectionsAndPreservesProfile(t *testing.T) {
	template := `[General]
loglevel = notify

[Proxy]
Old=shadowsocks,old.example,443,aes-128-gcm,"old"

[Remote Proxy]
Remote=https://example.invalid/sub

[Remote Filter]
JP = NameRegex,Remote,FilterKey="^(?!.*Test).*JP.*$"
All = NameKeyword,Remote,FilterKey="Node"
Manual = NodeSelect,Remote

[Proxy Group]
Japan = select,JP
Everything = select,__ALL_PROXIES__
ManualGroup = select,Manual
Literal = select,NodeSelectFilter
WiFi = ssid,default=Everything,Home=Japan

[Remote Rule]
https://example.invalid/rules, policy=Japan, tag=remote, enabled=true

[Plugin]
https://example.invalid/plugin.plugin, enabled=true

[MITM]
hostname = example.com
`
	path := filepath.Join(t.TempDir(), "template.lcf")
	if err := os.WriteFile(path, []byte(template), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := DecodeLoon(
		[]string{`JP Node=shadowsocks,jp.example,443,aes-128-gcm,"secret"`, `US Node=shadowsocks,us.example,443,aes-128-gcm,"secret"`},
		[]string{"JP Node", "US Node"},
		path,
	)
	if err != nil {
		t.Fatalf("DecodeLoon() error = %v", err)
	}
	for _, unwanted := range []string{"[Remote Proxy]", "[Remote Filter]", "Old=shadowsocks", "https://example.invalid/sub"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("DecodeLoon() retained %q:\n%s", unwanted, got)
		}
	}
	for _, want := range []string{
		`JP Node=shadowsocks,jp.example,443,aes-128-gcm,"secret"`,
		`US Node=shadowsocks,us.example,443,aes-128-gcm,"secret"`,
		"Japan = select,JP Node",
		"Everything = select,JP Node,US Node",
		"ManualGroup = select,JP Node,US Node",
		"Literal = select,NodeSelectFilter",
		"WiFi = ssid,default=Everything,Home=Japan",
		"[Remote Rule]",
		"[Plugin]",
		"[MITM]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("DecodeLoon() missing %q:\n%s", want, got)
		}
	}
}

func TestDecodeLoonAddsProxySectionWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "minimal.lcf")
	if err := os.WriteFile(path, []byte("[General]\nloglevel=notify\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := DecodeLoon([]string{"Node=direct"}, []string{"Node"}, path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "[Proxy]\nNode=direct") {
		t.Fatalf("missing appended proxy section: %s", got)
	}
}

func TestSplitLoonCSVPreservesNestedValues(t *testing.T) {
	got := splitLoonCSV(`select,A,"B,C",peers=[{allowed-ips="0.0.0.0/0,::/0"}],url=https://example.com`)
	if len(got) != 5 {
		t.Fatalf("splitLoonCSV() len = %d, values = %#v", len(got), got)
	}
}

func TestPublicLoonTemplateIsSanitizedAndRenderable(t *testing.T) {
	path := filepath.Join("..", "..", "template", "loon.lcf")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read public Loon template: %v", err)
	}
	lower := strings.ToLower(string(content))
	for _, forbidden := range []string{
		"ca-p12",
		"ca-passphrase",
		"client-key-data",
		"private-key:",
		"[remote proxy]\nhttp",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("public Loon template contains forbidden private marker %q", forbidden)
		}
	}

	output, err := DecodeLoon(
		[]string{`Public Test=shadowsocks,127.0.0.1,443,aes-128-gcm,"test"`},
		[]string{"Public Test"},
		path,
	)
	if err != nil {
		t.Fatalf("render public Loon template: %v", err)
	}
	if !strings.Contains(output, "Public Test=shadowsocks") {
		t.Fatal("public Loon template did not receive generated proxy")
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "[Remote Proxy]" || line == "[Remote Filter]" {
			t.Fatal("public Loon output retained server-consumed remote sections")
		}
	}
}
