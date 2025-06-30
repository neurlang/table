package table

import (
	"fmt"
	"testing"
)

// sampleData generates a synthetic bucket data set with `rows` rows and `cols` columns.
func sampleData(rows, cols int) [][]string {
	data := make([][]string, rows)
	for i := 0; i < rows; i++ {
		row := make([]string, cols)
		for j := 0; j < cols; j++ {
			// simple pattern: "C{col}-R{row}"
			row[j] = fmt.Sprintf("C%d-R%d", j, i)
		}
		data[i] = row
	}
	return data
}

func TestGetBy_SingleClause(t *testing.T) {
	rows := [][]string{
		{"apple", "red"},
		{"banana", "yellow"},
		{"apple", "green"},
		{"cherry", "red"},
	}
	b := newBucket(rows)

	result := b.GetBy(map[int]string{0: "apple"})
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
	for _, r := range result {
		if r[0] != "apple" {
			t.Errorf("unexpected row %v", r)
		}
	}
}

func TestGetBy_MultiClause(t *testing.T) {
	rows := [][]string{
		{"user1", "admin", "active"},
		{"user2", "member", "inactive"},
		{"user3", "admin", "inactive"},
		{"user4", "member", "active"},
	}
	b := newBucket(rows)

	result := b.GetBy(map[int]string{1: "admin", 2: "inactive"})
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	expected := []string{"user3", "admin", "inactive"}
	if fmt.Sprint(result[0]) != fmt.Sprint(expected) {
		t.Errorf("got %v, want %v", result[0], expected)
	}
}

func TestGetBy_NoMatch(t *testing.T) {
	rows := [][]string{{"a", "b"}, {"c", "d"}}
	b := newBucket(rows)

	result := b.GetBy(map[int]string{0: "x"})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestRemoveBy_Basic(t *testing.T) {
	rows := [][]string{
		{"1", "keep"},
		{"2", "remove"},
		{"3", "remove"},
		{"4", "keep"},
	}
	b := newBucket(rows)

	b.RemoveBy(map[int]string{1: "remove"})

	// After removal, only 2 rows should remain non-nil
	count := 0
	for _, r := range b.All() {
		if r != nil {
			count++ // non-nil rows
			if r[1] == "remove" {
				t.Errorf("found removed row %v", r)
			}
		}
	}
	if count != 2 {
		t.Errorf("expected 2 remaining rows, got %d", count)
	}
}

func TestRemoveBy_MultiClause(t *testing.T) {
	rows := [][]string{
		{"a", "x", "1"},
		{"a", "y", "2"},
		{"b", "x", "2"},
		{"b", "y", "1"},
	}
	b := newBucket(rows)

	b.RemoveBy(map[int]string{0: "a", 2: "2"})
	remaining := b.All()
	for _, r := range remaining {
		if r != nil && r[0] == "a" && r[2] == "2" {
			t.Errorf("row %v should have been removed", r)
		}
	}
}

// BenchmarkGetBy compares GetBy against a manual filter + GetAll for varying sizes
func BenchmarkGetBy(b *testing.B) {
	const rows = 10000
	const cols = 5
	data := sampleData(rows, cols)
	bucket := newBucket(data)

	b.Run("GetBy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bucket.GetBy(map[int]string{1: "C1-R5000", 3: "C3-R5000"})
		}
	})

	b.Run("FilterAfterGetAll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			all := bucket.All()
			var filtered [][]string
			for _, r := range all {
				if r[1] == "C1-R5000" && r[3] == "C3-R5000" {
					filtered = append(filtered, r)
				}
			}
			_ = filtered
		}
	})
}

// BenchmarkRemoveBy measures the cost of a multi-clause RemoveBy
func BenchmarkRemoveBy(b *testing.B) {
	const rows = 5000
	const cols = 4
	data := sampleData(rows, cols)
	bucket := newBucket(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// re-seed for each iteration
		bucket = newBucket(data)
		bucket.RemoveBy(map[int]string{2: "C2-R2500", 0: "C0-R2500"})
	}
}
