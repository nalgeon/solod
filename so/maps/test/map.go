package maps_test

import (
	"solod.dev/so/maps"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestSetGet(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	m.Set("abc", 11)
	m.Set("def", 22)
	m.Set("xyz", 33)
	if m.Get("abc") != 11 {
		t.Error("want abc = 11")
	}
	key := "abc"
	if m.Get(key) != 11 {
		t.Error("want abc = 11 for key = abc")
	}
	if m.Get("def") != 22 {
		t.Error("want def = 22")
	}
	if m.Get("xyz") != 33 {
		t.Error("want xyz = 33")
	}
	if m.Get("missing") != 0 {
		t.Error("want missing = 0")
	}
	if m.Len() != 3 {
		t.Error("want len = 3")
	}
}

func TestStringValues(t *testing.T) {
	m := maps.New[int32, string](t.Allocator(), 0)
	defer m.Free()

	m.Set(11, "abc")
	m.Set(22, "def")
	m.Set(33, "xyz")
	if m.Get(11) != "abc" {
		t.Error("want 11 = abc")
	}
	if m.Get(22) != "def" {
		t.Error("want 22 = def")
	}
	if m.Get(33) != "xyz" {
		t.Error("want 33 = xyz")
	}
	if m.Get(44) != "" {
		t.Error("want 44 = empty string")
	}
}

func TestHas(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	m.Set("abc", 11)
	m.Set("def", 22)
	if !m.Has("abc") {
		t.Error("want has(abc)")
	}
	if !m.Has("def") {
		t.Error("want has(def)")
	}
	if m.Has("missing") {
		t.Error("want has(missing) == false")
	}
}

func TestDelete(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	m.Set("abc", 11)
	m.Set("def", 22)
	m.Set("xyz", 33)
	m.Delete("def")
	m.Delete("missing") // no-op
	if m.Get("def") != 0 {
		t.Error("want def = 0 after delete")
	}
	if m.Get("abc") != 11 {
		t.Error("want abc = 11 after delete")
	}
	if m.Get("xyz") != 33 {
		t.Error("want xyz = 33 after delete")
	}
	if m.Len() != 2 {
		t.Error("want len = 2 after delete")
	}
}

func TestOverwrite(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	m.Set("key", 100)
	m.Set("key", 200)
	if m.Get("key") != 200 {
		t.Error("want key = 200 after overwrite")
	}
	if m.Len() != 1 {
		t.Error("want len = 1 after overwrite")
	}
}

func TestMissing(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	if m.Get("missing") != 0 {
		t.Error("want missing = 0")
	}
}

func TestGrow(t *testing.T) {
	m := maps.New[int, int](t.Allocator(), 0)
	defer m.Free()

	for i := range 100 {
		m.Set(i, i*10)
	}
	for i := range 100 {
		if m.Get(i) != i*10 {
			t.Error("wrong value after grow")
		}
	}
	if m.Len() != 100 {
		t.Error("want len = 100 after grow")
	}
}

func TestNewSize(t *testing.T) {
	const size = 100
	m := maps.New[int, int](t.Allocator(), size)
	defer m.Free()

	if m.Len() != 0 {
		t.Error("want len = 0 for a new map")
	}
	// The map holds size entries without a resize.
	for i := range size {
		m.Set(i, i*10)
	}
	for i := range size {
		if m.Get(i) != i*10 {
			t.Errorf("Get(%d) = %d, want %d", i, m.Get(i), i*10)
		}
	}
	if m.Len() != size {
		t.Errorf("Len = %d, want %d", m.Len(), size)
	}

	// The map still grows past the initial size.
	for i := size; i < size*3; i++ {
		m.Set(i, i*10)
	}
	for i := range size * 3 {
		if m.Get(i) != i*10 {
			t.Errorf("Get(%d) = %d after grow, want %d", i, m.Get(i), i*10)
		}
	}
	if m.Len() != size*3 {
		t.Errorf("Len = %d after grow, want %d", m.Len(), size*3)
	}
}

func TestNewSize_One(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 1)
	defer m.Free()

	m.Set("abc", 11)
	if m.Get("abc") != 11 {
		t.Error("want abc = 11")
	}
	if m.Len() != 1 {
		t.Error("want len = 1")
	}
}

func TestReturnMap(t *testing.T) {
	m := makeMap(t.Allocator())
	defer m.Free()

	m.Set("mno", 99)
	if m.Get("abc") != 11 {
		t.Error("want abc = 11")
	}
	if m.Get("mno") != 99 {
		t.Error("want mno = 99")
	}
	if m.Len() != 4 {
		t.Error("want len = 4")
	}
}

func TestClear(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	defer m.Free()

	m.Set("abc", 11)
	m.Set("def", 22)
	m.Clear()
	if m.Len() != 0 {
		t.Error("want len = 0 after clear")
	}
	if m.Has("abc") {
		t.Error("want has(abc) == false after clear")
	}
	// reusable after Clear
	m.Set("xyz", 33)
	if m.Get("xyz") != 33 {
		t.Error("want xyz = 33 after reuse")
	}
	if m.Len() != 1 {
		t.Error("want len = 1 after reuse")
	}
}

func TestDoubleFree(t *testing.T) {
	m := maps.New[string, int](t.Allocator(), 0)
	m.Free()
	m.Free() // double Free is a no-op, not a crash
}

func TestDeleteAll(t *testing.T) {
	m := maps.New[int, int](t.Allocator(), 0)
	defer m.Free()

	for i := range 50 {
		m.Set(i, i)
	}
	for i := range 50 {
		m.Delete(i)
	}
	if m.Len() != 0 {
		t.Error("want len = 0 after delete all")
	}
	// re-insert works over tombstones
	m.Set(1, 100)
	if m.Get(1) != 100 {
		t.Error("want 1 = 100 after re-insert")
	}
	if m.Len() != 1 {
		t.Error("want len = 1 after re-insert")
	}
}

func makeMap(a mem.Allocator) maps.Map[string, int] {
	m := maps.New[string, int](a, 0)
	m.Set("abc", 11)
	m.Set("def", 22)
	m.Set("xyz", 33)
	return m
}
