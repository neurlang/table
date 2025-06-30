package table

import (
	"testing"
)

func TestSanityBucket(t *testing.T) {
	table := newBucket([][]string{{"1", "a"}, {"2", "b"}, {"3", "b"}})
	if table.countExisting(1, "b") != 2 {
		panic("fail")
	}
	if table.countExisting(1, "a") != 1 {
		panic("fail")
	}
	if table.countExisting(0, "2") != 1 {
		panic("fail")
	}
	if table.count(1, "x") != 0 {
		panic("fail")
	}
	if table.get(1, "b")[0] != "2" {
		panic("fail")
	}
	if len(table.getAll(1, "b")) != 2 {
		panic("fail")
	}
	table.remove(0, "3")
	if table.get(1, "b")[0] != "2" {
		panic("fail")
	}
	if table.getAll(1, "b")[0][0] != "2" {
		panic("fail")
	}
}
