package main

import (
	"solod.dev/so/maps"
	"solod.dev/so/math"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

// The map hashes a non-string key over its own bytes, so the key size
// selects the branch of the hash function. A bool, an int8 and an int16
// take the branch for 1 to 3 bytes. An int64 and a float64 take the
// branch for 4 to 16 bytes.

func TestKeyBool(t *testing.T) {
	m := maps.New[bool, int](t.Allocator(), 0)
	defer m.Free()

	m.Set(true, 11)
	m.Set(false, 22)
	if m.Get(true) != 11 {
		t.Error("want true = 11")
	}
	if m.Get(false) != 22 {
		t.Error("want false = 22")
	}
	if m.Len() != 2 {
		t.Error("want len = 2")
	}

	m.Delete(true)
	if m.Has(true) {
		t.Error("want has(true) == false after delete")
	}
	if m.Get(false) != 22 {
		t.Error("want false = 22 after delete")
	}
}

func TestKeyInt8(t *testing.T) {
	m := maps.New[int8, int](t.Allocator(), 0)
	defer m.Free()

	keys := []int8{-128, -1, 0, 1, 127}
	for i, key := range keys {
		m.Set(key, i+1)
	}
	if m.Len() != len(keys) {
		t.Errorf("Len = %d, want %d", m.Len(), len(keys))
	}
	for i, key := range keys {
		if got := m.Get(key); got != i+1 {
			t.Errorf("Get(%d) = %d, want %d", int(key), got, i+1)
		}
	}
	if m.Has(42) {
		t.Error("want has(42) == false")
	}
}

func TestKeyInt16(t *testing.T) {
	m := maps.New[int16, int](t.Allocator(), 0)
	defer m.Free()

	keys := []int16{-32768, -256, -1, 0, 1, 256, 32767}
	for i, key := range keys {
		m.Set(key, i+1)
	}
	if m.Len() != len(keys) {
		t.Errorf("Len = %d, want %d", m.Len(), len(keys))
	}
	for i, key := range keys {
		if got := m.Get(key); got != i+1 {
			t.Errorf("Get(%d) = %d, want %d", int(key), got, i+1)
		}
	}
}

func TestKeyUint64(t *testing.T) {
	m := maps.New[uint64, int](t.Allocator(), 0)
	defer m.Free()

	keys := []uint64{0, 1, 0xFFFF, 0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF}
	for i, key := range keys {
		m.Set(key, i+1)
	}
	if m.Len() != len(keys) {
		t.Errorf("Len = %d, want %d", m.Len(), len(keys))
	}
	for i, key := range keys {
		if got := m.Get(key); got != i+1 {
			t.Errorf("Get(%x) = %d, want %d", key, got, i+1)
		}
	}
	m.Delete(0xFFFFFFFFFFFFFFFF)
	if m.Has(0xFFFFFFFFFFFFFFFF) {
		t.Error("want the deleted key to be absent")
	}
	if m.Len() != len(keys)-1 {
		t.Errorf("Len = %d after delete, want %d", m.Len(), len(keys)-1)
	}
}

func TestKeyFloat64(t *testing.T) {
	m := maps.New[float64, int](t.Allocator(), 0)
	defer m.Free()

	keys := []float64{-2.25, -1, 0, 0.5, 1, 1e300}
	for i, key := range keys {
		m.Set(key, i+1)
	}
	if m.Len() != len(keys) {
		t.Errorf("Len = %d, want %d", m.Len(), len(keys))
	}
	for i, key := range keys {
		if got := m.Get(key); got != i+1 {
			t.Errorf("Get(%f) = %d, want %d", key, got, i+1)
		}
	}
}

func TestKeyFloatZero(t *testing.T) {
	// Checks that the two zeros are different keys.
	// The map compares float keys byte by byte, so it does not treat
	// negative zero and positive zero as one key.
	m := maps.New[float64, int](t.Allocator(), 0)
	defer m.Free()

	zero := 0.0
	negZero := math.Copysign(0, -1)

	m.Set(zero, 11)
	m.Set(negZero, 22)
	if m.Len() != 2 {
		t.Errorf("Len = %d, want 2", m.Len())
	}
	if got := m.Get(zero); got != 11 {
		t.Errorf("Get(+0) = %d, want 11", got)
	}
	if got := m.Get(negZero); got != 22 {
		t.Errorf("Get(-0) = %d, want 22", got)
	}
}

func TestKeyFloatNaN(t *testing.T) {
	// Checks that a NaN key stays reachable.
	// The map compares float keys byte by byte, so a NaN key behaves
	// like any other key.
	m := maps.New[float64, int](t.Allocator(), 0)
	defer m.Free()

	nan := math.NaN()
	m.Set(nan, 11)
	if m.Len() != 1 {
		t.Errorf("Len = %d, want 1", m.Len())
	}
	if !m.Has(nan) {
		t.Error("want has(nan)")
	}
	if got := m.Get(nan); got != 11 {
		t.Errorf("Get(nan) = %d, want 11", got)
	}

	m.Delete(nan)
	if m.Has(nan) {
		t.Error("want has(nan) == false after delete")
	}
	if m.Len() != 0 {
		t.Errorf("Len = %d after delete, want 0", m.Len())
	}
}

func TestKeyPointer(t *testing.T) {
	alloc := t.Allocator()
	m := maps.New[*int, int](alloc, 0)
	defer m.Free()

	// The keys are the addresses of the slice elements, not the values.
	vals := slices.Make[int](alloc, 4)
	defer slices.Free(alloc, vals)
	for i := range vals {
		vals[i] = 7 // the same value for every element
	}

	for i := range vals {
		m.Set(&vals[i], i+1)
	}
	if m.Len() != len(vals) {
		t.Errorf("Len = %d, want %d", m.Len(), len(vals))
	}
	for i := range vals {
		if got := m.Get(&vals[i]); got != i+1 {
			t.Errorf("Get(&vals[%d]) = %d, want %d", i, got, i+1)
		}
	}

	var other int
	if m.Has(&other) {
		t.Error("want has(&other) == false")
	}
	if m.Has(nil) {
		t.Error("want has(nil) == false")
	}
}

func TestKeyStringEmpty(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	m.Set("", 11)
	m.Set("a", 22)
	if !m.Has("") {
		t.Error("want has(empty)")
	}
	if m.Get("") != 11 {
		t.Error("want empty = 11")
	}
	if m.Get("a") != 22 {
		t.Error("want a = 22")
	}
	if m.Len() != 2 {
		t.Error("want len = 2")
	}

	m.Delete("")
	if m.Has("") {
		t.Error("want has(empty) == false after delete")
	}
	if m.Get("a") != 22 {
		t.Error("want a = 22 after delete")
	}
	if m.Len() != 1 {
		t.Error("want len = 1 after delete")
	}
}

func TestKeyStringLen(t *testing.T) {
	// Checks the four length branches of the string hash function:
	// zero, 1 to 3 bytes, 4 to 16 bytes, and above 16 bytes.
	// The keys share a prefix and differ only in length, so the map separates
	// them by length alone.
	alloc := t.Allocator()
	sizes := []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 15, 16, 17, 23, 24, 31, 32, 33, 64, 127}

	buf := slices.Make[byte](alloc, 127)
	defer slices.Free(alloc, buf)
	for i := range buf {
		buf[i] = byte('a' + i%26)
	}

	m := maps.New[string, int](alloc, 0)
	defer m.Free()

	for _, size := range sizes {
		m.Set(string(buf[:size]), size+1)
	}
	if m.Len() != len(sizes) {
		t.Errorf("Len = %d, want %d", m.Len(), len(sizes))
	}
	for _, size := range sizes {
		key := string(buf[:size])
		if !m.Has(key) {
			t.Errorf("Has(len %d) = false, want true", size)
		}
		if got := m.Get(key); got != size+1 {
			t.Errorf("Get(len %d) = %d, want %d", size, got, size+1)
		}
	}

	// Delete must find the key at every length.
	for _, size := range sizes {
		m.Delete(string(buf[:size]))
	}
	if m.Len() != 0 {
		t.Errorf("Len = %d after delete all, want 0", m.Len())
	}
}

func TestKeyStringByte(t *testing.T) {
	// Checks that keys of equal length stay distinct when they differ
	// in one byte only. The keys collide often, so the map relies on
	// the key comparison to tell them apart.
	alloc := t.Allocator()
	const size = 24

	buf := slices.Make[byte](alloc, size*size)
	defer slices.Free(alloc, buf)

	m := maps.New[string, int](alloc, 0)
	defer m.Free()

	// Key i is the base key with byte i changed.
	for i := range size {
		key := buf[i*size : i*size+size]
		for j := range key {
			key[j] = 'a'
		}
		key[i] = 'z'
		m.Set(string(key), i+1)
	}
	if m.Len() != size {
		t.Errorf("Len = %d, want %d", m.Len(), size)
	}
	for i := range size {
		key := string(buf[i*size : i*size+size])
		if got := m.Get(key); got != i+1 {
			t.Errorf("Get(%s) = %d, want %d", key, got, i+1)
		}
	}
}
