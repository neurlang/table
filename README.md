# 📚 `table`

**A blazing-fast, memory-efficient multicolumn string table (multimap/bimap) for Go.**
Store billions of rows in RAM, filter by multiple `(column → value)` clauses, and perform high-performance queries with optional “holes” for instant deletion performance.

---

## ✨ Features

✅ In-memory multicolumn string storage

✅ Supports duplicate rows, multi-key lookups, and range queries

✅ Explicit *holes* model for cheap deletes

✅ Manual compaction for full reuse of space

✅ Tiny memory footprint with optional quaternary filters

✅ Zero external dependencies except for `quaternary`

---

## 🚀 Quick Example

```go
package main

import (
	"fmt"
	"github.com/neurlang/table"
)

func main() {
	var t table.Table

	// Insert rows
	data := [][]string{
		{"play", "pièce", "obra"},
		{"cup", "tasse", "taza"},
		{"bank", "banque", "banco"},
		{"coin", "pièce", "moneda"},
		{"boat", "bateau", "barco"},
		{"cup", "verre", "copa"},
		{"earth", "terre", "tierra"},
		{"land", "terre", "tierra"},
		{"soap", "savon", "jabón"},
		{"ice", "glace", "hielo"},
		{"book", "livre", "libro"},
		{"room", "pièce", "habitación"},
		{"cup", "coupe", "copa"},
		{"glass", "verre", "copa"},
		{"pie", "gâteau", "tarta"},
	}
	t.Insert(data)
	t.Compact()

	// Query
	fmt.Println("Cups:", t.GetAll(0, "cup"))
	fmt.Println("terre:", t.GetAll(1, "terre"))
	fmt.Println("copa:", t.GetAll(2, "copa"))

	// Insert more
	t.Insert([][]string{
		{"bench", "banc", "banco"},
		{"faucet", "robinet", "llave"},
		{"key", "clé", "llave"},
	})
	t.Compact()

	fmt.Println("banco:", t.GetAll(2, "banco"))
	fmt.Println("llave:", t.GetAll(2, "llave"))

	// Delete by value
	t.Remove(1, "verre")
	fmt.Println("Cups after delete:", t.GetAll(0, "cup"))
}
```

---

## ⚙️ API Overview

| Method                  | Description                                                                           | Direction |
| ----------------------- | ------------------------------------------------------------------------------------- | --------- |
| `Insert(rows)`          | Insert rows. Holes are ignored.                                                       | Write     |
| `InsertHoles(rows)`     | Insert rows as-is, including holes.                                                   | Write     |
| `Remove(col, val)`      | Delete all rows where `col` equals `val`. Leaves holes for speed.                     | Write     |
| `DeleteBy(filters)`     | Delete rows matching every `(col → val)`. Panics if filter is nil or empty.           | Write     |
| `Get(col, val)`         | Get one arbitrary row where `col` equals `val`.                                       | Read      |
| `GetAll(col, val)`      | Get all rows where `col` equals `val`.                                                | Read      |
| `QueryBy(filters)`      | Find all rows matching every `(col → val)`. Skips holes. Panics if filters nil/empty. | Read      |
| `QueryByHoles(filters)` | Same as `QueryBy` but includes holes.                                                 | Read      |
| `All()`                 | Return all rows, skipping holes.                                                      | Read      |
| `AllHoles()`            | Return all rows including holes.                                                      | Read      |
| `Compact()`             | Physically remove holes to reclaim RAM, rebuilds the quaternary indices.              | Write     |
| `Count(col, val)`       | Count number of times `val` appears in `col`.                                         | Read      |

---

## 🧹 Holes & Compaction

* **What’s a “hole”?**
  Deletions just nullify slots for speed. Rows with holes still take space.
* **When to `Compact()`?**
  After bulk inserts or optionally after heavy deletes. Frequent compactions may hurt performance.
* **Do I have to handle holes?**
  Use `QueryBy` and `All()` to skip holes. Use `QueryByHoles` and `AllHoles()` for raw physical view.

---

## ⚡️ Best Practices

* 🗝️ Use a consistent schema: same column count per row.
* ⚠️ Never pass nil or empty filters to `QueryBy` or `DeleteBy` — they will panic!
* 🧹 Run `Compact()` wisely — it’s not automatic.
* 🚀 You can store millions of rows easily, but monitor RAM if you use `InsertHoles` a lot.
* 🐛 Note: `GetAll` may return holes in some versions. Use `QueryBy` if you need strict correctness.

---

## 📏 Limitations

* No schema enforcement: you must keep row length consistent yourself.
* No transactional batch operations.
* No OR filter logic — `QueryBy` is always AND.
* Panics on nil/empty filters — not error-safe by default.
* It’s pure in-memory: no persistence, snapshot, or on-disk mode yet.
* No mutex. Use mutex if threading, based on API call direction.

---

## 📚 Documentation

Full API reference: [pkg.go.dev](https://pkg.go.dev/github.com/neurlang/table)
Issues and improvements: [GitHub Issues](https://github.com/neurlang/table/issues)

---

## 🔑 License

MIT — do anything you want. Attribution appreciated.

---

## 🙌 Contributing

We welcome improvements!
File an issue for bug reports, feature requests, or performance tuning ideas.
Large-scale fuzz tests, schema enforcement, and auto-compaction PRs are especially welcome.

---

## 🏁 Example Output

```shell
[[cup tasse taza] [cup verre copa] [cup coupe copa]]
[[earth terre tierra] [land terre tierra]]
[[cup verre copa] [cup coupe copa] [glass verre copa]]
[[bank banque banco] [bench banc banco]]
[[faucet robinet llave] [key clé llave]]
[[cup tasse taza] [cup coupe copa]]
```

---

## ❤️ Built for performance fanatics, by performance fanatics.
