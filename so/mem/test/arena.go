package mem_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestArena(t *testing.T) {
	buf := make([]byte, 1024)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	// Allocate a Point.
	p, err := mem.TryAlloc[Point](a)
	if err != nil {
		t.Fatal("initial allocation failed")
		return
	}
	p.x = 11
	p.y = 22
	if p.x != 11 || p.y != 22 {
		t.Error("unexpected p.x or p.y")
	}

	// Free last allocation reclaims space.
	mem.Free(a, p)

	// Allocate again: should reuse the same space.
	p2, err := mem.TryAlloc[Point](a)
	if err != nil {
		t.Fatal("allocation after free failed")
		return
	}
	// Memory should be zeroed.
	if p2.x != 0 || p2.y != 0 {
		t.Error("memory not zeroed after free")
	}
	p2.x = 33
	p2.y = 44

	// Free non-last allocation is a no-op.
	p3, err := mem.TryAlloc[Point](a)
	if err != nil {
		t.Fatal("allocation for p3 failed")
		return
	}
	p3.x = 55
	p3.y = 66
	mem.Free(a, p2) // not last, no-op

	// Reset and reallocate.
	arena.Reset()
	p4, err := mem.TryAlloc[Point](a)
	if err != nil {
		t.Fatal("allocation after reset failed")
		return
	}
	if p4.x != 0 || p4.y != 0 {
		t.Error("memory not zeroed after reset")
	}
	p4.x = 77
	p4.y = 88
}

func TestArena_Placement(t *testing.T) {
	// The arena places each block at a known offset in the buffer,
	// so the returned pointer tells where the block starts.
	buf := make([]byte, 256)
	arena := mem.NewArena(buf)

	p1, err := arena.Alloc(1, 1)
	if err != nil {
		t.Fatal("Alloc(1, 1) failed")
		return
	}
	if p1.(*byte) != &buf[0] {
		t.Error("Alloc(1, 1): want the block at offset 0")
	}

	// The offset is 1 now, so an 8-byte alignment pads it to 8.
	p2, err := arena.Alloc(8, 8)
	if err != nil {
		t.Fatal("Alloc(8, 8) failed")
		return
	}
	if p2.(*byte) != &buf[8] {
		t.Error("Alloc(8, 8): want the block at offset 8")
	}

	// The offset is 16 now, which a 4-byte alignment does not pad.
	p3, err := arena.Alloc(4, 4)
	if err != nil {
		t.Fatal("Alloc(4, 4) failed")
		return
	}
	if p3.(*byte) != &buf[16] {
		t.Error("Alloc(4, 4): want the block at offset 16")
	}

	// A failed allocation keeps the offset.
	_, err = arena.Alloc(1024, 1)
	if err != mem.ErrOutOfMemory {
		t.Error("Alloc(1024, 1): want ErrOutOfMemory")
	}
	p4, err := arena.Alloc(1, 1)
	if err != nil {
		t.Fatal("Alloc(1, 1) after a failed allocation failed")
		return
	}
	if p4.(*byte) != &buf[20] {
		t.Error("the failed allocation moved the offset")
	}

	// Reset returns the offset to 0.
	arena.Reset()
	p5, err := arena.Alloc(1, 1)
	if err != nil {
		t.Fatal("Alloc(1, 1) after Reset failed")
		return
	}
	if p5.(*byte) != &buf[0] {
		t.Error("Reset: want the block at offset 0")
	}
}

func TestArena_ExactFit(t *testing.T) {
	// A block as large as the buffer fits.
	buf := make([]byte, 256)
	arena := mem.NewArena(buf)

	p, err := arena.Alloc(256, 1)
	if err != nil {
		t.Fatal("Alloc(256, 1) failed")
		return
	}
	if p.(*byte) != &buf[0] {
		t.Error("Alloc(256, 1): want the block at offset 0")
	}

	// The arena is full now.
	_, err = arena.Alloc(1, 1)
	if err != mem.ErrOutOfMemory {
		t.Error("want ErrOutOfMemory when full")
	}
}

func TestArena_EmptyBuffer(t *testing.T) {
	// An arena over an empty buffer allocates nothing.
	var buf []byte
	arena := mem.NewArena(buf)
	_, err := arena.Alloc(1, 1)
	if err != mem.ErrOutOfMemory {
		t.Error("want ErrOutOfMemory")
	}
}

func TestArena_FreeSize(t *testing.T) {
	// The arena reclaims the last block only if the freed size
	// matches the allocated size.
	buf := make([]byte, 16)
	arena := mem.NewArena(buf)

	p, err := arena.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc(8, 1) failed")
		return
	}

	// A size that does not match keeps the block.
	arena.Free(p, 4, 1)
	p2, err := arena.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc(8, 1) after a mismatched free failed")
		return
	}
	if p2.(*byte) != &buf[8] {
		t.Error("a mismatched free reclaimed the block")
	}

	// The matching size reclaims the block.
	arena.Free(p2, 8, 1)
	p3, err := arena.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc(8, 1) after a matching free failed")
		return
	}
	if p3.(*byte) != &buf[8] {
		t.Error("a matching free did not reclaim the block")
	}
}

