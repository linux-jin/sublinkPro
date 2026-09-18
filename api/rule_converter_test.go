package api

import (
	"strings"
	"testing"
)

func TestDetectTemplateTypeRecognizesLoonBeforeSurgeCore(t *testing.T) {
	template := "[General]\n[Proxy]\n[Proxy Group]\n[Remote Filter]\n"
	if got := detectTemplateType(template); got != "loon" {
		t.Fatalf("detectTemplateType() = %q, want loon", got)
	}
}

func TestGenerateLoonProxyGroupsUsesRemoteFiltersAndAllPlaceholder(t *testing.T) {
	got := generateLoonProxyGroups([]ACLProxyGroup{
		{Name: "Japan", Type: "url-test", Filter: "(JP|Japan)", URL: "https://example.com/generate_204", Interval: 600, Tolerance: 100},
		{Name: "All", Type: "select", IncludeAll: true},
	}, true)

	for _, want := range []string{
		"[Remote Filter]",
		`__SublinkPro_Filter_1 = NameRegex,SublinkPro,FilterKey="JP|Japan"`,
		"[Proxy Group]",
		"Japan = url-test,__SublinkPro_Filter_1,url=https://example.com/generate_204,interval=600,tolerance=100",
		"All = select,__ALL_PROXIES__",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generateLoonProxyGroups() missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateLoonRulesSeparatesLocalAndRemoteRules(t *testing.T) {
	got, err := generateLoonRules([]ACLRuleset{
		{Group: "Direct", RuleURL: "[]DOMAIN-SUFFIX,example.com"},
		{Group: "Proxy", RuleURL: "https://example.invalid/proxy.list,3600"},
		{Group: "Final", RuleURL: "[]MATCH"},
	}, false, false, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"[Rule]",
		"DOMAIN-SUFFIX,example.com,Direct",
		"FINAL,Final",
		"[Remote Rule]",
		"https://example.invalid/proxy.list, policy=Proxy, tag=SublinkPro-2-Proxy, enabled=true",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generateLoonRules() missing %q:\n%s", want, got)
		}
	}
}

func TestMergeLoonTemplatePreservesUnmanagedSections(t *testing.T) {
	template := `[General]
loglevel=notify
[Proxy]
Old=direct
[Remote Filter]
OldFilter=NameRegex,old,FilterKey=".*"
[Proxy Group]
Old=select,DIRECT
[Rule]
FINAL,Old
[Remote Rule]
https://old.invalid, policy=Old
[Plugin]
https://example.invalid/plugin.plugin, enabled=true
[MITM]
hostname=example.com
`
	got := mergeLoonTemplate(template, "[Proxy Group]\nNew=select,__ALL_PROXIES__", "[Rule]\nFINAL,New")
	for _, want := range []string{"[Proxy]", "Old=direct", "[Plugin]", "[MITM]", "New=select,__ALL_PROXIES__", "FINAL,New"} {
		if !strings.Contains(got, want) {
			t.Fatalf("mergeLoonTemplate() missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"OldFilter=", "Old=select,DIRECT", "FINAL,Old", "https://old.invalid"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("mergeLoonTemplate() retained %q:\n%s", unwanted, got)
		}
	}
}

func TestGetDefaultTemplateLoonIsNativeProfile(t *testing.T) {
	got := getDefaultTemplate("loon")
	for _, want := range []string{"[Proxy]", "[Proxy Group]", "__ALL_PROXIES__", "[Rule]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("getDefaultTemplate(loon) missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, "proxies:") {
		t.Fatalf("getDefaultTemplate(loon) returned Clash YAML: %s", got)
	}
}

func TestGenerateLoonProxyGroupsAddsFallbackAndLoadBalanceOptions(t *testing.T) {
	got := generateLoonProxyGroups([]ACLProxyGroup{
		{Name: "Fallback", Type: "fallback", Proxies: []string{"A", "B"}},
		{Name: "Balance", Type: "load-balance", Proxies: []string{"A", "B"}},
	}, false)
	for _, want := range []string{
		"Fallback = fallback,A,B,url=http://www.gstatic.com/generate_204,interval=300,max-timeout=3000",
		"Balance = load-balance,A,B,url=http://www.gstatic.com/generate_204,interval=300,algorithm=PCC,max-timeout=3000",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generateLoonProxyGroups() missing %q:\n%s", want, got)
		}
	}
}
