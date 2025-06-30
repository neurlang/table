package table

import (
	"fmt"
	"sort"

	"github.com/neurlang/quaternary"
)

func (b *bucket) getBy(q map[int]string) [][]string {
	if len(b.data) == 0 || len(q) == 0 {
		return nil
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
			return nil
		}
		clauses = append(clauses, clause{col: col, val: val, count: cnt})
	}

	// 2) Sort by ascending count (then by descending val length)
	sort.Slice(clauses, func(i, j int) bool {
		if clauses[i].count != clauses[j].count {
			return clauses[i].count < clauses[j].count
		}
		return len(clauses[i].val) > len(clauses[j].val)
	})

	n := len(b.data)
	first := clauses[0]
	var positions []int

	// 3) Negation optimization: if >50% of rows match, build the complement
	useNeg := first.count*2 > n

	if useNeg {
		// build exclusion set E
		exclude := make(map[int]struct{}, first.count)
		for j := 1; j <= first.count; j++ {
			key := fmt.Sprint(j) + ":" + fmt.Sprint(first.col) + ":" + first.val
			var pos int
			for bit := 0; bit < b.loglen; bit++ {
				if quaternary.Filter(b.filters[bit]).GetString(key) {
					pos |= 1 << bit
				}
			}
			idx := pos % n
			row := b.data[idx]
			if first.col < len(row) && row[first.col] == first.val {
				exclude[idx] = struct{}{}
			}
		}
		// positions = all rows not in exclude
		positions = make([]int, 0, n-len(exclude))
		for i := range b.data {
			if _, found := exclude[i]; !found {
				positions = append(positions, i)
			}
		}
	} else {
		// the usual seed: only those that match
		positions = make([]int, 0, first.count)
		for j := 1; j <= first.count; j++ {
			key := fmt.Sprint(j) + ":" + fmt.Sprint(first.col) + ":" + first.val
			var pos int
			for bit := 0; bit < b.loglen; bit++ {
				if quaternary.Filter(b.filters[bit]).GetString(key) {
					pos |= 1 << bit
				}
			}
			idx := pos % n
			row := b.data[idx]
			if first.col < len(row) && row[first.col] == first.val {
				positions = append(positions, idx)
			}
		}
	}

	// 4) For negation, the remaining clauses become exclusion tests.
	//    For normal, they remain inclusion tests.
	for _, cl := range clauses[1:] {
		var out []int
		if useNeg {
			// remove any that *should* be excluded by cl
			// i.e. if row[col]==val, skip it
			for _, idx := range positions {
				row := b.data[idx]
				if cl.col < len(row) && row[cl.col] == cl.val {
					continue
				}
				out = append(out, idx)
			}
		} else {
			// keep only those that match cl
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

	// 5) Collect
	result := make([][]string, 0, len(positions))
	for _, idx := range positions {
		result = append(result, b.data[idx])
	}
	return result
}
