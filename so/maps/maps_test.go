package maps

import (
	"testing"
)

func FuzzMapInt(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, ops []byte) {
		// The keys are both negative and positive, so the key bytes
		// cover both a set and a clear sign bit.
		fuzzOps(t, ops, func(b byte) int { return int(b) - 128 })
	})
}

func FuzzMapString(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, ops []byte) {
		fuzzOps(t, ops, func(b byte) string { return fuzzKeys[b] })
	})
}

// fuzzOps runs a sequence of operations against both a Map and Go's builtin
// map, and compares the two after every operation. The first input byte sets
// the initial map size. The rest are operation triples: kind, key, value.
// keyOf turns a key byte into a map key.
func fuzzOps[K comparable](t *testing.T, ops []byte, keyOf func(byte) K) {
	if len(ops) == 0 {
		return
	}
	m := New[K, int](nil, int(ops[0]))
	defer m.Free()
	oracle := map[K]int{}

	ops = ops[1:]
	for i := 0; i+2 < len(ops); i += 3 {
		key := keyOf(ops[i+1])
		// The value is never zero, so a stale value differs
		// from the zero value Get returns for a missing key.
		val := int(ops[i+2]) + 1
		switch ops[i] % 16 {
		case 0, 1, 2, 3, 4, 5, 6:
			m.Set(key, val)
			oracle[key] = val
		case 7, 8, 9, 10:
			m.Delete(key)
			delete(oracle, key)
		case 11, 12, 13, 14:
			want, ok := oracle[key]
			if got := m.Get(key); got != want {
				t.Fatalf("op %d: Get(%v) = %d, want %d", i/3, key, got, want)
			}
			if got := m.Has(key); got != ok {
				t.Fatalf("op %d: Has(%v) = %t, want %t", i/3, key, got, ok)
			}
		case 15:
			m.Clear()
			clear(oracle)
		}
		if m.Len() != len(oracle) {
			t.Fatalf("op %d: Len() = %d, want %d", i/3, m.Len(), len(oracle))
		}
	}

	for key, want := range oracle {
		if !m.Has(key) {
			t.Fatalf("Has(%v) = false, want true", key)
		}
		if got := m.Get(key); got != want {
			t.Fatalf("Get(%v) = %d, want %d", key, got, want)
		}
	}

	// The iterator must yield every key of the oracle exactly once.
	seen := map[K]bool{}
	it := m.Iter()
	for it.Next() {
		key, val := it.Key(), it.Value()
		want, ok := oracle[key]
		if !ok {
			t.Fatalf("Iter yielded absent key %v", key)
		}
		if val != want {
			t.Fatalf("Iter value for %v = %d, want %d", key, val, want)
		}
		if seen[key] {
			t.Fatalf("Iter yielded key %v twice", key)
		}
		seen[key] = true
	}
	if len(seen) != len(oracle) {
		t.Fatalf("Iter yielded %d keys, want %d", len(seen), len(oracle))
	}
}

// addSeeds adds the starting inputs for the fuzzers.
func addSeeds(f *testing.F) {
	// Every operation once, on a present key and on an absent one.
	f.Add([]byte{
		0,        // initial size
		0, 1, 10, // Set
		0, 2, 20, // Set
		11, 1, 0, // Get and Has, key present
		7, 1, 0, // Delete
		11, 1, 0, // Get and Has, key absent
		15, 0, 0, // Clear
		0, 3, 30, // Set, reusing the map
		11, 3, 0, // Get and Has, key present
	})

	// Enough Set operations to grow the map twice, then deletes
	// that leave a chain of shifted entries.
	grow := []byte{0} // the smallest initial size
	for i := range 30 {
		grow = append(grow, 0, byte(i), byte(i))
	}
	for i := range 10 {
		grow = append(grow, 7, byte(i*3), 0)
	}
	f.Add(grow)
}

// fuzzKeys holds one string key per key byte. The key lengths cover the
// four length branches of the string hash function.
var fuzzKeys = makeFuzzKeys()

func makeFuzzKeys() []string {
	sizes := []int{0, 1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 24, 31, 32, 33, 64}
	keys := make([]string, 256)
	for i := range keys {
		buf := make([]byte, sizes[i%len(sizes)])
		for j := range buf {
			buf[j] = byte('a' + (i+j)%26)
		}
		keys[i] = string(buf)
	}
	return keys
}
