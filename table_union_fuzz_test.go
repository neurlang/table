package table

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Utility to sort slice-of-rows for comparison.
func sortRows(rows [][]string) {
	sort.Slice(rows, func(i, j int) bool {
		return strings.Join(rows[i], "|") < strings.Join(rows[j], "|")
	})
}

// Merges two sorted slices of rows (with possible duplicates).
func unionRows(a, b [][]string) [][]string {
	out := append(append([][]string(nil), a...), b...)
	sortRows(out)
	return out
}

func TestFuzzUnion(t *testing.T) {
	const (
		maxCols     = 5
		totalIters  = 5000
		lookupIters = 50
		maxValLen   = 8
		seed        = 42
	)

	r := rand.New(rand.NewSource(seed))
	A := &Table{}
	B := &Table{}
	var C *Table

	for iter := 0; iter < totalIters; iter++ {
		// Randomly choose insert vs delete (80% inserts)
		if r.Float64() < 0.8 {
			row := randomRow(r, maxCols, maxValLen)
			if r.Float64() < 0.5 {
				A.Insert([][]string{row})
			} else {
				B.Insert([][]string{row})
			}
		} else {
			// Random delete: pick non-empty table A or B
			var T *Table
			if r.Float64() < 0.5 {
				T = A
			} else {
				T = B
			}
			// Pick a random existing row and delete by one of its columns
			all := T.All()
			if len(all) > 0 {
				row := all[r.Intn(len(all))]
				col := r.Intn(len(row))
				val := row[col]
				T.DeleteBy(map[int]string{col: val})
			}
		}

		// Rebuild C = A ∪ B
		C = &Table{}
		C.Insert(A.All())
		C.Insert(B.All())
		C.Compact()

		// Now test random single-column lookups
		for li := 0; li < lookupIters; li++ {
			col := r.Intn(maxCols)
			// choose a test value: either random or drawn from C
			var val string
			if r.Float64() < 0.5 {
				val = randString(r, 3)
			} else {
				h := C.All()
				if len(h) > 0 {
					row := h[r.Intn(len(h))]
					if col < len(row) {
						val = row[col]
					}
				}
			}

			aRows := A.GetAll(col, val)
			bRows := B.GetAll(col, val)
			exp := unionRows(aRows, bRows)

			cRows := C.GetAll(col, val)
			sortRows(cRows)

			if !equalRows(cRows, exp) {
				t.Fatalf("Invariant broken at iter %d lookup %d:\n"+
					"col=%d val=%q\nA rows = %v\nB rows = %v\n"+
					"C rows = %v\nexpected union = %v",
					iter, li, col, val, aRows, bRows, cRows, exp)
			}
		}

		// Occasional compaction to reclaim holes
		if iter%1000 == 0 {
			A.Compact()
			B.Compact()
			C.Compact()
		}
	}
}

// equalRows assumes both are sorted by sortRows().
func equalRows(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func randomRow(rnd *rand.Rand, maxCols, maxValLen int) []string {
	n := rnd.Intn(maxCols) + 1
	row := make([]string, n)
	for i := range row {
		if rnd.Float64() < 0.05 {
			continue // randomly keep some cols empty
		}
		row[i] = randString(rnd, maxValLen)
	}
	return row
}

func randString(rnd *rand.Rand, length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	s := make([]byte, length)
	for i := range s {
		s[i] = chars[rnd.Intn(len(chars))]
	}
	return string(s)
}
