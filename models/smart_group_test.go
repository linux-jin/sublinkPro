package models

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"sublink/database"
)

func TestSmartGroupDynamicMembersAcrossOriginalGroups(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	old := time.Now().Add(-96 * time.Hour).Format("2006-01-02 15:04:05")
	healthy := func(name, group, country, tested string) Node {
		return createSubcriptionTestNode(t, Node{Name: name, LinkName: name, Group: group, LinkCountry: country,
			DelayStatus: "success", SpeedStatus: "success", DelayTime: 120, Speed: 4,
			LatencyCheckAt: tested, SpeedCheckAt: tested})
	}
	gb := healthy("gb-a", "original-a", "GB", now)
	de := healthy("de-b", "original-b", "DE", now)
	healthy("fr-expired", "original-c", "FR", old)
	healthy("us-other", "original-d", "US", now)
	createSubcriptionTestNode(t, Node{Name: "gb-failed", Group: "original-e", LinkCountry: "GB", DelayStatus: "timeout"})
	group := SmartGroup{Name: "Europe", Countries: "gb, fr, DE,GB", MaxDelay: 500, MinSpeed: 4, MaxAgeHours: 72}
	if err := group.Validate(); err != nil {
		t.Fatal(err)
	}
	if group.Countries != "GB,FR,DE" {
		t.Fatalf("normalized countries = %q", group.Countries)
	}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	candidates, err := group.CandidateIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 4 {
		t.Fatalf("candidate IDs = %v, want four regardless of test status", candidates)
	}
	members, err := group.Members()
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 || members[0].ID != gb.ID || members[1].ID != de.ID {
		t.Fatalf("members = %v, want GB and DE nodes", nodeNames(members))
	}
	if members[0].Group != "original-a" || members[1].Group != "original-b" {
		t.Fatalf("source groups changed: %q, %q", members[0].Group, members[1].Group)
	}
	sub := &Subcription{Name: "smart-only", SmartGroupIDs: strconv.Itoa(group.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil {
		t.Fatal(err)
	}
	if got := nodeNames(sub.Nodes); !reflect.DeepEqual(got, []string{"gb-a", "de-b"}) {
		t.Fatalf("smart-only subscription nodes = %v", got)
	}
}

func TestNormalizeSmartGroupIDsRejectsPartialInvalidInput(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	group := SmartGroup{Name: "valid", Countries: "GB"}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	id := strconv.Itoa(group.ID)
	for _, input := range []string{id + ",abc", "," + id, id + ",-1", id + ","} {
		if _, err := NormalizeSmartGroupIDs(input); err == nil {
			t.Errorf("expected invalid %q to be rejected", input)
		}
	}
	if got, err := NormalizeSmartGroupIDs(id + "," + id); err != nil || got != id {
		t.Errorf("deduplication = %q, %v", got, err)
	}
}
