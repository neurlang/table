package table

// QueryBy finds all rows matching every (col→val).
// Panics if filters is nil or empty.
// Returns nil for no matches.
func (t *Table) QueryBy(filters map[int]string) [][]string {
	if filters == nil || len(filters) == 0 {
		panic("QueryBy: filters must not be nil or empty")
	}
	var result [][]string
	// delegate to each bucket
	for _, buck := range t.buckets {
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
// Returns the total number of rows deleted.
func (t *Table) DeleteBy(filters map[int]string) int {
	if filters == nil || len(filters) == 0 {
		panic("DeleteBy: filters must not be nil or empty")
	}
	deleted := 0
	for i := range t.buckets {
		// count before
		before := len(t.buckets[i].getBy(filters))
		t.buckets[i].removeBy(filters)
		// count after
		afterRows := t.buckets[i].getBy(filters)
		if afterRows == nil {
			deleted += before
		} else {
			deleted += before - len(afterRows)
		}
	}
	return deleted
}
