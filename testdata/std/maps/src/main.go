package main

import "solod.dev/so/maps"

func main() {
	{
		// SetGet: insert 3 entries, verify all values
		m := maps.New[string, int](nil, 0)
		m.Set("abc", 11)
		m.Set("def", 22)
		m.Set("xyz", 33)
		if m.Get("abc") != 11 {
			panic("want abc = 11")
		}
		key := "abc"
		if m.Get(key) != 11 {
			panic("want abc = 11 for key = abc")
		}
		if m.Get("def") != 22 {
			panic("want def = 22")
		}
		if m.Get("xyz") != 33 {
			panic("want xyz = 33")
		}
		if m.Get("missing") != 0 {
			panic("want missing = 0")
		}
		if m.Len() != 3 {
			panic("want len = 3")
		}
		m.Free()
	}
	{
		// String values.
		m := maps.New[int32, string](nil, 0)
		m.Set(11, "abc")
		m.Set(22, "def")
		m.Set(33, "xyz")
		if m.Get(11) != "abc" {
			panic("want 11 = abc")
		}
		if m.Get(22) != "def" {
			panic("want 22 = def")
		}
		if m.Get(33) != "xyz" {
			panic("want 33 = xyz")
		}
		if m.Get(44) != "" {
			panic("want 44 = empty string")
		}
		m.Free()
	}
	{
		// Has.
		m := maps.New[string, int](nil, 0)
		m.Set("abc", 11)
		m.Set("def", 22)
		if !m.Has("abc") {
			panic("want has(abc)")
		}
		if !m.Has("def") {
			panic("want has(def)")
		}
		if m.Has("missing") {
			panic("want has(missing) == false")
		}
		m.Free()
	}
	{
		// Delete: insert 3 entries, delete one, verify
		m := maps.New[string, int](nil, 0)
		m.Set("abc", 11)
		m.Set("def", 22)
		m.Set("xyz", 33)
		m.Delete("def")
		m.Delete("missing") // no-op
		if m.Get("def") != 0 {
			panic("want def = 0 after delete")
		}
		if m.Get("abc") != 11 {
			panic("want abc = 11 after delete")
		}
		if m.Get("xyz") != 33 {
			panic("want xyz = 33 after delete")
		}
		if m.Len() != 2 {
			panic("want len = 2 after delete")
		}
		m.Free()
	}
	{
		// Overwrite: set same key twice, verify latest value wins
		m := maps.New[string, int](nil, 0)
		m.Set("key", 100)
		m.Set("key", 200)
		if m.Get("key") != 200 {
			panic("want key = 200 after overwrite")
		}
		if m.Len() != 1 {
			panic("want len = 1 after overwrite")
		}
		m.Free()
	}
	{
		// Missing: get non-existent key returns zero value
		m := maps.New[string, int](nil, 0)
		if m.Get("missing") != 0 {
			panic("want missing = 0")
		}
		m.Free()
	}
	{
		// Grow: insert 100 int-keyed entries, verify all retrievable
		m := maps.New[int, int](nil, 0)
		for i := range 100 {
			m.Set(i, i*10)
		}
		for i := range 100 {
			if m.Get(i) != i*10 {
				panic("wrong value after grow")
			}
		}
		if m.Len() != 100 {
			panic("want len = 100 after grow")
		}
		m.Free()
	}
}
