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

	// Setup: multiple buckets with holes
	tbl := &Table{}
	tbl.Insert([][]string{
		{"u1", "admin", "active"},
		{"u2", "member", "inactive"},
		{"u3", "admin", "inactive"},
	})
	tbl.InsertHoles([][]string{
		{"u4", "member", "active"},
		nil,
		{"u5", "guest", "active"},
	})
	tbl.Insert([][]string{
		{"u6", "guest", "inactive"},
		{"u7", "admin", "active"},
	})

	// 3. QueryBy no-match returns nil
	if got := tbl.QueryBy(map[int]string{0: "noone"}); got != nil {
		t.Errorf("QueryBy(no-match) = %v; want nil", got)
	}

	// 4. QueryBy single-column
	if got := tbl.Count(1, "admin"); got != 3 {
		t.Errorf("Count(admin) = %d; want 3", got)
	}
	adminRows := tbl.QueryBy(map[int]string{1: "admin"})
	var nonNilRows int
	for _, row := range adminRows {
		if row != nil && len(row) > 0 {
			nonNilRows++
		}
	}

	if nonNilRows != 3 {
		t.Fatalf("QueryBy(admin) non-nil count = %d; want 3", nonNilRows)
	}
	// 5. QueryBy multi-column
	expected := [][]string{
		{"u1", "admin", "active"},
		{"u7", "admin", "active"},
	}
	if got := tbl.QueryBy(map[int]string{1: "admin", 2: "active"}); !reflect.DeepEqual(got, expected) {
		t.Errorf("QueryBy(admin+active) = %v; want %v", got, expected)
	}

	// 6. QueryBy is read-only
	before := tbl.All()
	_ = tbl.QueryBy(map[int]string{2: "inactive"})
	after := tbl.All()
	if !reflect.DeepEqual(before, after) {
		t.Errorf("QueryBy mutated table: before=%v after=%v", before, after)
	}

	// 7. DeleteBy removes inactive rows
	tbl.DeleteBy(map[int]string{2: "inactive"})
	if tbl.QueryBy(map[int]string{2: "inactive"}) != nil {
		t.Error("after DeleteBy(inactive), found inactive rows; expected none")
	}
	// ensure active rows remain
	if cnt := tbl.Count(2, "active"); cnt == 0 {
		t.Error("after DeleteBy(inactive), no active rows found; expected some")
	}

	// 8. DeleteBy overlapping filters (remove all admins)
	tbl.DeleteBy(map[int]string{1: "admin"})
	if tbl.QueryBy(map[int]string{1: "admin"}) != nil {
		t.Error("after DeleteBy(admin), found admin rows; expected none")
	}

	// 9. Idempotence: calling DeleteBy again on same filter is a no-op
	tbl.DeleteBy(map[int]string{1: "admin"}) // should not panic or alter anything
	if tbl.QueryBy(map[int]string{1: "admin"}) != nil {
		t.Error("DeleteBy(admin) second call should be no-op, still no admin rows")
	}

	// 10. Final cleanup: remove everything active
	tbl.DeleteBy(map[int]string{2: "active"})
	if any := tbl.All(); len(any) != 0 {
		t.Errorf("table not empty after final DeleteBy(active): %v", any)
	}
}
