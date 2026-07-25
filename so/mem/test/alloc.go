package main

import (
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

type Point struct {
	x, y int
}

func TestTryAlloc(t *testing.T) {
	p, err := mem.TryAlloc[Point](mem.System)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	defer mem.Free(mem.System, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("Alloc: unexpected value")
	}
}

func TestTryAllocSlice(t *testing.T) {
	slice, err := mem.TryAllocSlice[int](mem.System, 3, 3)
	if err != nil {
		t.Fatal("AllocSlice: allocation failed")
		return
	}
	defer mem.FreeSlice(mem.System, slice)

	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("AllocSlice: unexpected value")
	}
}

func TestAlloc(t *testing.T) {
	p := mem.Alloc[Point](mem.System)
	defer mem.Free(mem.System, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("New: unexpected value")
	}
}

func TestAllocDefault(t *testing.T) {
	p := mem.Alloc[Point](nil)
	defer mem.Free(nil, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("New: unexpected value")
	}
}

func TestAllocSlice(t *testing.T) {
	slice := mem.AllocSlice[int](mem.System, 3, 3)
	defer mem.FreeSlice(mem.System, slice)

	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("NewSlice: unexpected value")
	}
}

func TestAllocSliceDefault(t *testing.T) {
	slice := mem.AllocSlice[int](nil, 3, 3)
	defer mem.FreeSlice(nil, slice)

	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("NewSlice: unexpected value")
	}
}

func TestTryReallocSlice(t *testing.T) {
	slice, err := mem.TryAllocSlice[int](mem.System, 3, 3)
	if err != nil {
		t.Fatal("ReallocSlice: initial allocation failed")
		return
	}
	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	slice, err = mem.TryReallocSlice(mem.System, slice, 3, 6)
	if err != nil {
		t.Fatal("ReallocSlice: reallocation failed")
		return
	}
	defer mem.FreeSlice(mem.System, slice)

	if len(slice) != 3 || cap(slice) != 6 {
		t.Error("ReallocSlice: unexpected len/cap")
	}
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("ReallocSlice: data not preserved")
	}
}

func TestReallocSlice(t *testing.T) {
	slice := mem.AllocSlice[int](mem.System, 2, 2)
	slice[0] = 44
	slice[1] = 55
	slice = mem.ReallocSlice(mem.System, slice, 4, 8)
	defer mem.FreeSlice(mem.System, slice)

	if len(slice) != 4 || cap(slice) != 8 {
		t.Error("ReallocSlice: unexpected len/cap")
	}
	if slice[0] != 44 || slice[1] != 55 {
		t.Error("ReallocSlice: data not preserved")
	}
	// New elements should be zeroed.
	if slice[2] != 0 || slice[3] != 0 {
		t.Error("ReallocSlice: new elements not zeroed")
	}
}

func TestReallocSlice_Empty(t *testing.T) {
	var empty []int
	slice := mem.ReallocSlice(mem.System, empty, 3, 4)
	defer mem.FreeSlice(mem.System, slice)

	if len(slice) != 3 || cap(slice) != 4 {
		t.Error("ReallocSlice empty: unexpected len/cap")
	}
	if slice[0] != 0 || slice[1] != 0 || slice[2] != 0 {
		t.Error("ReallocSlice empty: not zeroed")
	}
}

func TestFreeNil(t *testing.T) {
	// Freeing a nil pointer or an empty slice is a no-op and must not crash.
	_ = t
	var p *Point
	mem.Free(mem.System, p)
	var empty []int
	mem.FreeSlice(mem.System, empty)
}

func TestFreeString(t *testing.T) {
	_ = t
	b := mem.AllocSlice[byte](mem.System, 3, 3)
	b[0] = 'h'
	b[1] = 'i'
	b[2] = '!'
	s1 := string(b)
	mem.FreeString(mem.System, s1)
	s2 := ""
	mem.FreeString(mem.System, s2)
}

func TestFreeStringStats(t *testing.T) {
	// FreeString must report the string length to the allocator,
	// matching the byte count AllocSlice reported.
	tr := mem.Tracker{Allocator: mem.System}
	b := mem.AllocSlice[byte](&tr, 3, 3)
	b[0] = 'h'
	b[1] = 'i'
	b[2] = '!'
	s := string(b)

	stats := tr.Stats()
	if stats.Alloc != 3 {
		t.Error("Stats.Alloc != 3")
	}

	mem.FreeString(&tr, s)
	stats = tr.Stats()
	if stats.Alloc != 0 {
		t.Error("Stats.Alloc != 0 after free")
	}
	if stats.TotalAlloc != 3 {
		t.Error("Stats.TotalAlloc != 3")
	}
	if stats.Mallocs != 1 || stats.Frees != 1 {
		t.Error("unexpected Mallocs/Frees")
	}
}

func TestFreeStringArena(t *testing.T) {
	// An arena reclaims the last allocation only if the freed size
	// matches the allocated one.
	buf := make([]byte, 8)
	a := mem.NewArena(buf)
	b := mem.AllocSlice[byte](&a, 5, 5)
	s := string(b)
	mem.FreeString(&a, s)

	// The whole buffer is available again.
	_, err := mem.TryAllocSlice[byte](&a, 8, 8)
	if err != nil {
		t.Error("string space not reclaimed")
	}
}
