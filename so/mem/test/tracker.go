package mem_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestTrackerAlloc(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}
	p, err := tr.Alloc(16, 8)
	if err != nil {
		t.Fatal("Alloc: allocation failed")
		return
	}
	if p == nil {
		t.Fatal("Alloc: want non-nil pointer")
		return
	}
	stats := tr.Stats()
	if stats.Alloc != 16 {
		t.Error("Alloc: Stats.Alloc != 16")
	}
	if stats.TotalAlloc != 16 {
		t.Error("Alloc: Stats.TotalAlloc != 16")
	}
	if stats.Mallocs != 1 {
		t.Error("Alloc: Stats.Mallocs != 1")
	}
	if stats.Frees != 0 {
		t.Error("Alloc: Stats.Frees != 0")
	}
	tr.Free(p, 16, 8)
}

func TestTrackerAllocMultiple(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}
	p1, _ := tr.Alloc(16, 8)
	p2, _ := tr.Alloc(32, 8)
	stats := tr.Stats()
	if stats.Alloc != 48 {
		t.Error("AllocMultiple: Stats.Alloc != 48")
	}
	if stats.TotalAlloc != 48 {
		t.Error("AllocMultiple: Stats.TotalAlloc != 48")
	}
	if stats.Mallocs != 2 {
		t.Error("AllocMultiple: Stats.Mallocs != 2")
	}
	tr.Free(p1, 16, 8)
	tr.Free(p2, 32, 8)
}

func TestTrackerAllocError(t *testing.T) {
	buf := make([]byte, 16)
	a := mem.NewArena(buf)
	tr := mem.Tracker{Allocator: &a}
	_, err := tr.Alloc(32, 8)
	if err != mem.ErrOutOfMemory {
		t.Error("AllocError: want ErrOutOfMemory")
	}
	// Stats unchanged after failed alloc.
	stats := tr.Stats()
	if stats.Alloc != 0 {
		t.Error("AllocError: Stats.Alloc != 0")
	}
	if stats.Mallocs != 0 {
		t.Error("AllocError: Stats.Mallocs != 0")
	}
}

func TestTrackerReallocGrow(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}
	p, _ := tr.Alloc(16, 8)
	p2, err := tr.Realloc(p, 16, 32, 8)
	if err != nil {
		t.Fatal("ReallocGrow: reallocation failed")
		return
	}
	if p2 == nil {
		t.Fatal("ReallocGrow: want non-nil pointer")
		return
	}
	stats := tr.Stats()
	if stats.Alloc != 32 {
		t.Error("ReallocGrow: Stats.Alloc != 32")
	}
	if stats.TotalAlloc != 32 {
		t.Error("ReallocGrow: Stats.TotalAlloc != 32")
	}
	if stats.Mallocs != 2 {
		t.Error("ReallocGrow: Stats.Mallocs != 2")
	}
	if stats.Frees != 1 {
		t.Error("ReallocGrow: Stats.Frees != 1")
	}
	tr.Free(p2, 32, 8)
}

func TestTrackerReallocShrink(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}
	p, _ := tr.Alloc(32, 8)
	p2, err := tr.Realloc(p, 32, 16, 8)
	if err != nil {
		t.Fatal("ReallocShrink: reallocation failed")
		return
	}
	stats := tr.Stats()
	if stats.Alloc != 16 {
		t.Error("ReallocShrink: Stats.Alloc != 16")
	}
	if stats.TotalAlloc != 32 {
		t.Error("ReallocShrink: Stats.TotalAlloc != 32") // no increase
	}
	if stats.Mallocs != 2 {
		t.Error("ReallocShrink: Stats.Mallocs != 2")
	}
	if stats.Frees != 1 {
		t.Error("ReallocShrink: Stats.Frees != 1")
	}
	tr.Free(p2, 16, 8)
}

