package main

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