func TestArena_FreeNotLast(t *testing.T) {
	// Freeing a block that is not the last one is a no-op.
	buf := make([]byte, 24)
	arena := mem.NewArena(buf)

	p1, err := arena.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc for p1 failed")
		return
	}
	_, err = arena.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc for p2 failed")
		return
	}

	arena.Free(p1, 8, 1)
	p3, err := arena.Alloc(8, 1)
	if err != nil {
		t.Fatal("Alloc for p3 failed")
		return
	}
	if p3.(*byte) != &buf[16] {
		t.Error("the arena reused a block that is not the last one")
	}
}

func TestArena_OutOfMemory(t *testing.T) {
	buf := make([]byte, 16)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	_, err := mem.TryAllocSlice[byte](a, 32, 32)
	if err != mem.ErrOutOfMemory {
		t.Error("want ErrOutOfMemory")
	}

	// A failed allocation does not consume space,
	// so a block of the whole buffer still fits.
	s, err := mem.TryAllocSlice[byte](a, 16, 16)
	if err != nil {
		t.Error("exact fit failed")
		return
	}
	if len(s) != 16 {
		t.Error("unexpected len after exact fit")
	}

	// The arena is full now.
	_, err = mem.TryAllocSlice[byte](a, 1, 1)
	if err != mem.ErrOutOfMemory {
		t.Error("want ErrOutOfMemory when full")
	}
}

func TestArena_Alignment(t *testing.T) {
	buf := make([]byte, 64)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	// One byte first, so the next allocation needs alignment padding.
	one := mem.AllocSlice[byte](a, 1, 1)
	one[0] = 7

	p, err := mem.TryAlloc[int64](a)
	if err != nil {
		t.Fatal("aligned allocation failed")
		return
	}
	// A misaligned store traps under the sanitizers.
	*p = 1234567890123
	if *p != 1234567890123 {
		t.Error("unexpected value after an aligned store")
	}
	if one[0] != 7 {
		t.Error("the padded allocation overwrote the previous one")
	}
}

func TestArena_ReallocGrow(t *testing.T) {
	// The buffer holds the grown block, but not the old block plus a copy.
	// A successful grow proves the arena resized the block in place.
	buf := make([]byte, 8)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	s, err := mem.TryAllocSlice[byte](a, 4, 4)
	if err != nil {
		t.Fatal("initial allocation failed")
		return
	}
	s[0] = 11
	s[3] = 44

	s, err = mem.TryReallocSlice(a, s, 8, 8)
	if err != nil {
		t.Fatal("grow failed")
		return
	}
	if len(s) != 8 || cap(s) != 8 {
		t.Error("grow: unexpected len/cap")
	}
	if s[0] != 11 || s[3] != 44 {
		t.Error("grow: data not preserved")
	}
	if s[4] != 0 || s[7] != 0 {
		t.Error("grow: new bytes not zeroed")
	}
}

func TestArena_ReallocShrink(t *testing.T) {
	buf := make([]byte, 16)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	s := mem.AllocSlice[byte](a, 16, 16)
	s[0] = 11

	// Shrinking the last block returns its space to the arena.
	s = mem.ReallocSlice(a, s, 8, 8)
	if len(s) != 8 || cap(s) != 8 {
		t.Error("shrink: unexpected len/cap")
	}
	if s[0] != 11 {
		t.Error("shrink: data not preserved")
	}
	_, err := mem.TryAllocSlice[byte](a, 8, 8)
	if err != nil {
		t.Error("shrink: space not reclaimed")
	}
}

func TestArena_ReallocGrowNotLast(t *testing.T) {
	buf := make([]byte, 64)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	s1 := mem.AllocSlice[byte](a, 4, 4)
	s1[0] = 11
	s1[3] = 44
	s2 := mem.AllocSlice[byte](a, 4, 4) // s1 is not the last block now
	s2[0] = 99

	// Growing a block that is not the last one allocates
	// a new block and copies the data.
	s1 = mem.ReallocSlice(a, s1, 8, 8)
	if s1[0] != 11 || s1[3] != 44 {
		t.Error("grow not last: data not copied")
	}
	if s2[0] != 99 {
		t.Error("grow not last: the other block changed")
	}

	// The new block does not overlap the other block.
	s1[0] = 55
	if s2[0] != 99 {
		t.Error("grow not last: the blocks overlap")
	}
}

func TestArena_ReallocShrinkNotLast(t *testing.T) {
	buf := make([]byte, 16)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	s1 := mem.AllocSlice[byte](a, 8, 8)
	s1[0] = 11
	s2 := mem.AllocSlice[byte](a, 8, 8) // the arena is full now
	s2[0] = 22

	// Shrinking a block that is not the last one keeps the block
	// in place and does not return space to the arena.
	s1 = mem.ReallocSlice(a, s1, 4, 4)
	if s1[0] != 11 {
		t.Error("shrink not last: data not preserved")
	}
	if s2[0] != 22 {
		t.Error("shrink not last: the other block changed")
	}
	_, err := mem.TryAllocSlice[byte](a, 1, 1)
	if err != mem.ErrOutOfMemory {
		t.Error("shrink not last: space unexpectedly reclaimed")
	}
}

func TestArena_ReallocOutOfMemory(t *testing.T) {
	buf := make([]byte, 16)
	arena := mem.NewArena(buf)
	var a mem.Allocator = &arena

	s := mem.AllocSlice[byte](a, 8, 8)
	_, err := mem.TryReallocSlice(a, s, 32, 32)
	if err != mem.ErrOutOfMemory {
		t.Error("want ErrOutOfMemory")
	}
}
