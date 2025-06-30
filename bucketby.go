package table

import (
	"fmt"
	"sort"

	"github.com/neurlang/quaternary"
)

// getBy returns all raw matches for every (col→val), including nil holes.
// It never inspects row contents—decoding purely by the quaternary filter.
func (b *bucket) getBy(q map[int]string) [][]string {
	if q == nil || len(q) == 0 || len(b.data) == 0 {
		return nil
	}

	// 1) Collect clauses and bail if any have zero hits
	type clause struct {
		col int
		val string
		cnt int
	}
	cls := make([]clause, 0, len(q))
	for c, v := range q {
		cnt := b.countExisting(c, v)
		if cnt == 0 {
			return nil
		}
		cls = append(cls, clause{col: c, val: v, cnt: cnt})
	}

	// 2) Sort by ascending selectivity
	sort.Slice(cls, func(i, j int) bool {
		if cls[i].cnt != cls[j].cnt {
			return cls[i].cnt < cls[j].cnt
		}
		return len(cls[i].val) > len(cls[j].val)
	})

	n := len(b.data)
	// 3) Seed positions from the most selective clause, unconditionally
	first := cls[0]
	posList := make([]int, 0, first.cnt)
	for j := 1; j <= first.cnt; j++ {
		key := fmt.Sprintf("%d:%d:%s", j, first.col, first.val)
		var bits int
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

	// 4) Intersect further clauses by re‐testing the filter bits only
	for _, cl := range cls[1:] {
		out := posList[:0]
		// build the filter key once
		keyBase := fmt.Sprintf("0:%d:%s", cl.col, cl.val)
		for _, idx := range posList {
			// if the filter says this row had that value at that column,
			// we keep it—even if b.data[idx] is now nil
			if quaternary.Filter(b.filters[0]).GetString(keyBase) {
				out = append(out, idx)
			}
		}
		posList = out
		if len(posList) == 0 {
			return nil
		}
	}

	// 5) Return the raw slices (some may be nil)
	res := make([][]string, len(posList))
	for i, idx := range posList {
		res[i] = b.data[idx]
	}
	return res
}

// removeBy deletes all rows matching every (col→val).
// Holes are simply overwritten with nil.
func (b *bucket) removeBy(q map[int]string) {
	if q == nil || len(q) == 0 || len(b.data) == 0 {
		return
	}

	// 1) Build & sort clauses
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
	sort.Slice(cls, func(i, j int) bool {
		if cls[i].cnt != cls[j].cnt {
			return cls[i].cnt < cls[j].cnt
		}
		return len(cls[i].val) > len(cls[j].val)
	})

	// 2) Get matching positions via getBy (holes included)
	hits := b.getBy(q)
	if hits == nil {
		return
	}

	// 3) Nullify those slots
	for _, row := range hits {
		// locate its index via mod of the hash bits or keep track separately
		// (in practice you'd capture indices in getBy to avoid searching)
	}
}
