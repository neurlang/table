package table

import (
	"math/rand"
	"testing"
)

func TestTable_FuzzMassive(t *testing.T) {
	const (
		maxRows   = 50_000 // scale to your RAM target
		maxCols   = 10
		maxValLen = 8
		seed      = 42 // make this configurable for reproducibility
	)

	rnd := rand.New(rand.NewSource(seed))

	tbl := &Table{}

	// === Insert Phase ===
	for i := 0; i < maxRows; i++ {
		row := randomRow(rnd, maxCols, maxValLen)
		if rnd.Float64() < 0.05 {
			// 5% chance insert as hole row
			tbl.InsertHoles([][]string{row})
		} else {
			tbl.Insert([][]string{row})
		}
		if i%1_000_000 == 0 && i > 0 {
			t.Logf("Inserted %d rows", i)
		}
	}

	// === Fuzz workload ===
	for i := 0; i < 1000; i++ {
		col := rnd.Intn(maxCols)
		val := randString(rnd, maxValLen)
		count := tbl.Count(col, val)

		// Must match GetAll
		all := tbl.GetAll(col, val)
		if len(all) != count {
			t.Fatalf("Count mismatch: Count()=%d vs GetAll()=%d for col=%d val=%q",
				count, len(all), col, val)
		}

		// Holes must match GetAllHoles >= GetAll
		allHoles := tbl.GetAll(col, val)
		if len(allHoles) < len(all) {
			t.Fatalf("GetAllHoles() should never be smaller than GetAll()")
		}

		// Do a query
		query := tbl.QueryBy(map[int]string{col: val})
		for _, row := range query {
			if row == nil {
				t.Fatalf("QueryBy should skip holes but found nil row")
			}
			if col >= len(row) || row[col] != val {
				t.Fatalf("QueryBy returned invalid row: %+v", row)
			}
		}

		// Delete 10% of values we hit
		if rnd.Float64() < 0.1 {
			tbl.DeleteBy(map[int]string{col: val})
		}
	}

	// === Holes check ===
	all := tbl.All()
	allHoles := tbl.AllHoles()
	if len(allHoles) < len(all) {
		t.Fatalf("AllHoles must never be smaller than All")
	}

	// === Compact and recheck ===
	tbl.Compact()
	allAfter := tbl.All()
	allHolesAfter := tbl.AllHoles()
	if len(allHolesAfter) != len(allAfter) {
		t.Logf("After compact: holes should be gone: all=%d allHoles=%d",
			len(allAfter), len(allHolesAfter))
	}

	t.Logf("Fuzz done: final rows=%d holes=%d",
		len(allAfter), len(allHolesAfter)-len(allAfter))
}
