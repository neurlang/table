package table

import (
	"fmt"
	"sort"

	"github.com/neurlang/quaternary"
)

// getBy returns all rows matching every (col→val) clause in q.
// Returns nil if q is nil/empty, no data, or no matches.
func (b *bucket) getBy(q map[int]string) [][]string {
	// Favor nil for empty filters or empty data
	if q == nil || len(q) == 0 || len(b.data) == 0 {
		return nil
	}

	type clause struct {
		col   int
		val   string
		count int
	}

	// 1) Collect counts & bail early if any clause has zero matches
	clauses := make([]clause, 0, len(q))
	for col, val := range q {
		cnt := b.countExisting(col, val)
		if cnt == 0 {
			return nil
		}
		clauses = append(clauses, clause{col: col, val: val, count: cnt})
	}

	// 2) Sort by ascending count, tie-breaker by descending val length
	sort.Slice(clauses, func(i, j int) bool {
		if clauses[i].count != clauses[j].count {
			return clauses[i].count < clauses[j].count
		}
		return len(clauses[i].val) > len(clauses[j].val)
	})

	n := len(b.data)
	first := clauses[0]
	var positions []int

	// 3) Seed candidate indices from the most selective clause
	useNeg := first.count*2 > n
	if useNeg {
		// exclusion set
		exclude := make(map[int]struct{}, first.count)
		for j := 1; j <= first.count; j++ {
			key := fmt.Sprintf("%d:%d:%s", j, first.col, first.val)
			var pos int
			for bit := 0; bit < b.loglen; bit++ {
				if quaternary.Filter(b.filters[bit]).GetString(key) {
					pos |= 1 << bit
				}
			}
			idx := pos % n
			if row := b.data[idx]; first.col < len(row) && row[first.col] == first.val {
				exclude[idx] = struct{}{}
			}
		}
		positions = make([]int, 0, n-len(exclude))
		for i := range b.data {
			if _, found := exclude[i]; !found {
				positions = append(positions, i)
			}
		}
	} else {
		positions = make([]int, 0, first.count)
		for j := 1; j <= first.count; j++ {
			key := fmt.Sprintf("%d:%d:%s", j, first.col, first.val)
			var pos int
			for bit := 0; bit < b.loglen; bit++ {
				if quaternary.Filter(b.filters[bit]).GetString(key) {
					pos |= 1 << bit
				}
			}
			idx := pos % n
			if row := b.data[idx]; first.col < len(row) && row[first.col] == first.val {
				positions = append(positions, idx)
			}
		}
	}

	// Early exit if seed yields no candidates
	if len(positions) == 0 {
		return nil
	}

	// 4) Filter by remaining clauses
	for _, cl := range clauses[1:] {
		out := positions[:0]
		if useNeg {
			// exclude rows matching cl
			for _, idx := range positions {
				row := b.data[idx]
				if cl.col < len(row) && row[cl.col] == cl.val {
					continue
				}
				out = append(out, idx)
			}
		} else {
			// include only rows matching cl
			for _, idx := range positions {
				row := b.data[idx]
				if cl.col < len(row) && row[cl.col] == cl.val {
					out = append(out, idx)
				}
			}
		}
		positions = out
		if len(positions) == 0 {
			return nil
		}
	}

	// 5) Collect final rows; if none, return nil
	result := make([][]string, len(positions))
	for i, idx := range positions {
		result[i] = b.data[idx]
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// removeBy deletes all rows matching every (col→val) clause in q.
// Returns immediately if q is nil/empty or no data.
func (b *bucket) removeBy(q map[int]string) {
	if q == nil || len(q) == 0 || len(b.data) == 0 {
		return
	}
	type clause struct {
		col   int
		val   string
		count int
	}

	// 1) Collect counts & bail early
	clauses := make([]clause, 0, len(q))
	for col, val := range q {
		cnt := b.countExisting(col, val)
		if cnt == 0 {
			return
		}
		clauses = append(clauses, clause{col: col, val: val, count: cnt})
	}

	// 2) Sort by ascending count, tie-breaker by descending val length
	sort.Slice(clauses, func(i, j int) bool {
		if clauses[i].count != clauses[j].count {
			return clauses[i].count < clauses[j].count
		}
		return len(clauses[i].val) > len(clauses[j].val)
	})

	n := len(b.data)
	first := clauses[0]

	// 3) Seed candidates from the most selective clause
	positions := make([]int, 0, first.count)
	for j := 1; j <= first.count; j++ {
		key := fmt.Sprintf("%d:%d:%s", j, first.col, first.val)
		var pos int
		for bit := 0; bit < b.loglen; bit++ {
			if quaternary.Filter(b.filters[bit]).GetString(key) {
				pos |= 1 << bit
			}
		}
		idx := pos % n
		if row := b.data[idx]; first.col < len(row) && row[first.col] == first.val {
			positions = append(positions, idx)
		}
	}
	if len(positions) == 0 {
		return
	}

	// 4) Filter remaining clauses
	for _, cl := range clauses[1:] {
		out := positions[:0]
		for _, idx := range positions {
			row := b.data[idx]
			if cl.col < len(row) && row[cl.col] == cl.val {
				out = append(out, idx)
			}
		}
		positions = out
		if len(positions) == 0 {
			return
		}
	}

	// 5) Nullify matching rows
	for _, idx := range positions {
		b.data[idx] = nil
	}
}
