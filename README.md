# Table
Package table implements a memory efficient in-memory multicolumn string table aka multimap/bimap in golang

# Example

```go
package main

import "github.com/neurlang/table"
import "fmt"

var data = [][]string{
	{"play",	"pièce",	"obra"},
	{"cup",		"tasse",	"taza"},
	{"bank",	"banque",	"banco"},
	{"coin",	"pièce",	"moneda"},
	{"boat",	"bateau",	"barco"},
	{"cup",		"verre",	"copa"},
	{"earth",	"terre",	"tierra"},
	{"land",	"terre",	"tierra"},
	{"soap",	"savon",	"jabón"},
	{"ice",		"glace",	"hielo"},
	{"book",	"livre",	"libro"},
	{"room",	"pièce",	"habitación"},
	{"cup",		"coupe",	"copa"},
	{"glass",	"verre",	"copa"},
	{"pie",		"gâteau",	"tarta"},
}

func main() {
	var table table.Table
	
	// Insert initial data
	table.Insert(data)
	table.Compact()

	// Do lookups
	fmt.Println(table.GetAll(0, "cup"))
	fmt.Println(table.GetAll(1, "terre"))
	fmt.Println(table.GetAll(2, "copa"))

	// Insert something
	table.Insert([][]string{{"bench", "banc", "banco"}, {"faucet", "robinet", "llave"}, {"key", "clé", "llave"}})
	table.Compact()

	// Do lookups on new data as well
	fmt.Println(table.GetAll(2, "banco"))
	fmt.Println(table.GetAll(2, "llave"))

	// Deletion
	table.Remove(1, "verre")
	// No need to compact after deletions, only after insertions

	// Do lookups
	fmt.Println(table.GetAll(0, "cup"))
}
```

# Documentation

- [pkg.go.dev](https://pkg.go.dev/github.com/neurlang/table)
