package table

import "github.com/neurlang/quaternary"
import "fmt"

type bucket struct {
	data    [][]string
	filters [][]byte
	loglen  int
}

func newBucket(rows [][]string) (ret *bucket) {
	var loglen = 0
	for i := 0; 1<<i < len(rows); i++ {
		loglen++
	}
	ret = &bucket{
		data:    rows,
		filters: [][]byte{},
		loglen:  loglen,
	}

	for b := 0; b < loglen; b++ {
		var counter = make(map[string]int)
		var collection = make(map[string]bool)
		var maxcols = 0
		for y := range rows {
			if maxcols < len(rows[y]) {
				maxcols = len(rows[y])
			}
			for x := range rows[y] {
				key := fmt.Sprint(x) + ":" + rows[y][x]
				cnt := counter[key]
				strkey := fmt.Sprint(cnt+1) + ":" + key
				boolval := (y >> b) & 1
				collection[strkey] = boolval == 1
				//println(strkey, "=>", boolval)
				counter[key]++
			}
		}
		for k, v := range counter {
			strkey := "0:" + k
			boolval := ((v - 1) >> b) & 1
			collection[strkey] = boolval == 1
			//println(strkey, "=>", boolval)
		}
		ret.filters = append(ret.filters, []byte(quaternary.MakeString(collection)))
	}
	return
}

func (b *bucket) countExisting(col int, val string) (out int) {
	key := "0:" + fmt.Sprint(col) + ":" + val
	for i := 0; i < b.loglen; i++ {
		if quaternary.Filter(b.filters[i]).GetString(key) {
			out |= 1 << i
		}
	}
	out++
	return
}
func (b *bucket) count(col int, val string) (out int) {
	if len(b.data) == 0 {
		return 0
	}
	key1 := "0:" + fmt.Sprint(col) + ":" + val
	key2 := "1:" + fmt.Sprint(col) + ":" + val
	var pos int
	for i := 0; i < b.loglen; i++ {
		if quaternary.Filter(b.filters[i]).GetString(key2) {
			pos |= 1 << i
		}
	}
	if col >= len(b.data[pos%len(b.data)]) {
		return 0
	}
	if b.data[pos%len(b.data)][col] != val {
		return 0
	}
	for i := 0; i < b.loglen; i++ {
		if quaternary.Filter(b.filters[i]).GetString(key1) {
			out |= 1 << i
		}
	}
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
		key := fmt.Sprint(j) + ":" + fmt.Sprint(col) + ":" + val
		var pos int
		for i := 0; i < b.loglen; i++ {
			if quaternary.Filter(b.filters[i]).GetString(key) {
				pos |= 1 << i
			}
		}
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
		key := fmt.Sprint(j) + ":" + fmt.Sprint(col) + ":" + val
		var pos int
		for i := 0; i < b.loglen; i++ {
			if quaternary.Filter(b.filters[i]).GetString(key) {
				pos |= 1 << i
			}
		}
		//println(key, pos)
		fetched := b.data[pos%len(b.data)]
		if col < len(fetched) && fetched[col] == val {
			b.data[pos%len(b.data)] = nil
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
		key := fmt.Sprint(j) + ":" + fmt.Sprint(col) + ":" + val
		var pos int
		for i := 0; i < b.loglen; i++ {
			if quaternary.Filter(b.filters[i]).GetString(key) {
				pos |= 1 << i
			}
		}
		//println(key, pos)
		fetched := b.data[pos%len(b.data)]
		if col < len(fetched) && fetched[col] == val {
			data = fetched
			break
		}
	}
	return
}
