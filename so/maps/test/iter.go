package maps_test

import (
	"solod.dev/so/maps"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

func TestIter(t *testing.T) {
	m := makeMap(t.Allocator())
	defer m.Free()

	seen := make(map[string]bool, m.Len())
	it := m.Iter()
	for it.Next() {
		k, v := it.Key(), it.Value()
		if m.Get(k) != v {
			t.Error("invalid key-value pair")
		}
		if seen[k] {
			t.Error("duplicate key")
		}
		seen[k] = true
	}
	if len(seen) != m.Len() {
		t.Error("missing keys")
	}
}

func TestIter_Empty(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	it := m.Iter()
	if it.Next() {
		t.Error("expected no elements")
	}
}

func TestIter_AfterDelete(t *testing.T) {
	alloc := t.Allocator()
	const size = 50

	m := maps.New[int, int](alloc, 0)
	defer m.Free()
	want := slices.Make[bool](alloc, size)
	defer slices.Free(alloc, want)

	for i := range size {
		m.Set(i, i*10)
		want[i] = true
	}
	// Deleting an entry shifts the entries after it back one bucket,
	// so the iterator must still see each remaining entry once.
	for i := range size {
		if i%2 == 0 {
			m.Delete(i)
			want[i] = false
		}
	}
	checkIter(t, &m, want)
}

func TestIter_AfterDeleteAll(t *testing.T) {
	alloc := t.Allocator()
	const size = 20

	m := maps.New[int, int](alloc, 0)
	defer m.Free()
	want := slices.Make[bool](alloc, size)
	defer slices.Free(alloc, want)

	for i := range size {
		m.Set(i, i*10)
	}
	for i := range size {
		m.Delete(i)
	}
	checkIter(t, &m, want) // want is all false: the map is empty
}

func TestIter_AfterClear(t *testing.T) {
	alloc := t.Allocator()
	const size = 20

	m := maps.New[int, int](alloc, 0)
	defer m.Free()
	want := slices.Make[bool](alloc, size)
	defer slices.Free(alloc, want)

	for i := range size {
		m.Set(i, i*10)
	}
	m.Clear()
	checkIter(t, &m, want) // want is all false: the map is empty

	// The iterator sees the entries added after Clear.
	m.Set(3, 30)
	want[3] = true
	checkIter(t, &m, want)
}

func TestIter_AfterGrow(t *testing.T) {
	alloc := t.Allocator()
	const size = 200 // forces several resizes

	m := maps.New[int, int](alloc, 0)
	defer m.Free()
	want := slices.Make[bool](alloc, size)
	defer slices.Free(alloc, want)

	for i := range size {
		m.Set(i, i*10)
		want[i] = true
	}
	checkIter(t, &m, want)
}

func TestIter_Twice(t *testing.T) {
	m := makeMap(t.Allocator())
	defer m.Free()

	// Each iterator keeps its own position, so the second one
	// starts from the first entry.
	first, second := 0, 0
	it1 := m.Iter()
	for it1.Next() {
		first++
	}
	it2 := m.Iter()
	for it2.Next() {
		second++
	}
	if first != m.Len() || second != m.Len() {
		t.Errorf("iterators yielded %d and %d keys, want %d each",
			first, second, m.Len())
	}
}

// checkIter compares the map against a model. want[k] reports whether the
// map holds key k, and the value for key k must be k*10.
func checkIter(t *testing.T, m *maps.Map[int, int], want []bool) {
	alloc := t.Allocator()
	found := slices.Make[bool](alloc, len(want))
	defer slices.Free(alloc, found)

	nwant := 0
	for _, ok := range want {
		if ok {
			nwant++
		}
	}

	seen := 0
	it := m.Iter()
	for it.Next() {
		key := it.Key()
		if key < 0 || key >= len(want) {
			t.Fatalf("iterator yielded unknown key %d", key)
			return
		}
		if !want[key] {
			t.Fatalf("iterator yielded absent key %d", key)
			return
		}
		if found[key] {
			t.Fatalf("iterator yielded key %d twice", key)
			return
		}
		if it.Value() != key*10 {
			t.Fatalf("iterator value for %d = %d, want %d",
				key, it.Value(), key*10)
			return
		}
		found[key] = true
		seen++
	}
	if seen != nwant {
		t.Errorf("iterator yielded %d keys, want %d", seen, nwant)
	}
	if m.Len() != nwant {
		t.Errorf("Len = %d, want %d", m.Len(), nwant)
	}
}
