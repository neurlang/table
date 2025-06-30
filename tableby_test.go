package table

import (
	"reflect"
	"testing"
)

func TestSanityTableBy(t *testing.T) {
	// helper to assert panic
	expectsPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s: expected panic, but none occurred", name)
			}
		}()
		fn()
	}

	// 1. QueryBy should panic on nil or empty filters
	expectsPanic("QueryBy(nil)", func() { (&Table{}).QueryBy(nil) })
	expectsPanic("QueryBy(empty)", func() { (&Table{}).QueryBy(map[int]string{}) })

	// 2. DeleteBy should panic on nil or empty filters
	expectsPanic("DeleteBy(nil)", func() { (&Table{}).DeleteBy(nil) })
	expectsPanic("DeleteBy(empty)", func() { (&Table{}).DeleteBy(map[int]string{}) })

	// Prepare a table and multi-bucket scenario
	tbl := &Table{}
	// bucket1: simple two-column rows
	tbl.Insert([][]string{
		{"u1", "admin", "active"},
		{"u2", "member", "inactive"},
		{"u3", "admin", "inactive"},
	})
	// bucket2: overlapping and new data, including holes
	tbl.InsertHoles([][]string{
		{"u4", "member", "active"},
		nil, // hole row
		{"u5", "guest", "active"},
	})
	// bucket3: more data
	tbl.Insert([][]string{
		{"u6", "guest", "inactive"},
		{"u7", "admin", "active"},
	})

	// 3. QueryBy no-match returns nil
	if got := tbl.QueryBy(map[int]string{0: "noone"}); got != nil {
		t.Errorf("QueryBy(no-match) = %v; want nil", got)
	}

	// 4. QueryBy single-col matches across buckets
	adminRows := tbl.QueryBy(map[int]string{1: "admin"})
	if len(adminRows) != 3 {
		t.Fatalf("QueryBy(admin) count = %d; want 3", len(adminRows))
	}
	// Verify all returned rows have "admin" in col 1
	for _, r := range adminRows {
		if r[1] != "admin" {
			t.Errorf("QueryBy(admin) returned non-admin row %v", r)
		}
	}

	// 5. QueryBy multi-col filter
	activeAdmin := tbl.QueryBy(map[int]string{1: "admin", 2: "active"})
	want := [][]string{{"u1", "admin", "active"}, {"u7", "admin", "active"}}
	if !reflect.DeepEqual(activeAdmin, want) {
		t.Errorf("QueryBy(admin+active) = %v; want %v", activeAdmin, want)
	}

	// 6. Ensure QueryBy does not mutate underlying data
	_ = tbl.QueryBy(map[int]string{2: "inactive"})
	allBefore := tbl.All()
	tbl.QueryBy(map[int]string{2: "inactive"})
	allAfter := tbl.All()
	if !reflect.DeepEqual(allBefore, allAfter) {
		t.Errorf("QueryBy mutated table: before=%v after=%v", allBefore, allAfter)
	}

	// 7. DeleteBy removes matching rows
	deleted := tbl.DeleteBy(map[int]string{2: "inactive"})
	// We had u2, u3, u6 as inactive → 3 deletions
	if deleted != 3 {
		t.Errorf("DeleteBy(inactive) deleted %d; want 3", deleted)
	}
	// QueryBy should no longer return any inactive
	if got := tbl.QueryBy(map[int]string{2: "inactive"}); got != nil {
		t.Errorf("after DeleteBy, QueryBy(inactive) = %v; want nil", got)
	}
	// Count of any active should remain
	if got := tbl.QueryBy(map[int]string{2: "active"}); len(got) == 0 {
		t.Error("after DeleteBy(inactive), no active rows found; expected some")
	}

	// 8. DeleteBy on overlapping filters
	// Remove everyone with role="admin"
	del2 := tbl.DeleteBy(map[int]string{1: "admin"})
	// Only u1 and u7 were admin+active originally; they should now be deleted
	if del2 != 2 {
		t.Errorf("DeleteBy(admin) deleted %d; want 2", del2)
	}
	remaining := tbl.All()
	// Remaining should be only non-admin, non-inactive: u4 and u5 (and possibly holes already filtered)
	for _, r := range remaining {
		if r[1] == "admin" || r[2] == "inactive" {
			t.Errorf("bad remaining row after DeleteBy: %v", r)
		}
	}

	// 9. DeleteBy complete cleanup
	// Now delete all remaining active members
	del3 := tbl.DeleteBy(map[int]string{2: "active"})
	if del3 == 0 {
		t.Errorf("DeleteBy(active) deleted %d; expected >0", del3)
	}
	if any := tbl.All(); len(any) != 0 {
		t.Errorf("table not empty after final DeleteBy: %v", any)
	}
}
