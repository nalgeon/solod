package main

import (
	"solod.dev/so/maps"
	"solod.dev/so/math/rand"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

// The random tests run a sequence of random operations against a reference
// model and compare the results after every operation. The model is two
// parallel slices: vals[i] is the value for keys[i], and present[i] reports
// whether the map holds keys[i].
const (
	numKeys = 64   // distinct keys the operations choose from
	numOps  = 5000 // operations per test

	// The keys are far fewer than the operations, so the operations hit the
	// same key many times and force collisions, growth, and deletion.
	maxKeyLen = 40 // longest generated string key
)

const (
	seed1 uint64 = 0x9E3779B97F4A7C15
	seed2 uint64 = 0xBF58476D1CE4E5B9
)

func TestRandomInt(t *testing.T) {
	alloc := t.Allocator()

	keys := slices.Make[int](alloc, numKeys)
	defer slices.Free(alloc, keys)
	for i := range keys {
		keys[i] = i
	}

	vals := slices.Make[int](alloc, numKeys)
	defer slices.Free(alloc, vals)
	present := slices.Make[bool](alloc, numKeys)
	defer slices.Free(alloc, present)
	size := 0

	m := maps.New[int, int](alloc, 0)
	defer m.Free()

	src := rand.NewPCG(seed1, seed2)
	r := rand.New(&src)

	for op := range numOps {
		i := r.IntN(numKeys)
		key := keys[i]
		switch r.IntN(10) {
		case 0, 1, 2, 3:
			// Set (40%). The value is unique and non-zero, so a stale value
			// differs from a fresh one and from a missing key.
			val := op + 1
			m.Set(key, val)
			vals[i] = val
			if !present[i] {
				present[i] = true
				size++
			}
		case 4, 5:
			// Delete (20%).
			m.Delete(key)
			if present[i] {
				present[i] = false
				size--
			}
		case 6, 7:
			// Get (20%).
			want := 0
			if present[i] {
				want = vals[i]
			}
			if got := m.Get(key); got != want {
				t.Fatalf("op %d: Get(%d) = %d, want %d", op, key, got, want)
				return
			}
		case 8:
			// Has (10%).
			if m.Has(key) != present[i] {
				t.Fatalf("op %d: Has(%d) is wrong", op, key)
				return
			}
		case 9:
			// Len (10%).
			if m.Len() != size {
				t.Fatalf("op %d: Len = %d, want %d", op, m.Len(), size)
				return
			}
		}
	}

	// Every key must agree with the model.
	if m.Len() != size {
		t.Fatalf("Len = %d, want %d", m.Len(), size)
		return
	}
	for i, key := range keys {
		if m.Has(key) != present[i] {
			t.Fatalf("Has(%d) is wrong", key)
			return
		}
		want := 0
		if present[i] {
			want = vals[i]
		}
		if got := m.Get(key); got != want {
			t.Fatalf("Get(%d) = %d, want %d", key, got, want)
			return
		}
	}

	// The iterator must yield every present key exactly once.
	found := slices.Make[bool](alloc, numKeys)
	defer slices.Free(alloc, found)
	seen := 0
	it := m.Iter()
	for it.Next() {
		i := slices.Index(keys, it.Key())
		if i < 0 {
			t.Fatalf("Iter yielded unknown key %d", it.Key())
			return
		}
		if !present[i] {
			t.Fatalf("Iter yielded deleted key %d", keys[i])
			return
		}
		if found[i] {
			t.Fatalf("Iter yielded key %d twice", keys[i])
			return
		}
		if it.Value() != vals[i] {
			t.Fatalf("Iter value for %d = %d, want %d", keys[i], it.Value(), vals[i])
			return
		}
		found[i] = true
		seen++
	}
	if seen != size {
		t.Fatalf("Iter yielded %d keys, want %d", seen, size)
		return
	}
}

func TestRandomString(t *testing.T) {
	alloc := t.Allocator()

	buf := slices.Make[byte](alloc, numKeys*maxKeyLen)
	defer slices.Free(alloc, buf)
	keys := slices.Make[string](alloc, numKeys)
	defer slices.Free(alloc, keys)
	buildKeys(keys, buf)

	vals := slices.Make[int](alloc, numKeys)
	defer slices.Free(alloc, vals)
	present := slices.Make[bool](alloc, numKeys)
	defer slices.Free(alloc, present)
	size := 0

	m := maps.New[string, int](alloc, 0)
	defer m.Free()

	src := rand.NewPCG(seed1, seed2)
	r := rand.New(&src)

	for op := range numOps {
		i := r.IntN(numKeys)
		key := keys[i]
		switch r.IntN(10) {
		case 0, 1, 2, 3:
			// Set (40%). The value is unique and non-zero, so a stale value
			// differs from a fresh one and from a missing key.
			val := op + 1
			m.Set(key, val)
			vals[i] = val
			if !present[i] {
				present[i] = true
				size++
			}
		case 4, 5:
			// Delete (20%).
			m.Delete(key)
			if present[i] {
				present[i] = false
				size--
			}
		case 6, 7:
			// Get (20%).
			want := 0
			if present[i] {
				want = vals[i]
			}
			if got := m.Get(key); got != want {
				t.Fatalf("op %d: Get(%s) = %d, want %d", op, key, got, want)
				return
			}
		case 8:
			// Has (10%).
			if m.Has(key) != present[i] {
				t.Fatalf("op %d: Has(%s) is wrong", op, key)
				return
			}
		case 9:
			// Len (10%).
			if m.Len() != size {
				t.Fatalf("op %d: Len = %d, want %d", op, m.Len(), size)
				return
			}
		}
	}

	// Every key must agree with the model.
	if m.Len() != size {
		t.Fatalf("Len = %d, want %d", m.Len(), size)
		return
	}
	for i, key := range keys {
		if m.Has(key) != present[i] {
			t.Fatalf("Has(%s) is wrong", key)
			return
		}
		want := 0
		if present[i] {
			want = vals[i]
		}
		if got := m.Get(key); got != want {
			t.Fatalf("Get(%s) = %d, want %d", key, got, want)
			return
		}
	}

	// The iterator must yield every present key exactly once.
	found := slices.Make[bool](alloc, numKeys)
	defer slices.Free(alloc, found)
	seen := 0
	it := m.Iter()
	for it.Next() {
		i := slices.Index(keys, it.Key())
		if i < 0 {
			t.Fatalf("Iter yielded unknown key %s", it.Key())
			return
		}
		if !present[i] {
			t.Fatalf("Iter yielded deleted key %s", keys[i])
			return
		}
		if found[i] {
			t.Fatalf("Iter yielded key %s twice", keys[i])
			return
		}
		if it.Value() != vals[i] {
			t.Fatalf("Iter value for %s = %d, want %d", keys[i], it.Value(), vals[i])
			return
		}
		found[i] = true
		seen++
	}
	if seen != size {
		t.Fatalf("Iter yielded %d keys, want %d", seen, size)
		return
	}
}

// buildKeys fills keys with distinct strings, using buf as the backing store.
// The key lengths run from 3 to maxKeyLen bytes, which covers the short,
// medium, and long branches of the map hash function.
func buildKeys(keys []string, buf []byte) {
	pos := 0
	for i := range keys {
		size := 3 + i%(maxKeyLen-3+1)
		key := buf[pos : pos+size]
		pos += size
		// The three-digit index prefix keeps the keys distinct.
		key[0] = byte('0' + i/100%10)
		key[1] = byte('0' + i/10%10)
		key[2] = byte('0' + i%10)
		for j := 3; j < size; j++ {
			key[j] = byte('a' + j%26)
		}
		keys[i] = string(key)
	}
}
