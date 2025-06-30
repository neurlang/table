package table

import (
	"reflect"
	"testing"
)

func TestHoleApis(t *testing.T) {
	// 1) Start with exactly two rows
	tbl := &Table{}
	tbl.Insert([][]string{
		{"row1", "X"},
		{"row2", "Y"},
	})

	// Sanity check both rows are present
	if got := tbl.All(); !reflect.DeepEqual(got, [][]string{{"row1", "X"}, {"row2", "Y"}}) {
		t.Fatalf("before delete, All = %v; want both rows", got)
	}

	// 2) Delete the first row by matching on column 0 == "row1"
	tbl.DeleteBy(map[int]string{0: "row1"})

	// 3) Now inspect AllHoles (should show a nil at index 0, and the second row)
	wantHoles := [][]string{nil, {"row2", "Y"}}
	if got := tbl.AllHoles(); !reflect.DeepEqual(got, wantHoles) {
		t.Errorf("AllHoles after delete = %v; want %v", got, wantHoles)
	}

	// 4) Inspect All (should skip the hole and return only the existing row)
	wantAll := [][]string{{"row2", "Y"}}
	if got := tbl.All(); !reflect.DeepEqual(got, wantAll) {
		t.Errorf("All after delete = %v; want %v", got, wantAll)
	}

	// 5) QueryByHoles for col0="row2" should still see row2
	hh := tbl.QueryByHoles(map[int]string{0: "row2"})
	if !reflect.DeepEqual(hh, [][]string{{"row2", "Y"}}) {
		t.Errorf("QueryByHoles(row2) = %v; want [[\"row2\",\"Y\"]]", hh)
	}

	// 6) QueryBy for col0="row2" should also see row2
	q := tbl.QueryBy(map[int]string{0: "row2"})
	if !reflect.DeepEqual(q, [][]string{{"row2", "Y"}}) {
		t.Errorf("QueryBy(row2) = %v; want [[\"row2\",\"Y\"]]", q)
	}

	// 7) QueryByHoles for the deleted key="row1" will still return a hole entry
	hhDead := tbl.QueryByHoles(map[int]string{0: "row1"})
	if len(hhDead) != 1 || hhDead[0] != nil {
		t.Errorf("QueryByHoles(row1) = %v; want [nil]", hhDead)
	}

	// 8) QueryBy for the deleted key="row1" should return nil (skipping the hole)
	qDead := tbl.QueryBy(map[int]string{0: "row1"})
	if qDead != nil {
		t.Errorf("QueryBy(row1) = %v; want nil", qDead)
	}
}
