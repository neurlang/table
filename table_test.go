package table

import (
	"testing"
)

func TestSanityTable(t *testing.T) {
	table := (&Table{})
	table.Insert([][]string{{"1","a"},{"2","b"},{"3", "b"}})
	if (table.Count(1, "b") != 2) {panic("fail");}
	if (table.Count(1, "a") != 1) {panic("fail");}
	if (table.Count(0, "2") != 1) {panic("fail");}
	if (table.Count(1, "x") != 0) {panic("fail");}
	if (table.Get(1, "b")[0] != "2") {panic("fail");}
	if (len(table.GetAll(1, "b")) != 2) {panic("fail");}
	table.Remove(0, "3")
	if (table.Get(1, "b")[0] != "2") {panic("fail");}
	if (table.GetAll(1, "b")[0][0] != "2") {panic("fail");}
	table.Insert([][]string{{"4","a"},{"5","b"},{"6", "c"}})
	if (table.Count(1, "b") != 3) {panic("fail");}
	if (table.Count(1, "a") != 2) {panic("fail");}
	if (table.Count(0, "2") != 1) {panic("fail");}
	if (table.Count(1, "x") != 0) {panic("fail");}
	if (table.Get(1, "b")[0] != "2") {panic("fail");}
	if (len(table.GetAll(1, "b")) != 2) {panic("fail");}
	table.Remove(0, "5")
	if (table.Get(1, "b")[0] != "2") {panic("fail");}
	if (table.GetAll(1, "b")[0][0] != "2") {panic("fail");}
	if(len(table.All()) != 4) {panic("fail");}
}
