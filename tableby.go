package table

// QueryByHoles finds all rows matching every (col→val), including any holes.
// Panics if filters is nil or empty.
// Returns nil for no matches.
func (t *Table) QueryByHoles(filters map[int]string) [][]string {
    if filters == nil || len(filters) == 0 {
        panic("QueryByHoles: filters must not be nil or empty")
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

// QueryBy finds all rows matching every (col→val), skipping any holes.
// Panics if filters is nil or empty.
// Returns nil for no matches.
func (t *Table) QueryBy(filters map[int]string) [][]string {
    raw := t.QueryByHoles(filters)
    if raw == nil {
        return nil
    }
    var filtered [][]string
    for _, row := range raw {
        if row != nil && len(row) > 0 {
            filtered = append(filtered, row)
        }
    }
    if len(filtered) == 0 {
        return nil
    }
    return filtered
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
