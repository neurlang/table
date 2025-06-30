package table

import (
	"fmt"
	"sort"

	"github.com/neurlang/quaternary"
)

// getBy returns all raw matches for every (col→val), including nil holes.
// It checks row contents, but is free to return holes (nil rows)
func (b *bucket) getBy(q map[int]string) [][]string {
	if q == nil || len(q) == 0 || len(b.data) == 0 {
		return nil
	}

	type clause struct {
		col, cnt int
		val      string
	}
	cls := make([]clause, 0, len(q))
	for c, v := range q {
		cnt := b.countExisting(c, v)
		if cnt == 0 {
			return nil
		}
		cls = append(cls, clause{col: c, val: v, cnt: cnt})
	}
	sort.Slice(cls, func(i, j int) bool {
		if cls[i].cnt != cls[j].cnt {
			return cls[i].cnt < cls[j].cnt
		}
		return len(cls[i].val) > len(cls[j].val)
	})

	n := len(b.data)
	first := cls[0]
	posList := make([]int, 0, first.cnt)
	// seed positions via index
	for j := 1; j <= first.cnt; j++ {
		key := fmt.Sprintf("%d:%d:%s", j, first.col, first.val)
		bits := 0
		for bit := 0; bit < b.loglen; bit++ {
			if quaternary.Filter(b.filters[bit]).GetString(key) {
				bits |= 1 << bit
			}
		}
		posList = append(posList, bits%n)
	}
	if len(posList) == 0 {
		return nil
	}

	// now post-filter each candidate:
	// - keep any nil (hole)
	// - for non-nil, ensure every clause is satisfied in-row
	var result [][]string
	for _, idx := range posList {
		row := b.data[idx]
		if row == nil {
			// hole: emit as-is
			result = append(result, nil)
			continue
		}
		// verify all clauses
		ok := true
		for _, cl := range cls {
			if cl.col >= len(row) || row[cl.col] != cl.val {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, row)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}


// removeBy deletes all rows matching every (col→val).
// Holes are simply overwritten with nil.
func (b *bucket) removeBy(q map[int]string) {
	if q == nil || len(q) == 0 || len(b.data) == 0 {
		return
	}

	// 1) Gather clauses & bail early
	type clause struct {
		col int
		val string
		cnt int
	}
	cls := make([]clause, 0, len(q))
	for c, v := range q {
		cnt := b.countExisting(c, v)
		if cnt == 0 {
			return
		}
		cls = append(cls, clause{col: c, val: v, cnt: cnt})
	}

	// 2) Sort by selectivity
	sort.Slice(cls, func(i, j int) bool {
		if cls[i].cnt != cls[j].cnt {
			return cls[i].cnt < cls[j].cnt
		}
		return len(cls[i].val) > len(cls[j].val)
	})

	// 3) Seed & intersect exactly as getBy does, but keep indices
	n := len(b.data)
	first := cls[0]
	positions := make([]int, 0, first.cnt)
	// seed from first clause
	for j := 1; j <= first.cnt; j++ {
		key := fmt.Sprintf("%d:%d:%s", j, first.col, first.val)
		var bits int
		for bit := 0; bit < b.loglen; bit++ {
			if quaternary.Filter(b.filters[bit]).GetString(key) {
				bits |= 1 << bit
			}
		}
		positions = append(positions, bits%n)
	}
	if len(positions) == 0 {
		return
	}
	// intersect remaining clauses
	for _, cl := range cls[1:] {
		out := positions[:0]
		for _, idx := range positions {
			// use the same filter logic—no row content checks
			key := fmt.Sprintf("0:%d:%s", cl.col, cl.val)
			var keep bool
			for bit := 0; bit < b.loglen; bit++ {
				if quaternary.Filter(b.filters[bit]).GetString(key) {
					keep = true
					break
				}
			}
			if keep {
				out = append(out, idx)
			}
		}
		positions = out
		if len(positions) == 0 {
			return
		}
	}

	// 4) Nullify exactly those slots
	for _, idx := range positions {
		b.data[idx] = nil
	}
}

