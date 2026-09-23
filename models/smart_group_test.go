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

func TestSmartGroupKeywordSourceGroupsAndLatencyOnly(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	matched := createSubcriptionTestNode(t, Node{Name: "PH-Manila-A", LinkName: "PH-Manila-A", Group: "airport-a", LinkCountry: "PH",
		DelayStatus: "success", SpeedStatus: "untested", DelayTime: 240, Speed: 0, LatencyCheckAt: now})
	createSubcriptionTestNode(t, Node{Name: "PH-Manila-B", Group: "airport-b", LinkCountry: "PH", DelayStatus: "success", DelayTime: 120, LatencyCheckAt: now})
	cebu := createSubcriptionTestNode(t, Node{Name: "PH-Cebu", Group: "airport-a", LinkCountry: "PH", DelayStatus: "success", DelayTime: 90, LatencyCheckAt: now})
	createSubcriptionTestNode(t, Node{Name: "PH-Manila-failed", Group: "airport-a", LinkCountry: "PH", DelayStatus: "timeout"})
	gb := createSubcriptionTestNode(t, Node{Name: "GB-Manila", Group: "airport-a", LinkCountry: "GB", DelayStatus: "success", DelayTime: 110, LatencyCheckAt: now})

	group := SmartGroup{Name: "Philippines", Countries: "ph", Keyword: "  manila  ", SourceGroups: []string{" airport-a ", "AIRPORT-A"}, MaxAgeHours: 72}
	if err := group.Validate(); err != nil {
		t.Fatal(err)
	}
	if group.Keyword != "manila" || !reflect.DeepEqual(group.SourceGroups, []string{"airport-a"}) {
		t.Fatalf("normalized conditions: %+v", group)
	}
	ids, err := group.CandidateIDs()
	if err != nil || len(ids) != 4 {
		t.Fatalf("country OR keyword candidates = %v, %v", ids, err)
	}
	if group.MatchSource(gb) != "keyword" || group.MatchSource(cebu) != "landingCountry" {
		t.Fatalf("match sources: GB=%q Cebu=%q", group.MatchSource(gb), group.MatchSource(cebu))
	}
	members, err := group.Members()
	if err != nil || len(members) != 3 || members[0].ID != matched.ID || members[1].ID != cebu.ID || members[2].ID != gb.ID {
		t.Fatalf("latency-only members = %v, %v", nodeNames(members), err)
	}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	sub := &Subcription{Name: "ph-only", SmartGroupIDs: strconv.Itoa(group.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil || len(sub.Nodes) != 3 || sub.Nodes[0].ID != matched.ID || sub.Nodes[1].ID != cebu.ID || sub.Nodes[2].ID != gb.ID {
		t.Fatalf("filtered subscription nodes = %v, %v", nodeNames(sub.Nodes), err)
	}
	group.MinSpeed = 1
	members, err = group.Members()
	if err != nil || len(members) != 0 {
		t.Fatalf("speed-required members = %v, %v", nodeNames(members), err)
	}
	group.MinSpeed = 0
	group.MaxDelay = 200
	members, err = group.Members()
	if err != nil || len(members) != 2 || members[0].ID != cebu.ID || members[1].ID != gb.ID {
		t.Fatalf("latency-limit members = %v, %v", nodeNames(members), err)
	}
}

func TestSmartGroupLatencyFreshnessWithoutSpeed(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	old := time.Now().Add(-96 * time.Hour).Format("2006-01-02 15:04:05")
	createSubcriptionTestNode(t, Node{Name: "old-tcp", LinkCountry: "PH", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: old})
	group := SmartGroup{Name: "recent", Countries: "PH", MaxAgeHours: 72}
	members, err := group.Members()
	if err != nil || len(members) != 0 {
		t.Fatalf("stale latency members = %v, %v", nodeNames(members), err)
	}
	group.MaxAgeHours = 0
	members, err = group.Members()
	if err != nil || len(members) != 1 {
		t.Fatalf("unlimited freshness members = %v, %v", nodeNames(members), err)
	}
}

func TestSmartGroupNameCountryFallbackAndExclusionCounts(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	oldRules := countryRuleCache
	t.Cleanup(func() { countryRuleCache = oldRules })
	if err := database.DB.AutoMigrate(&CountryRule{}); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&CountryRule{ID: 910001, CountryCode: "PH", CountryName: "菲律宾", Pattern: "(?i)菲律宾|Philippines|🇵🇭", Enabled: true}).Error; err != nil {
		t.Fatal(err)
	}
	resetCountryRuleCacheForTest()
	if err := InitCountryRuleCache(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	good := createSubcriptionTestNode(t, Node{Name: "Manila 1", LinkName: "🇵🇭 Manila", Group: "source-a", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: now})
	createSubcriptionTestNode(t, Node{Name: "Manila 2", LinkName: "Philippines slow", Group: "source-b", DelayStatus: "timeout"})
	mislabeled := createSubcriptionTestNode(t, Node{Name: "Manila 3", LinkName: "🇵🇭 mislabeled", LinkCountry: "GB", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: now})
	group := SmartGroup{Name: "Philippines", Countries: "PH", MaxAgeHours: 72}
	ids, err := group.CandidateIDs()
	if err != nil || len(ids) != 3 {
		t.Fatalf("country-or-name candidate IDs = %v, %v", ids, err)
	}
	if group.MatchSource(mislabeled) != "countryName" {
		t.Fatalf("stored GB / name PH should match by name, got %q", group.MatchSource(mislabeled))
	}
	members, stats, err := group.MembersWithStats()
	if err != nil || len(members) != 2 || members[0].ID != good.ID || members[1].ID != mislabeled.ID || stats.CandidateCount != 3 || stats.DelayUnusable != 1 {
		t.Fatalf("members = %v, stats = %+v, err = %v", nodeNames(members), stats, err)
	}
	if members[0].LinkCountry != "" {
		t.Fatalf("name inference must not overwrite landing country: %+v", members[0])
	}
	if country, source := SmartGroupCountryForDisplay(members[0]); country != "PH" || source != "name" {
		t.Fatalf("inferred display country = %q from %q", country, source)
	}
	group.SourceGroups = []string{"source-b"}
	ids, err = group.CandidateIDs()
	if err != nil || len(ids) != 1 || ids[0] == good.ID {
		t.Fatalf("filtered inferred candidates = %v, %v", ids, err)
	}
}

func TestSmartGroupCountryCodeTokenMatchesNameWithoutRule(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	matched := createSubcriptionTestNode(t, Node{Name: "PH / Manila", LinkName: "PH / Manila", LinkCountry: "US"})
	createSubcriptionTestNode(t, Node{Name: "ALPHABET node", LinkName: "ALPHABET node", LinkCountry: "US"})
	group := SmartGroup{Name: "PH", Countries: "PH"}
	ids, err := group.CandidateIDs()
	if err != nil || len(ids) != 1 || ids[0] != matched.ID {
		t.Fatalf("country code token candidates = %v, %v", ids, err)
	}
	if got := group.MatchSource(matched); got != "countryName" {
		t.Fatalf("match source = %q, want countryName", got)
	}
}
