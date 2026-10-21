package cache

import (
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestFilterSortedPageMatchesReference(t *testing.T) {
	c := NewMapCache(func(v int) int { return v })
	for _, v := range rand.New(rand.NewSource(42)).Perm(321) {
		c.Set(v, v)
	}
	maxInt := int(^uint(0) >> 1)
	for _, divisor := range []int{1, 3, 1000} {
		predicate := func(v int) bool { return v%divisor == 0 }
		for _, desc := range []bool{false, true} {
			less := func(a, b *int) bool {
				if desc {
					return *a > *b
				}
				return *a < *b
			}
			wantAll := c.Filter(predicate)
			sort.Slice(wantAll, func(i, j int) bool { return less(&wantAll[i], &wantAll[j]) })
			for _, offset := range []int{-1, 0, 1, 30, 150, 300, 321, maxInt} {
				for _, limit := range []int{0, 1, 20, 200, maxInt} {
					got, total := c.FilterSortedPage(predicate, less, offset, limit)
					start := min(max(offset, 0), len(wantAll))
					end := start + min(limit, len(wantAll)-start)
					want := append([]int{}, wantAll[start:end]...)
					if total != len(wantAll) || !reflect.DeepEqual(got, want) {
						t.Fatalf("divisor=%d desc=%v offset=%d limit=%d got=%v want=%v total=%d", divisor, desc, offset, limit, got, want, total)
					}
				}
			}
		}
	}
	got, _ := c.FilterSortedPage(func(int) bool { return true }, func(a, b *int) bool { return *a < *b }, 0, 10)
	got[0] = -999
	if value, _ := c.Get(0); value != 0 {
		t.Fatal("page modified cache")
	}
}

func TestIndexValueCountsTracksMutations(t *testing.T) {
	type item struct {
		ID    int
		Group string
	}
	c := NewMapCache(func(v item) int { return v.ID })
	c.AddIndex("group", func(v item) string { return v.Group })
	c.LoadAll([]item{{1, ""}, {2, "a"}, {3, "a"}})
	assert := func(want map[string]int) {
		t.Helper()
		if got := c.IndexValueCounts("group"); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	assert(map[string]int{"": 1, "a": 2})
	detached := c.IndexValueCounts("group")
	detached["a"] = 999
	c.Set(3, item{3, "b"})
	assert(map[string]int{"": 1, "a": 1, "b": 1})
	c.Delete(2)
	assert(map[string]int{"": 1, "b": 1})
	c.Clear()
	assert(map[string]int{})
	if len(c.IndexValueCounts("absent")) != 0 {
		t.Fatal("unknown index")
	}
}

func TestFilterSortedPageConcurrentUpdates(t *testing.T) {
	c := NewMapCache(func(v int) int { return v })
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 100; i < 300; i++ {
			c.Set(i, i)
			c.Delete(i)
		}
	}()
	for i := 0; i < 100; i++ {
		got, total := c.FilterSortedPage(func(int) bool { return true }, func(a, b *int) bool { return *a < *b }, 0, 10)
		if total < 100 || len(got) != 10 {
			t.Fatalf("inconsistent page %v / %d", got, total)
		}
		for j, v := range got {
			if j != v {
				t.Fatalf("unstable page %v", got)
			}
		}
	}
	wg.Wait()
}
