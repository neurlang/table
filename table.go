// Package table implements a memory efficient in-memory multicolumn string table aka multimap
package table

// Table is a memory efficient in-memory multicolumn string table aka multimap
type Table struct {
	b []bucket
}

// Count counts the number of occurences of string val in column col
func (b *Table) Count(col int, val string) (out int) {
	for _, buck := range b.b {
		out += buck.count(col, val)
	}
	return
}

// GetAll loads all the rows which have string val in column col
// Note: This does have a bug where holes are returned as well, this will be fixed in future relase
func (b *Table) GetAll(col int, val string) (data [][]string) {
	return b.GetAllHoles(col, val)
}

// GetAllHoles loads all the rows which have string val in column col
func (b *Table) GetAllHoles(col int, val string) (data [][]string) {
	for _, buck := range b.b {
		fetched := buck.getAll(col, val)
		if len(fetched) > 0 {
			data = append(data, fetched...)
		}
	}
	return
}

// Remove deletes all the rows which have string val in column col
func (b *Table) Remove(col int, val string) {
	for _, buck := range b.b {
		buck.remove(col, val)
	}
}

// Get loads arbitrary single row which does have string val in column col
func (b *Table) Get(col int, val string) (data []string) {
	for _, buck := range b.b {
		data = buck.get(col, val)
		if len(data) > 0 {
			return
		}
	}
	return
}

// InsertHoles inserts rows even if they contain holes (0 column rows) to the table as-is
func (b *Table) InsertHoles(data [][]string) {
	if len(data) > 0 {
		b.b = append(b.b, *newBucket(data))
	}
}

// Insert inserts rows to the table ignoring holes
func (b *Table) Insert(data [][]string) {
	var in [][]string
	for _, row := range data {
		if len(row) > 0 {
			in = append(in, row)
		}
	}
	if len(in) > 0 {
		b.b = append(b.b, *newBucket(in))
	}
}

// Compact compacts the table after multiple inserts
func (b *Table) Compact() {
	var data [][]string
	for _, buck := range b.b {
		data = append(data, buck.all()...)
	}
	b.b = []bucket{*newBucket(data)}
}

// AllHoles returns all data from the table even if there are deletion holes
func (b *Table) AllHoles() (data [][]string) {
	for _, buck := range b.b {
		data = append(data, buck.all()...)
	}
	return
}

// All returns all data from the table skipping the deletion holes
func (b *Table) All() (out [][]string) {
	for _, row := range b.AllHoles() {
		if len(row) > 0 {
			out = append(out, row)
		}
	}
	return
}
