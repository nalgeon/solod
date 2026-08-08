package mem_test

import (
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
}
