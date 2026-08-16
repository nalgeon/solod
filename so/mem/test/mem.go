package mem_test

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestClear(t *testing.T) {
	buf := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	mem.Clear(&buf[2], 4)
	if buf != [8]byte{1, 2, 0, 0, 0, 0, 7, 8} {
		t.Error("Clear: unexpected buffer")
	}

	// A zero size clears nothing.
	mem.Clear(&buf[0], 0)
	if buf[0] != 1 {
		t.Error("Clear: zero size cleared a byte")
	}
}

func TestCompare(t *testing.T) {
	a := [4]byte{1, 2, 3, 4}
	b := [4]byte{1, 2, 3, 4}
	big := [4]byte{1, 2, 3, 5}

	if mem.Compare(&a[0], &b[0], 4) != 0 {
		t.Error("Compare(a, b) != 0")
	}
	// Compare normalizes the result to -1 or +1.
	if mem.Compare(&a[0], &big[0], 4) != -1 {
		t.Error("Compare(a, big) != -1")
	}
	if mem.Compare(&big[0], &a[0], 4) != 1 {
		t.Error("Compare(big, a) != +1")
	}
	// Compare reads only the given number of bytes.
	if mem.Compare(&a[0], &big[0], 3) != 0 {
		t.Error("Compare(a, big, 3) != 0")
	}
	if mem.Compare(&a[0], &big[0], 0) != 0 {
		t.Error("Compare(a, big, 0) != 0")
	}
}

func TestCompareUnsigned(t *testing.T) {
	// Compare reads the bytes as unsigned values.
	low := [1]byte{0x01}
	high := [1]byte{0xff}

	if mem.Compare(&high[0], &low[0], 1) != 1 {
		t.Error("Compare(0xff, 0x01) != +1")
	}
	if mem.Compare(&low[0], &high[0], 1) != -1 {
		t.Error("Compare(0x01, 0xff) != -1")
	}

	// The first byte that differs decides the result.
	a := [3]byte{0x00, 0xff, 0x00}
	b := [3]byte{0x00, 0x01, 0xff}
	if mem.Compare(&a[0], &b[0], 3) != 1 {
		t.Error("Compare(a, b) != +1")
	}
}

func TestCopy(t *testing.T) {
	src := [4]byte{11, 22, 33, 44}
	dst := [4]byte{0, 0, 0, 0}

	res := mem.Copy(&dst[0], &src[0], 4)
	if dst != src {
		t.Error("Copy: dst != src")
	}
	// Copy returns dst.
	p := res.(*byte)
	if p != &dst[0] {
		t.Error("Copy: result != dst")
	}

	// Copy writes only the given number of bytes.
	part := [4]byte{9, 9, 9, 9}
	mem.Copy(&part[0], &src[0], 2)
	if part != [4]byte{11, 22, 9, 9} {
		t.Error("Copy: wrote past the given size")
	}

	// A zero size copies nothing.
	mem.Copy(&part[0], &src[0], 0)
	if part != [4]byte{11, 22, 9, 9} {
		t.Error("Copy: a zero size wrote a byte")
	}
}

func TestMove(t *testing.T) {
	// Move forward. The source and the destination overlap.
	fwd := [6]byte{1, 2, 3, 4, 5, 6}
	res := mem.Move(&fwd[2], &fwd[0], 4)
	if fwd != [6]byte{1, 2, 1, 2, 3, 4} {
		t.Error("Move forward: unexpected buffer")
	}
	// Move returns dst.
	p := res.(*byte)
	if p != &fwd[2] {
		t.Error("Move: result != dst")
	}

	// Move backward. The source and the destination overlap.
	back := [6]byte{1, 2, 3, 4, 5, 6}
	mem.Move(&back[0], &back[2], 4)
	if back != [6]byte{3, 4, 5, 6, 5, 6} {
		t.Error("Move backward: unexpected buffer")
	}

	// The source and the destination are the same.
	same := [4]byte{1, 2, 3, 4}
	mem.Move(&same[0], &same[0], 4)
	if same != [4]byte{1, 2, 3, 4} {
		t.Error("Move to itself: unexpected buffer")
	}

	// A zero size moves nothing.
	mem.Move(&same[0], &same[2], 0)
	if same != [4]byte{1, 2, 3, 4} {
		t.Error("Move: a zero size wrote a byte")
	}
}

func TestSwap(t *testing.T) {
	x, y := 11, 22
	mem.Swap(&x, &y)
	if x != 22 || y != 11 {
		t.Error("Swap(int): unexpected values")
	}

	p1 := Point{x: 11, y: 22}
	p2 := Point{x: 33, y: 44}
	mem.Swap(&p1, &p2)
	if p1.x != 33 || p1.y != 44 {
		t.Error("Swap(Point): unexpected p1")
	}
	if p2.x != 11 || p2.y != 22 {
		t.Error("Swap(Point): unexpected p2")
	}

	// Swapping a value with itself keeps the value.
	mem.Swap(&x, &x)
	if x != 22 {
		t.Error("Swap(x, x): unexpected value")
	}

	// Swap works on a byte-sized type.
	b1, b2 := byte(1), byte(2)
	mem.Swap(&b1, &b2)
	if b1 != 2 || b2 != 1 {
		t.Error("Swap(byte): unexpected values")
	}
}

func TestSwapByte(t *testing.T) {
	a := [4]byte{1, 2, 3, 4}
	b := [4]byte{5, 6, 7, 8}

	mem.SwapByte(&a[0], &b[0], 4)
	if a != [4]byte{5, 6, 7, 8} {
		t.Error("SwapByte: unexpected a")
	}
	if b != [4]byte{1, 2, 3, 4} {
		t.Error("SwapByte: unexpected b")
	}

	// SwapByte exchanges only the given number of bytes.
	mem.SwapByte(&a[0], &b[0], 2)
	if a != [4]byte{1, 2, 7, 8} {
		t.Error("SwapByte: swapped past the given size")
	}
	if b != [4]byte{5, 6, 3, 4} {
		t.Error("SwapByte: swapped past the given size")
	}

	// A zero size swaps nothing.
	mem.SwapByte(&a[0], &b[0], 0)
	if a != [4]byte{1, 2, 7, 8} || b != [4]byte{5, 6, 3, 4} {
		t.Error("SwapByte: a zero size swapped a byte")
	}
}

func TestSwapByteStruct(t *testing.T) {
	// SwapByte works on any type, not only on bytes.
	p1 := Point{x: 11, y: 22}
	p2 := Point{x: 33, y: 44}

	mem.SwapByte(&p1, &p2, c.Sizeof[Point]())
	if p1.x != 33 || p1.y != 44 {
		t.Error("SwapByte(Point): unexpected p1")
	}
	if p2.x != 11 || p2.y != 22 {
		t.Error("SwapByte(Point): unexpected p2")
	}
}
