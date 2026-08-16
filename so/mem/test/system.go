package mem_test

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestSystemAlloc(t *testing.T) {
	// The system allocator zeroes new memory.
	p, err := mem.System.Alloc(32, 8)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	if p == nil {
		t.Fatal("Alloc: want a non-nil pointer")
		return
	}
	defer mem.System.Free(p, 32, 8)

	b := c.Bytes(c.PtrAs[byte](p), 32)
	for i := range b {
		if b[i] != 0 {
			t.Error("Alloc: new memory is not zeroed")
			return
		}
	}
}

func TestSystemReallocGrow(t *testing.T) {
	p, err := mem.System.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	b := c.Bytes(c.PtrAs[byte](p), 8)
	for i := range b {
		b[i] = byte(i + 1)
	}

	newPtr, err := mem.System.Realloc(p, 8, 24, 1)
	if err != nil {
		t.Fatal("Realloc: reallocation failed")
		return
	}
	defer mem.System.Free(newPtr, 24, 1)

	nb := c.Bytes(c.PtrAs[byte](newPtr), 24)
	for i := range 8 {
		if nb[i] != byte(i+1) {
			t.Error("Realloc: data not preserved")
			return
		}
	}
	// Realloc zeroes the memory beyond the old size.
	for i := 8; i < 24; i++ {
		if nb[i] != 0 {
			t.Error("Realloc: new memory is not zeroed")
			return
		}
	}
}

func TestSystemReallocShrink(t *testing.T) {
	p, err := mem.System.Alloc(16, 1)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	b := c.Bytes(c.PtrAs[byte](p), 16)
	for i := range b {
		b[i] = byte(i + 1)
	}

	newPtr, err := mem.System.Realloc(p, 16, 4, 1)
	if err != nil {
		t.Fatal("Realloc: reallocation failed")
		return
	}
	defer mem.System.Free(newPtr, 4, 1)

	nb := c.Bytes(c.PtrAs[byte](newPtr), 4)
	for i := range nb {
		if nb[i] != byte(i+1) {
			t.Error("Realloc: data not preserved")
			return
		}
	}
}

func TestSystemFree(t *testing.T) {
	// Freeing a nil pointer is a no-op.
	mem.System.Free(nil, 0, 0)

	// Free ignores the size and the alignment.
	p, err := mem.System.Alloc(16, 8)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	mem.System.Free(p, 16, 8)
}
