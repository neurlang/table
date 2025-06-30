package table

// QueryBy finds all rows matching every (col→val).
// Panics if filters is nil or empty.
// Returns nil for no matches.
func (t *Table) QueryBy(filters map[int]string) [][]string {
	if filters == nil || len(filters) == 0 {
		panic("QueryBy: filters must not be nil or empty")
	}
	var result [][]string
	for _, buck := range t.b {
		rows := buck.getBy(filters)
		if rows != nil {
			result = append(result, rows...)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// DeleteBy deletes all rows matching every (col→val).
// Panics if filters is nil or empty.
func (t *Table) DeleteBy(filters map[int]string) {
	if filters == nil || len(filters) == 0 {
		panic("DeleteBy: filters must not be nil or empty")
	}

	for i := range t.b {
		t.b[i].removeBy(filters)
	}
}
