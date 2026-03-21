// _Maps_ are So's built-in associative data type
// (sometimes called _hashes_ or _dicts_ in other languages).
//
// Built-in maps in So are fixed-size and stack-allocated,
// so only use them when you have a small, fixed number of
// key-value pairs. For anything else, use heap-allocated
// maps from the `so/maps` package (planned).
package main

func main() {
	// To create a map, use the builtin `make`:
	// `make(map[key-type]val-type, size)`.
	m := make(map[string]int, 3)

	// Set key/value pairs using typical `name[key] = val` syntax.
	m["k1"] = 7
	m["k2"] = 13

	// Get a value for a key with `name[key]`.
	v1 := m["k1"]
	println("v1:", v1)

	// If the key doesn't exist, the zero value of the
	// value type is returned.
	v3 := m["k3"]
	println("v3:", v3)

	// The builtin `len` returns the number of key/value
	// pairs when called on a map.
	println("len:", len(m))

	// `delete` and `clear` are not supported.
	// delete(m, "k2")
	// clear(m)

	// The optional second return value when getting a
	// value from a map indicates if the key was present
	// in the map. This can be used to disambiguate
	// between missing keys and keys with zero values
	// like `0` or `""`. Here we didn't need the value
	// itself, so we ignored it with the _blank identifier_
	// `_`.
	_, prs := m["k2"]
	println("prs:", prs)

	// You can also declare and initialize a new map in
	// the same line with this syntax.
	n := map[string]int{"foo": 1, "bar": 2}
	println("foo:", n["foo"], "bar:", n["bar"])
}
