package mem_test

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestArray(t *testing.T) {
	arr := mem.NewArray(t.Allocator(), c.Sizeof[Point](), 3)
	defer arr.Free()

	if arr.Len() != 3 {
		t.Error("want arr.Len() == 3")
	}

	var p Point
	arr.Load(1, &p)
	if p.x != 0 || p.y != 0 {
		t.Error("want arr[1] == {0, 0}")
	}

	p1 := Point{x: 11, y: 22}
	p2 := Point{x: 33, y: 44}
	p3 := Point{x: 55, y: 66}
	arr.Store(0, &p1)
	arr.Store(1, &p2)
	arr.Store(2, &p3)

	arr.Load(0, &p)
	if p.x != 11 || p.y != 22 {
		t.Error("want arr[0] == {11, 22}")
	}

	arr.Load(1, &p)
	if p.x != 33 || p.y != 44 {
		t.Error("want arr[1] == {33, 44}")
	}

	// At returns a pointer into the storage.
	pp := arr.At(2).(*Point)
	if pp.x != 55 || pp.y != 66 {
		t.Error("want arr[2] == {55, 66}")
	}

	// The pointer stays valid until the slot is overwritten.
	p4 := Point{x: 77, y: 88}
	arr.Store(2, &p4)
	if pp.x != 77 || pp.y != 88 {
		t.Error("At: the pointer does not point into the storage")
	}
}

func TestArrayBytes(t *testing.T) {
	// A one-byte element size gives one slot per byte.
	arr := mem.NewArray(t.Allocator(), 1, 4)
	defer arr.Free()

	for i := range arr.Len() {
		v := byte(i + 1)
		arr.Store(i, &v)
	}
	for i := range arr.Len() {
		var v byte
		arr.Load(i, &v)
		if v != byte(i+1) {
			t.Errorf("arr[%d] = %d, want %d", i, v, i+1)
		}
	}

	// The slots follow each other in one allocation.
	first := arr.At(0).(*byte)
	if arr.At(3).(*byte) != c.PtrAdd(first, 3) {
		t.Error("At: the slots are not next to each other")
	}
}

func TestArrayIsolation(t *testing.T) {
	// A store writes vsize bytes into one slot only.
	arr := mem.NewArray(t.Allocator(), c.Sizeof[Point](), 3)
	defer arr.Free()

	zero := Point{}
	full := Point{x: -1, y: -1}
	arr.Store(0, &zero)
	arr.Store(1, &full)
	arr.Store(2, &zero)

	var p Point
	arr.Load(0, &p)
	if p.x != 0 || p.y != 0 {
		t.Error("the store into slot 1 changed slot 0")
	}
	arr.Load(2, &p)
	if p.x != 0 || p.y != 0 {
		t.Error("the store into slot 1 changed slot 2")
	}
}

func TestArrayFree(t *testing.T) {
	// Free releases the storage and empties the array.
	arr := mem.NewArray(t.Allocator(), c.Sizeof[Point](), 3)
	arr.Free()
	if arr.Len() != 0 {
		t.Error("want arr.Len() == 0 after Free")
	}
}
