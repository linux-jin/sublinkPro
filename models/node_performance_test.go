package models

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func loadNodePerformanceFixture(tb testing.TB, count int) {
	tb.Helper()
	previous := nodeCache
	tb.Cleanup(func() { nodeCache = previous })
	resetNodeCacheForTest()
	summary := `{"providers":[{"provider":"gemini","status":"available","region":"US","detail":"` + strings.Repeat("synthetic detail ", 16) + `"},{"provider":"netflix","status":"restricted","region":"JP"}]}`
	nodes := make([]Node, count)
	for i := range nodes {
		id := i + 1
		nodes[i] = Node{ID: id, Name: fmt.Sprintf("remark-%06d", id), LinkName: fmt.Sprintf("node-%06d", id), Link: fmt.Sprintf("trojan://synthetic@example.invalid:%d#node", 1000+id), Group: fmt.Sprintf("group-%d", id%20), Source: fmt.Sprintf("source-%d", id%15), Protocol: "trojan", LinkCountry: []string{"US", "TR", "JP", "DE"}[id%4], DelayTime: id, Speed: float64(id), DelayStatus: "success", FraudScore: -1, UnlockSummary: summary, Tags: "a,b"}
	}
	nodeCache.LoadAll(nodes)
}

func BenchmarkNodeLargeList(b *testing.B) {
	for _, count := range []int{10000, 50000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			loadNodePerformanceFixture(b, count)
			for _, tc := range []struct {
				name   string
				filter NodeFilter
			}{
				{"first_page", NodeFilter{}},
				{"name_sort", NodeFilter{SortBy: "name"}},
				{"name_search", NodeFilter{Search: "node-"}},
				{"narrow_unlock", NodeFilter{Group: "group-1", UnlockRules: []UnlockFilterRule{{Provider: "gemini", Status: "available"}}}},
			} {
				b.Run(tc.name, func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						if _, _, err := (&Node{}).ListWithFiltersPaginated(tc.filter, 1, 50); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
			b.Run("filter_stats", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					GetNodeCountryStats()
					GetNodeGroupStats()
					GetNodeSourceStats()
				}
			})
		})
	}
}

func TestNodePaginationMatchesFullList(t *testing.T) {
	loadNodePerformanceFixture(t, 137)
	for _, sortBy := range []string{"", "unknown", "name", "delay", "speed", "protocol", "group", "source", "country"} {
		for _, order := range []string{"asc", "desc"} {
			filter := NodeFilter{SortBy: sortBy, SortOrder: order}
			// Existing equal-key order is undefined; use unique values for baseline comparison.
			for _, n := range nodeCache.GetAll() {
				n.Group = fmt.Sprintf("g-%04d", n.ID)
				nodeCache.Set(n.ID, n)
			}
			all, err := (&Node{}).ListWithFilters(filter)
			if err != nil {
				t.Fatal(err)
			}
			for _, page := range []int{1, 2, 5, 6, 100} {
				got, total, err := (&Node{}).ListWithFiltersPaginated(filter, page, 30)
				if err != nil {
					t.Fatal(err)
				}
				start := min((page-1)*30, len(all))
				end := min(start+30, len(all))
				want := all[start:end]
				if total != int64(len(all)) || len(got) != len(want) {
					t.Fatalf("%s/%s page %d total=%d len=%d want=%d", sortBy, order, page, total, len(got), len(want))
				}
				for i := range got {
					if got[i].ID != want[i].ID {
						t.Fatalf("%s/%s page %d IDs differ", sortBy, order, page)
					}
				}
			}
		}
	}
}

func TestNodeFilterCheapAndUnlockSemantics(t *testing.T) {
	loadNodePerformanceFixture(t, 3)
	n, _ := nodeCache.Get(1)
	n.Group = ""
	n.Source = ""
	n.UnlockSummary = "invalid"
	n.Tags = " a, b, a "
	nodeCache.Set(1, n)
	n, _ = nodeCache.Get(2)
	n.Group = "Mixed"
	n.Source = "manual"
	n.Protocol = "VMESS"
	n.FraudScore = 10
	n.QualityStatus = QualityStatusSuccess
	n.IsResidential = true
	nodeCache.Set(2, n)
	for _, tc := range []struct {
		name   string
		filter NodeFilter
		ids    []int
	}{
		{"name with invalid summary", NodeFilter{Search: "node-000001"}, []int{1}},
		{"summary search", NodeFilter{Search: "netflix"}, []int{2, 3}},
		{"ungrouped", NodeFilter{Group: "未分组"}, []int{1}},
		{"case insensitive", NodeFilter{Group: "mixed", Protocol: "vmess"}, []int{2}},
		{"manual", NodeFilter{Source: "手动添加"}, []int{1, 2}},
		{"threshold", NodeFilter{MinSpeed: 1, MaxDelay: 2}, []int{2}},
		{"tags and excluded", NodeFilter{Tags: []string{"a"}, ExcludeIDs: []int{2, 3}}, []int{1}},
		{"quality", NodeFilter{MaxFraudScore: 20, ResidentialType: "residential"}, []int{2}},
		{"all unlock", NodeFilter{UnlockRuleMode: "and", UnlockRules: []UnlockFilterRule{{Provider: "gemini", Status: "available"}, {Provider: "netflix", Status: "restricted"}}}, []int{2, 3}},
		{"any unlock", NodeFilter{UnlockRules: []UnlockFilterRule{{Provider: "gemini", Status: "error"}, {Provider: "netflix", Status: "restricted"}}}, []int{2, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := (&Node{}).GetFilteredNodeIDs(tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.ids) {
				t.Fatalf("got %v want %v", got, tc.ids)
			}
		})
	}
}

