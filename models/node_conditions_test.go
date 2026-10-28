package models

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCompiledNodeConditionsPreserveLegacyEvaluation(t *testing.T) {
	nodes := []Node{
		{Name: "first", LinkName: "GB fast", Protocol: "vless", Group: "a", DelayTime: 80, DelayStatus: "success", Speed: 3, UnlockSummary: `{"providers":[{"provider":"openai","status":"available"},{"provider":"netflix","status":"restricted"}]}`},
		{Name: "second", LinkName: "GB slow", Protocol: "trojan", Group: "b", DelayTime: 400, DelayStatus: "timeout", UnlockSummary: `{"providers":[{"provider":"openai","status":"restricted"},{"provider":"netflix","status":"available"}]}`},
		{},
	}
	for _, raw := range []string{
		`{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":"vless"},{"field":"delay_time","operator":"less_than","value":"100"}]}`,
		`{"logic":"or","conditions":[{"field":"protocol","operator":"not_equals","value":"trojan"},{"field":"link_name","operator":"regex","value":"slow$"}]}`,
		`{"logic":"and","conditions":[{"field":"unlock_provider","operator":"equals","value":"openai"},{"field":"unlock_status","operator":"equals","value":"available"}]}`,
		`{"logic":"or","conditions":[{"field":"unlock_provider","operator":"equals","value":"openai"},{"field":"unlock_status","operator":"equals","value":"available"},{"field":"group","operator":"equals","value":"missing"}]}`,
	} {
		legacy, err := ParseConditions(raw)
		if err != nil {
			t.Fatal(err)
		}
		compiled, err := CompileNodeConditionMatcher(raw, NodeConditionSmartMember)
		if err != nil {
			t.Fatal(err)
		}
		for _, node := range nodes {
			if got, want := compiled(node), legacy.EvaluateNode(node); got != want {
				t.Fatalf("%s node=%s compiled=%v legacy=%v", raw, node.Name, got, want)
			}
		}
	}
	legacy, err := ParseConditions(`{"logic":"legacy-default","conditions":[{"field":"group","operator":"equals","value":"a"}]}`)
	if err != nil || !legacy.EvaluateNode(nodes[0]) {
		t.Fatalf("legacy parser semantics changed: %v", err)
	}
	legacy, err = ParseConditions(`{"logic":"and","conditions":[]}`)
	if err != nil || legacy.EvaluateNode(nodes[0]) {
		t.Fatal("legacy empty conditions must remain false")
	}
	disabled, err := CompileNodeConditionMatcher("", NodeConditionSmartCandidate)
	if err != nil || !disabled(Node{}) {
		t.Fatal("disabled additional filter must not restrict old groups")
	}
}

func TestSmartConditionValidationRejectsInvalidAndCheckDependentCandidateRules(t *testing.T) {
	invalid := []string{
		`null`, `{}`, `[]`, `{"logic":"and","conditions":[]}`, `{"logic":"xor","conditions":[{"field":"protocol","operator":"equals","value":"ss"}]}`,
		`{"logic":"and","conditions":[{"field":"unknown","operator":"not_equals","value":"x"}]}`,
		`{"logic":"and","conditions":[{"field":"protocol","operator":"unknown","value":"ss"}]}`,
		`{"logic":"and","conditions":[{"field":"protocol","operator":"regex","value":"["}]}`,
		`{"logic":"and","conditions":[{"field":"protocol","operator":"contains","value":""}]}`,
		`{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":null}]}`,
		`{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":{}}]}`,
		`{"logic":"and","conditions":[{"field":"delay_time","operator":"less_than","value":"NaN"}]}`,
		`{"logic":"and","conditions":[{"field":"delay_status","operator":"equals","value":"made-up"}]}`,
		`{"logic":"and","typo":true,"conditions":[{"field":"protocol","operator":"equals","value":"ss"}]}`,
		`{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":"ss"}]} null`,
		strings.Repeat("x", 16385),
	}
	for _, raw := range invalid {
		if _, err := CompileNodeConditionMatcher(raw, NodeConditionSmartMember); err == nil {
			t.Fatalf("accepted invalid rule: %s", raw)
		}
	}
	for _, field := range GetNodeConditionFields() {
		if isSmartCandidateField(field.Value) != slices.Contains(field.Surfaces, NodeConditionSmartCandidate) {
			t.Fatalf("metadata mismatch: %s", field.Value)
		}
		if !slices.Contains(field.Surfaces, NodeConditionSmartMember) {
			t.Fatalf("missing membership surface: %s", field.Value)
		}
	}
	for _, field := range []string{"speed", "delay_time", "delay_status", "quality_status", "unlock_status", "ip_type", "residential_type", "tags"} {
		raw, err := json.Marshal(NodeConditions{Logic: "and", Conditions: []NodeCondition{{Field: field, Operator: "equals", Value: "success"}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := CompileNodeConditionMatcher(string(raw), NodeConditionSmartCandidate); err == nil {
			t.Fatalf("check-dependent field allowed in candidates: %s", field)
		}
	}
}

func TestSmartConditionDynamicUnlockEnumsValidateRegisteredValues(t *testing.T) {
	for _, field := range []string{"unlock_provider", "unlock_status"} {
		raw := `{"logic":"and","conditions":[{"field":"` + field + `","operator":"not_equals","value":"made-up"}]}`
		if _, err := CompileNodeConditionMatcher(raw, NodeConditionSmartMember); err == nil {
			t.Fatalf("unknown dynamic enum accepted: %s", field)
		}
		group := SmartGroup{MemberConditions: raw}
		if got := group.CandidateExclusionReason(Node{DelayStatus: "success", DelayTime: 80, UnlockSummary: `{"providers":[{"provider":"openai","status":"available"}]}`}, time.Now()); got != "memberConditionsInvalid" {
			t.Fatalf("invalid restored enum admitted: %s", got)
		}
	}
	previous := unlockProviderListResolver
	t.Cleanup(func() { unlockProviderListResolver = previous })
	RegisterUnlockProviderListResolver(func() []string { return []string{"custom_checker"} })
	if _, err := CompileNodeConditionMatcher(`{"logic":"and","conditions":[{"field":"unlock_provider","operator":"equals","value":"custom_checker"}]}`, NodeConditionSmartMember); err != nil {
		t.Fatalf("registered custom provider rejected: %v", err)
	}
}
