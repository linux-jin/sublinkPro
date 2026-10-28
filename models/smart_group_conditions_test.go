package models

import (
	"reflect"
	"strconv"
	"sublink/database"
	"testing"
	"time"
)

func TestSmartGroupAdditionalConditionsKeepFailedCandidatesAndFilterMembers(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	makeNode := func(name, protocol, status string, score int) Node {
		return createSubcriptionTestNode(t, Node{Name: name, LinkName: name, Protocol: protocol, LinkCountry: "GB", Group: "original", DelayStatus: status, DelayTime: 80, QualityStatus: QualityStatusSuccess, FraudScore: score, Tags: "manual"})
	}
	pass := makeNode("pass", "vless", "success", 10)
	highRisk := makeNode("high-risk", "vless", "success", 90)
	failed := makeNode("failed", "vless", "timeout", 10)
	makeNode("outside", "trojan", "success", 10)
	group := SmartGroup{Name: "advanced", Countries: "GB", SourceGroups: []string{"ORIGINAL"}, CandidateConditions: `{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":"vless"}]}`, MemberConditions: `{"logic":"and","conditions":[{"field":"quality_status","operator":"equals","value":"success"},{"field":"fraud_score","operator":"less_or_equal","value":30}]}`}
	if err := group.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	candidates, members, stats, cutoff, err := group.EvaluateMembers()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 3 || len(members) != 1 || members[0].ID != pass.ID || stats.DelayUnusable != 1 || stats.MemberConditionMismatch != 1 {
		t.Fatalf("candidates=%d members=%v stats=%+v", len(candidates), nodeNames(members), stats)
	}
	if stats.CandidateCount != stats.DelayUnusable+stats.MemberConditionMismatch+len(members) {
		t.Fatal("exclusion counts inconsistent")
	}
	if group.CandidateMatcher()(failed) == "" || group.CandidateExclusionReason(highRisk, cutoff) != "memberConditionMismatch" {
		t.Fatal("member conditions affected detection scope or reason")
	}
	sub := &Subcription{Name: "advanced-sub", SmartGroupIDs: strconv.Itoa(group.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(nodeNames(sub.Nodes), []string{"pass"}) {
		t.Fatalf("subscription nodes=%v", nodeNames(sub.Nodes))
	}
	var saved Node
	if err := database.DB.First(&saved, pass.ID).Error; err != nil || saved.Tags != "manual" || saved.Group != "original" {
		t.Fatalf("node classification mutated: %+v %v", saved, err)
	}
	group.MemberConditions = `{"logic":"and","conditions":[{"field":"unlock_provider","operator":"equals","value":"openai"},{"field":"unlock_status","operator":"equals","value":"available"}]}`
	if nodes, err := group.CandidateIDs(); err != nil || len(nodes) != 3 {
		t.Fatalf("unmeasured unlock removed candidates: %v %v", nodes, err)
	}
	if got := group.CandidateExclusionReason(pass, time.Now()); got != "memberConditionMismatch" {
		t.Fatalf("unmeasured unlock accepted: %s", got)
	}
	group.MemberConditions = `{"broken":`
	_, members, stats, _, err = group.EvaluateMembers()
	if err != nil || len(members) != 0 || stats.MemberConditionsInvalid != 2 {
		t.Fatalf("invalid restored member rules failed open: %v %+v %v", members, stats, err)
	}
	group.CandidateConditions = `{"broken":`
	if _, err := group.Candidates(); err == nil {
		t.Fatal("invalid candidate conditions accepted")
	}
	if group.CandidateMatcher()(pass) != "" {
		t.Fatal("invalid matcher failed open")
	}
}

func TestSmartGroupAdditionalFiltersPreserveCountryNameAlternatives(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	node := createSubcriptionTestNode(t, Node{Name: "PH candidate", LinkName: "PH candidate", LinkCountry: "GB", Protocol: "ss", Group: "A"})
	group := SmartGroup{Name: "PH", Countries: "PH", SourceGroups: []string{"a"}, CandidateConditions: `{"logic":"or","conditions":[{"field":"protocol","operator":"equals","value":"ss"},{"field":"source","operator":"equals","value":"other"}]}`}
	if err := group.Validate(); err != nil {
		t.Fatal(err)
	}
	nodes, err := group.CandidateIDs()
	if err != nil || !reflect.DeepEqual(nodes, []int{node.ID}) || group.CandidateMatcher()(node) != "countryName" {
		t.Fatalf("country/name OR changed: %v %v", nodes, err)
	}
}
