package table

import quaternary "github.com/neurlang/quaternary/v1"

//import "sync"
//import "math/bits"
//import "runtime"

type bucket struct {
	data  [][]string
	index [][][]byte
	//blooms  [][]byte
	loglen int
}

func (b *bucket) filter(j, c int, val string) uint64 {
	if b.loglen == 0 {
		return 0
	}
	if j >= len(b.index) {
		return 0
	}
	if c >= len(b.index[j]) {
		return 0
	}
	if b.index[j][c] == nil {
		return 0
	}
	ret := quaternary.GetNum(b.index[j][c], uint64(b.loglen), val)
	//println(&b.index[[2]int{j, c}][0], j, c, val, "->", ret, "(", b.loglen, ")")
	return ret
}

func (ret *bucket) presentBucket(col int, val string) bool {
	return true

}

/*
	func newBucketfast(rows [][]string) *bucket {
		// Initialize bucket and handle empty input
		ret := &bucket{
			data:  rows,
			index: make(map[[2]int][]byte),
		}
		if len(rows) <= 1 {
			ret.loglen = 0
			return ret
		}

		// Compute number of bits
		loglen := bits.Len(uint(len(rows) - 1))
		ret.loglen = loglen

		// Precompute masks
		bitMasks := make([]uint64, loglen)
		for b := 0; b < loglen; b++ {
			bitMasks[b] = 1 << b
		}

		// Sequential counter pass + collect all partials
		type countKey struct {
			b, x int
			s    string
		}
		counter := make(map[countKey]int, len(rows)*len(rows[0])*loglen)
		type partial struct {
			ikey [2]int
			key  string
			bits uint64
		}
		parts := make([]partial, 0, len(rows)*len(rows[0])*loglen)

		for y, row := range rows {
			rowMask := uint64(y)
			for x, key := range row {
				for b := 0; b < loglen; b++ {
					ck := countKey{b, x, key}
					cnt := counter[ck]
					counter[ck] = cnt + 1

					ik := [2]int{cnt + 1, x}
					var bitsVal uint64
					if rowMask&bitMasks[b] != 0 {
						bitsVal = bitMasks[b]
					}
					parts = append(parts, partial{ik, key, bitsVal})
				}
			}
		}

		// Zero‑count entries
		for k, tot := range counter {
			ik := [2]int{0, k.x}
			boolVal := uint64(tot-1) & bitMasks[k.b]
			parts = append(parts, partial{ik, k.s, boolVal})
		}

		// Shard parts for parallel collection
		nWorkers := runtime.GOMAXPROCS(0)
		shardSize := (len(parts) + nWorkers - 1) / nWorkers
		workerCols := make([]map[[2]int]map[string]uint64, nWorkers)
		for i := range workerCols {
			workerCols[i] = make(map[[2]int]map[string]uint64)
		}

		var wg sync.WaitGroup
		wg.Add(nWorkers)
		for w := 0; w < nWorkers; w++ {
			go func(w int) {
				defer wg.Done()
				start := w * shardSize
				if start >= len(parts) {
					return // nothing to do
				}
				end := start + shardSize
				if end > len(parts) {
					end = len(parts)
				}
				cols := workerCols[w]
				for _, p := range parts[start:end] {
					m := cols[p.ikey]
					if m == nil {
						m = make(map[string]uint64)
						cols[p.ikey] = m
					}
					m[p.key] |= p.bits
				}
			}(w)
		}
		wg.Wait()

		// Merge workerCols into global
		global := make(map[[2]int]map[string]uint64, len(workerCols))
		for _, cols := range workerCols {
			for ik, km := range cols {
				gm := global[ik]
				if gm == nil {
					gm = make(map[string]uint64, len(km))
					global[ik] = gm
				}
				for k, v := range km {
					gm[k] |= v
				}
			}
		}

		// Phase 3: build quaternary filters in parallel
		type task struct {
			ikey [2]int
			val  map[string]uint64
		}
		tasks := make(chan task, len(global))
		results := make(chan struct {
			ik [2]int
			fs []byte
		}, nWorkers)

		go func() {
			for ik, val := range global {
				tasks <- task{ik, val}
			}
			close(tasks)
		}()

		wg.Add(nWorkers)
		for w := 0; w < nWorkers; w++ {
			go func() {
				defer wg.Done()
				for t := range tasks {
					var fs = quaternary.Make(t.val, byte(loglen))
					//for k, v := range t.val {
						//println(&fs[0], t.ikey[0], t.ikey[1], k, v, "(", loglen, ")")
					//}
					results <- struct {
						ik [2]int
						fs []byte
					}{t.ikey, fs}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for r := range results {
			ret.index[r.ik] = r.fs
		}

		return ret
	}
*/
func newBucket(rows [][]string) (ret *bucket) {
	ret = &bucket{
		data:   rows,
		loglen: 0,
	}
	if len(rows) <= 1 {
		return
	}
	for i := 0; 1<<i < len(rows); i++ {
		ret.loglen++
	}
	ret.index = make([][][]byte, ret.loglen+1, ret.loglen+1)
	var counter = make(map[struct {
		b int
		n int
		s string
	}]int)
	var collection = make(map[[2]int]map[string]uint64)
	var maxlen int
	for y := range rows {
		if len(rows[y]) > maxlen {
			maxlen = len(rows[y])
		}
		for x := range rows[y] {
			// do other stuff
			var bval uint64
			key := rows[y][x]
			for b := 0; b < ret.loglen; b++ {
				cnt := counter[struct {
					b int
					n int
					s string
				}{b, x, key}]
				counter[struct {
					b int
					n int
					s string
				}{b, x, key}]++
				intkey := [2]int{cnt + 1, x}
				boolval := uint64(y) & (uint64(1) << b)
				bval |= boolval
				//println(strkey, "=>", boolval)
				if collection[intkey] == nil {
					collection[intkey] = make(map[string]uint64)
				}
				collection[intkey][key] = bval
			}
		}
	}
	for i := range ret.index {
		ret.index[i] = make([][]byte, maxlen, maxlen)
	}
	for k, w := range counter {
		intkey := [2]int{0, k.n}
		strkey := k.s
		boolval := uint64(w-1) & (uint64(1) << k.b)
		if collection[intkey] == nil {
			collection[intkey] = make(map[string]uint64)
		}
		collection[intkey][strkey] |= boolval
		//println(strkey, "=>", boolval)
	}
	for key, val := range collection {
		ret.index[key[0]][key[1]] = quaternary.Make(val, byte(ret.loglen))
		//for k, v := range val {
		//println(&ret.index[key][0], key[0], key[1], k, v, "(", ret.loglen, ")")
		//}
	}
	return
}