func TestTrackerReallocSameSize(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}
	p, _ := tr.Alloc(16, 8)
	p2, err := tr.Realloc(p, 16, 16, 8)
	if err != nil {
		t.Fatal("ReallocSameSize: reallocation failed")
		return
	}
	stats := tr.Stats()
	if stats.Alloc != 16 {
		t.Error("ReallocSameSize: Stats.Alloc != 16")
	}
	if stats.TotalAlloc != 16 {
		t.Error("ReallocSameSize: Stats.TotalAlloc != 16")
	}
	tr.Free(p2, 16, 8)
}

func TestTrackerReallocError(t *testing.T) {
	buf := make([]byte, 32)
	a := mem.NewArena(buf)
	tr := mem.Tracker{Allocator: &a}
	p, _ := tr.Alloc(16, 8)
	_, err := tr.Realloc(p, 16, 64, 8)
	if err != mem.ErrOutOfMemory {
		t.Error("ReallocError: want ErrOutOfMemory")
	}
	// Stats unchanged after failed realloc.
	stats := tr.Stats()
	if stats.Alloc != 16 {
		t.Error("ReallocError: Stats.Alloc != 16")
	}
	if stats.TotalAlloc != 16 {
		t.Error("ReallocError: Stats.TotalAlloc != 16")
	}
	if stats.Mallocs != 1 {
		t.Error("ReallocError: Stats.Mallocs != 1")
	}
	if stats.Frees != 0 {
		t.Error("ReallocError: Stats.Frees != 0")
	}
}

func TestTrackerFree(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}
	p, _ := tr.Alloc(16, 8)
	tr.Free(p, 16, 8)
	stats := tr.Stats()
	if stats.Alloc != 0 {
		t.Error("Free: Stats.Alloc != 0")
	}
	if stats.TotalAlloc != 16 {
		t.Error("Free: Stats.TotalAlloc != 16")
	}
	if stats.Mallocs != 1 {
		t.Error("Free: Stats.Mallocs != 1")
	}
	if stats.Frees != 1 {
		t.Error("Free: Stats.Frees != 1")
	}
}

func TestTrackerFreeNil(t *testing.T) {
	// Freeing a nil pointer does not change the statistics.
	// Without this check the byte count wraps around.
	tr := mem.Tracker{Allocator: mem.System}
	tr.Free(nil, 16, 8)
	stats := tr.Stats()
	if stats.Alloc != 0 {
		t.Error("FreeNil: Stats.Alloc != 0")
	}
	if stats.Frees != 0 {
		t.Error("FreeNil: Stats.Frees != 0")
	}
}

func TestTrackerLifecycle(t *testing.T) {
	tr := mem.Tracker{Allocator: mem.System}

	// Alloc 16 bytes, then 32 bytes.
	p1, _ := tr.Alloc(16, 8)
	p2, _ := tr.Alloc(32, 8)
	stats := tr.Stats()
	if stats.Alloc != 48 || stats.TotalAlloc != 48 || stats.Mallocs != 2 {
		t.Error("Lifecycle: unexpected stats after alloc")
	}

	// Grow p1: 16 -> 64.
	p1, _ = tr.Realloc(p1, 16, 64, 8)
	stats = tr.Stats()
	if stats.Alloc != 96 || stats.TotalAlloc != 96 || stats.Mallocs != 3 || stats.Frees != 1 {
		t.Error("Lifecycle: unexpected stats after grow")
	}

	// Shrink p2: 32 -> 8.
	p2, _ = tr.Realloc(p2, 32, 8, 8)
	stats = tr.Stats()
	if stats.Alloc != 72 || stats.TotalAlloc != 96 || stats.Mallocs != 4 || stats.Frees != 2 {
		t.Error("Lifecycle: unexpected stats after shrink")
	}

	// Free both.
	tr.Free(p1, 64, 8)
	tr.Free(p2, 8, 8)
	stats = tr.Stats()
	if stats.Alloc != 0 || stats.TotalAlloc != 96 || stats.Mallocs != 4 || stats.Frees != 4 {
		t.Error("Lifecycle: unexpected stats after free")
	}
}