func TestNodePageTieOrderAndInvalidValues(t *testing.T) {
	loadNodePerformanceFixture(t, 120)
	for _, n := range nodeCache.GetAll() {
		n.Name = "same"
		n.LinkName = "same"
		n.DelayTime = 10
		n.Speed = 1
		n.Group = "same"
		nodeCache.Set(n.ID, n)
	}
	for _, field := range []string{"name", "delay", "speed", "group"} {
		for _, order := range []string{"asc", "desc"} {
			got, total, err := (&Node{}).ListWithFiltersPaginated(NodeFilter{SortBy: field, SortOrder: order}, 2, 20)
			if err != nil || total != 120 {
				t.Fatalf("%v total=%d", err, total)
			}
			for i, n := range got {
				if n.ID != 21+i {
					t.Fatalf("unstable ties %s/%s: %d", field, order, n.ID)
				}
			}
		}
	}
	n, _ := nodeCache.Get(1)
	n.DelayTime = 0
	n.Speed = 0
	nodeCache.Set(1, n)
	for _, field := range []string{"delay", "speed"} {
		for _, order := range []string{"asc", "desc"} {
			got, _, err := (&Node{}).ListWithFiltersPaginated(NodeFilter{SortBy: field, SortOrder: order}, 6, 20)
			if err != nil || got[len(got)-1].ID != 1 {
				t.Fatalf("invalid values must be last: %s/%s", field, order)
			}
		}
	}
	got, total, err := (&Node{}).ListWithFiltersPaginated(NodeFilter{}, int(^uint(0)>>1), 100)
	if err != nil || total != 120 || len(got) != 0 {
		t.Fatalf("overflow page got %d total=%d err=%v", len(got), total, err)
	}
}

func TestNodeIndexStatsPreserveSpecialBuckets(t *testing.T) {
	loadNodePerformanceFixture(t, 3)
	n, _ := nodeCache.Get(1)
	n.Group = ""
	n.Source = ""
	n.LinkCountry = ""
	nodeCache.Set(1, n)
	n, _ = nodeCache.Get(2)
	n.Group = "未分组"
	n.Source = "manual"
	n.LinkCountry = "未知"
	nodeCache.Set(2, n)
	n, _ = nodeCache.Get(3)
	n.Group = "未分组"
	n.Source = "手动添加"
	n.LinkCountry = "未知"
	nodeCache.Set(3, n)
	if !reflect.DeepEqual(GetNodeGroupStats(), map[string]int{"未分组": 3}) || !reflect.DeepEqual(GetNodeSourceStats(), map[string]int{"手动添加": 3}) || !reflect.DeepEqual(GetNodeCountryStats(), map[string]int{"未知": 3}) {
		t.Fatal("special bucket counts lost")
	}
	nodeCache.Delete(1)
	if GetNodeCountryStats()["未知"] != 2 {
		t.Fatal("deleted node still counted")
	}
}

func TestSmartGroupEvaluationUsesConsistentCandidatesAndCutoff(t *testing.T) {
	loadNodePerformanceFixture(t, 100)
	group := SmartGroup{Countries: "US", MaxDelay: 50}
	candidates, members, stats, cutoff, err := group.EvaluateMembers()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 25 || len(members) != 12 || stats.CandidateCount != 25 || stats.DelayOverLimit != 13 {
		t.Fatalf("candidates=%d members=%d stats=%+v", len(candidates), len(members), stats)
	}
	included := map[int]bool{}
	for i, n := range members {
		included[n.ID] = true
		if i > 0 && members[i-1].ID >= n.ID {
			t.Fatal("members not ordered")
		}
	}
	for i, n := range candidates {
		if i > 0 && candidates[i-1].ID >= n.ID {
			t.Fatal("candidates not ordered")
		}
		if included[n.ID] != (group.CandidateExclusionReason(n, cutoff) == "") {
			t.Fatalf("member/reason mismatch for %d", n.ID)
		}
	}
}

func BenchmarkSmartGroupCandidateEvaluation(b *testing.B) {
	loadNodePerformanceFixture(b, 10000)
	group := SmartGroup{Countries: "US"}
	b.Run("legacy_double_scan", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, _, err := group.MembersWithStats(); err != nil {
				b.Fatal(err)
			}
			if _, err := group.Candidates(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("single_snapshot", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, _, _, _, err := group.EvaluateMembers(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