func (b *bucket) countExisting(col int, val string) (out int) {
	if !b.presentBucket(col, val) {
		return 0
	}
	out = int(b.filter(0, col, val))
	out++
	return
}
func (b *bucket) count(col int, val string) (out int) {
	if len(b.data) == 0 {
		return 0
	}
	var pos int
	pos = int(b.filter(1, col, val))
	idx := pos % len(b.data)
	if col >= len(b.data[idx]) {
		return 0
	}
	if b.data[idx][col] != val {
		return 0
	}
	out = int(b.filter(0, col, val))
	out++
	return
}

func (b *bucket) all() (data [][]string) {
	return b.data
}

func (b *bucket) getAll(col int, val string) (data [][]string) {
	if len(b.data) == 0 {
		return nil
	}
	cnt := b.countExisting(col, val)
	if cnt == 0 {
		return nil
	}
	for j := 1; j <= cnt; j++ {
		var pos int
		pos = int(b.filter(j, col, val))
		//println(key, pos)
		fetched := b.data[pos%len(b.data)]
		if col < len(fetched) && fetched[col] == val {
			data = append(data, fetched)
		}
	}
	return
}
func (b *bucket) remove(col int, val string) {
	if len(b.data) == 0 {
		return
	}
	cnt := b.countExisting(col, val)
	if cnt == 0 {
		return
	}
	for j := 1; j <= cnt; j++ {
		var pos int
		pos = int(b.filter(j, col, val))
		idx := pos % len(b.data)
		//println(key, pos)
		fetched := b.data[idx]
		if col < len(fetched) && fetched[col] == val {
			b.data[idx] = nil
		}
	}
	return
}
func (b *bucket) get(col int, val string) (data []string) {
	if len(b.data) == 0 {
		return nil
	}
	cnt := b.countExisting(col, val)
	if cnt == 0 {
		return nil
	}
	for j := 1; j <= cnt; j++ {
		var pos int
		pos = int(b.filter(j, col, val))
		//println(key, pos)
		fetched := b.data[pos%len(b.data)]
		if col < len(fetched) && fetched[col] == val {
			data = fetched
			break
		}
	}
	return
}
