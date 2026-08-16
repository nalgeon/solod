package mem_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestNoAlloc(t *testing.T) {
	_, err := mem.NoAlloc.Alloc(16, 8)
	if err != mem.ErrNoAlloc {
		t.Error("Alloc: want ErrNoAlloc")
	}

	_, err = mem.NoAlloc.Realloc(nil, 16, 32, 8)
	if err != mem.ErrNoAlloc {
		t.Error("Realloc: want ErrNoAlloc")
	}

	mem.NoAlloc.Free(nil, 0, 0) // must not crash
}

func TestNoAlloc_TryAlloc(t *testing.T) {
	p, err := mem.TryAlloc[Point](mem.NoAlloc)
	if err != mem.ErrNoAlloc {
		t.Error("TryAlloc: want ErrNoAlloc")
	}
	if p != nil {
		t.Error("TryAlloc: want a nil pointer")
	}
}

func TestNoAlloc_TryAllocSlice(t *testing.T) {
	s, err := mem.TryAllocSlice[int](mem.NoAlloc, 3, 3)
	if err != mem.ErrNoAlloc {
		t.Error("TryAllocSlice: want ErrNoAlloc")
	}
	if len(s) != 0 || cap(s) != 0 {
		t.Error("TryAllocSlice: want an empty slice")
	}
}

func TestNoAlloc_TryReallocSlice(t *testing.T) {
	var empty []int
	s, err := mem.TryReallocSlice(mem.NoAlloc, empty, 3, 3)
	if err != mem.ErrNoAlloc {
		t.Error("TryReallocSlice: want ErrNoAlloc")
	}
	if len(s) != 0 || cap(s) != 0 {
		t.Error("TryReallocSlice: want an empty slice")
	}
}

func TestNoAlloc_ZeroCap(t *testing.T) {
	// A zero capacity does not reach the allocator,
	// so it succeeds even with NoAlloc.
	s, err := mem.TryAllocSlice[int](mem.NoAlloc, 0, 0)
	if err != nil {
		t.Error("TryAllocSlice: want no error")
	}
	if len(s) != 0 || cap(s) != 0 {
		t.Error("TryAllocSlice: want an empty slice")
	}

	s, err = mem.TryReallocSlice(mem.NoAlloc, s, 0, 0)
	if err != nil {
		t.Error("TryReallocSlice: want no error")
	}
	if len(s) != 0 || cap(s) != 0 {
		t.Error("TryReallocSlice: want an empty slice")
	}
}
