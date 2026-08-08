package main

import (
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

type Point struct {
	x, y int
}

func TestTryAlloc(t *testing.T) {
	alloc := t.Allocator()
	p, err := mem.TryAlloc[Point](alloc)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	defer mem.Free(alloc, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("Alloc: unexpected value")
	}
}

func TestTryAllocSlice(t *testing.T) {
	alloc := t.Allocator()
	slice, err := mem.TryAllocSlice[int](alloc, 3, 3)
	if err != nil {
		t.Fatal("AllocSlice: allocation failed")
		return
	}
	defer mem.FreeSlice(alloc, slice)

	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("AllocSlice: unexpected value")
	}
}

func TestAlloc(t *testing.T) {
	alloc := t.Allocator()
	p := mem.Alloc[Point](alloc)
	defer mem.Free(alloc, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("New: unexpected value")
	}
}

func TestAllocDefault(t *testing.T) {
	// A nil allocator means the system allocator,
	// so these allocations are not tracked.
	p := mem.Alloc[Point](nil)
	defer mem.Free(nil, p)

	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("New: unexpected value")
	}
}

func TestAllocSlice(t *testing.T) {
	alloc := t.Allocator()
	slice := mem.AllocSlice[int](alloc, 3, 3)
	defer mem.FreeSlice(alloc, slice)

	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("NewSlice: unexpected value")
	}
}

func TestAllocSliceDefault(t *testing.T) {
	// A nil allocator means the system allocator,
	// so these allocations are not tracked.
	slice := mem.AllocSlice[int](nil, 3, 3)
	defer mem.FreeSlice(nil, slice)

	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("NewSlice: unexpected value")
	}
}

func TestAllocSlice_ZeroCap(t *testing.T) {
	// A zero capacity does not allocate.
	alloc := t.Allocator()
	slice, err := mem.TryAllocSlice[int](alloc, 0, 0)
	if err != nil {
		t.Fatal("AllocSlice: allocation failed")
		return
	}
	if len(slice) != 0 || cap(slice) != 0 {
		t.Error("AllocSlice: unexpected len/cap")
	}

	tr := mem.Tracker{Allocator: alloc}
	_ = mem.AllocSlice[int](&tr, 0, 0)
	if tr.Stats().Mallocs != 0 {
		t.Error("AllocSlice: zero capacity allocated memory")
	}

	// Freeing an empty slice is a no-op.
	mem.FreeSlice(alloc, slice)
}

func TestTryReallocSlice(t *testing.T) {
	alloc := t.Allocator()
	slice, err := mem.TryAllocSlice[int](alloc, 3, 3)
	if err != nil {
		t.Fatal("ReallocSlice: initial allocation failed")
		return
	}
	slice[0] = 11
	slice[1] = 22
	slice[2] = 33
	slice, err = mem.TryReallocSlice(alloc, slice, 3, 6)
	if err != nil {
		t.Fatal("ReallocSlice: reallocation failed")
		return
	}
	defer mem.FreeSlice(alloc, slice)

	if len(slice) != 3 || cap(slice) != 6 {
		t.Error("ReallocSlice: unexpected len/cap")
	}
	if slice[0] != 11 || slice[1] != 22 || slice[2] != 33 {
		t.Error("ReallocSlice: data not preserved")
	}
}

func TestReallocSlice(t *testing.T) {
	alloc := t.Allocator()
	slice := mem.AllocSlice[int](alloc, 2, 2)
	slice[0] = 44
	slice[1] = 55
	slice = mem.ReallocSlice(alloc, slice, 4, 8)
	defer mem.FreeSlice(alloc, slice)

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
	alloc := t.Allocator()
	var empty []int
	slice := mem.ReallocSlice(alloc, empty, 3, 4)
	defer mem.FreeSlice(alloc, slice)

	if len(slice) != 3 || cap(slice) != 4 {
		t.Error("ReallocSlice empty: unexpected len/cap")
	}
	if slice[0] != 0 || slice[1] != 0 || slice[2] != 0 {
		t.Error("ReallocSlice empty: not zeroed")
	}
}

func TestReallocSlice_ToZero(t *testing.T) {
	// A zero capacity frees the slice and returns an empty one.
	tr := t.Allocator().(*mem.Tracker)
	slice := mem.AllocSlice[int](tr, 4, 4)
	slice = mem.ReallocSlice(tr, slice, 0, 0)

	if len(slice) != 0 || cap(slice) != 0 {
		t.Error("ReallocSlice to zero: unexpected len/cap")
	}
	stats := tr.Stats()
	if stats.Alloc != 0 {
		t.Error("ReallocSlice to zero: Stats.Alloc != 0")
	}
	if stats.Mallocs != 1 || stats.Frees != 1 {
		t.Error("ReallocSlice to zero: unexpected Mallocs/Frees")
	}
}

// countAllocator delegates to the system allocator and counts the Free calls.
type countAllocator struct {
	frees int
}

func (*countAllocator) Alloc(size int, align int) (any, error) {
	return mem.System.Alloc(size, align)
}

func (*countAllocator) Realloc(ptr any, oldSize int, newSize int, align int) (any, error) {
	return mem.System.Realloc(ptr, oldSize, newSize, align)
}

func (a *countAllocator) Free(ptr any, size int, align int) {
	a.frees++
	mem.System.Free(ptr, size, align)
}

func TestFreeNil(t *testing.T) {
	// Freeing a nil pointer or an empty slice is a no-op.
	// It must not crash and must not reach the allocator.
	var ca countAllocator
	var a mem.Allocator = &ca

	var p *Point
	mem.Free(a, p)
	var empty []int
	mem.FreeSlice(a, empty)
	if ca.frees != 0 {
		t.Error("FreeNil: the allocator was called")
	}

	// A real allocation still reaches the allocator.
	p2 := mem.Alloc[Point](a)
	mem.Free(a, p2)
	if ca.frees != 1 {
		t.Error("FreeNil: unexpected free count")
	}
}

func TestFreeString(t *testing.T) {
	alloc := t.Allocator()
	b := mem.AllocSlice[byte](alloc, 3, 3)
	b[0] = 'h'
	b[1] = 'i'
	b[2] = '!'
	s1 := string(b)
	mem.FreeString(alloc, s1)
	s2 := ""
	mem.FreeString(alloc, s2)
}

func TestFreeStringStats(t *testing.T) {
	// FreeString must report the string length to the allocator,
	// matching the byte count AllocSlice reported.
	tr := t.Allocator().(*mem.Tracker)
	b := mem.AllocSlice[byte](tr, 3, 3)
	b[0] = 'h'
	b[1] = 'i'
	b[2] = '!'
	s := string(b)

	stats := tr.Stats()
	if stats.Alloc != 3 {
		t.Error("Stats.Alloc != 3")
	}

	mem.FreeString(tr, s)
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
