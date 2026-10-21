package protocol

import (
	"encoding/json"
	"strings"
	"sublink/utils"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestVmessEncodeDecode 测试 VMess 编解码完整性
func TestVmessEncodeDecode(t *testing.T) {
	original := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Aid:  "0",
		Net:  "ws",
		Type: "none",
		Host: "cdn.example.com",
		Path: "/vmess",
		Tls:  "tls",
		Sni:  "sni.example.com",
		Alpn: "h2,http/1.1",
		Fp:   "chrome",
		Scy:  "auto",
		Ps:   "测试节点-VMess",
		V:    "2",
	}

	// 编码
	encoded := EncodeVmessURL(original)
	if !strings.HasPrefix(encoded, "vmess://") {
		t.Errorf("编码后应以 vmess:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeVMESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Add", original.Add, decoded.Add)
	assertEqualString(t, "Id", original.Id, decoded.Id)
	assertEqualString(t, "Net", original.Net, decoded.Net)
	assertEqualString(t, "Path", original.Path, decoded.Path)
	assertEqualString(t, "Ps(名称)", original.Ps, decoded.Ps)
	assertEqualString(t, "Sni", original.Sni, decoded.Sni)
	assertEqualString(t, "Alpn", original.Alpn, decoded.Alpn)
	assertEqualString(t, "Fp", original.Fp, decoded.Fp)
	assertEqualString(t, "Scy", original.Scy, decoded.Scy)
	assertEqualString(t, "Tls", original.Tls, decoded.Tls)

	t.Logf("✓ VMess 编解码测试通过，名称: %s", decoded.Ps)
}

func TestVmessBuildProxyPreservesTLSMetadata(t *testing.T) {
	vmess := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Net:  "ws",
		Path: "/vmess",
		Host: "cdn.example.com",
		Tls:  "tls",
		Sni:  "sni.example.com",
		Alpn: "h2,http/1.1",
		Fp:   "chrome",
		Scy:  "auto",
		Ps:   "测试节点-VMess-TLS",
		V:    "2",
	}

	proxy, err := buildVMessProxy(Urls{Url: EncodeVmessURL(vmess)}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildVMessProxy 失败: %v", err)
	}

	assertEqualString(t, "ClientFingerprint", vmess.Fp, proxy.Client_fingerprint)
	assertEqualString(t, "Servername", vmess.Sni, proxy.Servername)
	assertEqualString(t, "ALPN0", "h2", proxy.Alpn[0])
	assertEqualString(t, "ALPN1", "http/1.1", proxy.Alpn[1])
}

// TestVmessNameModification 测试 VMess 名称修改
func TestVmessNameModification(t *testing.T) {
	original := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Net:  "tcp",
		Ps:   "原始名称",
		V:    "2",
	}

	newName := "新名称-VMess-测试"
	encoded := EncodeVmessURL(original)
	decoded, _ := DecodeVMESSURL(encoded)
	decoded.Ps = newName
	reEncoded := EncodeVmessURL(decoded)
	final, _ := DecodeVMESSURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Ps)
	assertEqualString(t, "服务器(不变)", original.Add, final.Add)
	assertEqualString(t, "UUID(不变)", original.Id, final.Id)

	t.Logf("✓ VMess 名称修改测试通过: %s -> %s", original.Ps, final.Ps)
}

// TestVmessSpecialCharacters 测试 VMess 特殊字符
func TestVmessSpecialCharacters(t *testing.T) {
	specialNames := []string{
		"节点 with spaces",
		"节点-with-dashes",
		"节点_with_underscores",
		"节点中文测试",
		"Node🚀Emoji",
		"Node (parentheses)",
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			original := Vmess{
				Add:  "example.com",
				Port: "443",
				Id:   "12345678-1234-1234-1234-123456789abc",
				Net:  "tcp",
				Ps:   name,
				V:    "2",
			}

			encoded := EncodeVmessURL(original)
			decoded, err := DecodeVMESSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "特殊字符名称", name, decoded.Ps)
			t.Logf("✓ 特殊字符测试通过: %s", name)
		})
	}
}

// TestVmessClashWSHeadersRoundTrip 验证 Clash ws-opts.headers 中 Host 以外的请求头（如 User-Agent）经链接往返后不丢失
func TestVmessClashWSHeadersRoundTrip(t *testing.T) {
	const ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	const doc = `
name: vmess-ws-ua
type: vmess
server: example.com
port: 443
uuid: 12345678-1234-1234-1234-123456789abc
alterId: 0
cipher: auto
network: ws
ws-opts:
  path: /1234
  headers:
    Host: abc.com
    User-Agent: ` + ua + `
`
	var proxy Proxy
	if err := yaml.Unmarshal([]byte(doc), &proxy); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	link := EncodeVmessURL(ConvertProxyToVmess(proxy))
	decoded, err := DecodeVMESSURL(link)
	if err != nil {
		t.Fatalf("DecodeVMESSURL failed: %v", err)
	}
	assertEqualString(t, "Host", "abc.com", decoded.Host)

	out, err := LinkToProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy failed: %v", err)
	}
	assertEqualString(t, "Path", "/1234", mustString(t, "Path", out.Ws_opts["path"]))
	headers := mustMap(t, "Ws_opts headers", out.Ws_opts["headers"])
	assertEqualString(t, "Host", "abc.com", mustString(t, "Host", headers["Host"]))
	assertEqualString(t, "User-Agent", ua, mustString(t, "User-Agent", headers["User-Agent"]))
}

// Links carrying an object-valued extension were accepted before headers became
// a modeled string field. Keep accepting them and preserve usable WS headers.
func TestVmessHeadersInputCompatibility(t *testing.T) {
	for _, tt := range []struct {
		name    string
		headers any
		wantUA  string
	}{
		{"object", map[string]any{"User-Agent": "compat-client", "X-Test": "1"}, "compat-client"},
		{"string", `{"User-Agent":"compat-client"}`, "compat-client"},
		{"null", nil, ""},
		{"unrecognized numeric extension", 42, ""},
		{"unrecognized array extension", []string{"ignored"}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"add": "example.com", "port": 443, "id": "12345678-1234-1234-1234-123456789abc",
				"net": "ws", "host": "cdn.example.com", "headers": tt.headers,
			})
			if err != nil {
				t.Fatal(err)
			}
			link := "vmess://" + utils.Base64Encode(string(body))
			decoded, err := DecodeVMESSURL(link)
			if err != nil {
				t.Fatalf("previously accepted link rejected: %v", err)
			}
			out, err := LinkToProxy(Urls{Url: EncodeVmessURL(decoded)}, OutputConfig{})
			if err != nil {
				t.Fatal(err)
			}
			headers := mustMap(t, "WS headers", out.Ws_opts["headers"])
			assertEqualString(t, "Host", "cdn.example.com", mustString(t, "Host", headers["Host"]))
			if tt.wantUA != "" {
				assertEqualString(t, "User-Agent", tt.wantUA, mustString(t, "User-Agent", headers["User-Agent"]))
			} else if headers["User-Agent"] != nil {
				t.Fatalf("unexpected header from ignored extension: %v", headers)
			}
		})
	}
}
