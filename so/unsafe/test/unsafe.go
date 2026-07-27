package main

import (
	"unsafe"

	"solod.dev/so/testing"
)

type point struct {
	x, y int
}

const ptrSize = 4 << (uint64(^uintptr(0)) >> 63)

func TestSizeof(t *testing.T) {
	var x int = 42
	size := unsafe.Sizeof(x)
	if size != ptrSize {
		t.Error("invalid sizeof(int)")
	}

	var p = point{1, 2}
	size = unsafe.Sizeof(p)
	if size != 2*ptrSize {
		t.Error("invalid sizeof(point)")
	}
}

func TestAlignof(t *testing.T) {
	var x int = 42
	align := unsafe.Alignof(x)
	if align != ptrSize {
		t.Error("invalid alignof(int)")
	}

	var p = point{1, 2}
	align = unsafe.Alignof(p)
	if align != ptrSize {
		t.Error("invalid alignof(point)")
	}
}

func TestString(t *testing.T) {
	var b = []byte("hello")
	s := unsafe.String(&b[0], len(b))
	if s != "hello" {
		t.Error("want s == 'hello'")
	}
}

func TestStringData(t *testing.T) {
	var s = "hello"
	b := unsafe.StringData(s)
	if *b != 'h' {
		t.Error("want *b == 'h'")
	}
}

func TestSlice(t *testing.T) {
	var a = [5]int{1, 2, 3, 4, 5}
	slice := unsafe.Slice(&a[0], len(a))
	if len(slice) != 5 {
		t.Error("want len(slice) == 5")
	}
	if slice[0] != 1 || slice[4] != 5 {
		t.Error("want slice[0] == 1 and slice[4] == 5")
	}
}

func TestSliceData(t *testing.T) {
	var s = []int{1, 2, 3, 4, 5}
	p := unsafe.SliceData(s)
	if *p != 1 {
		t.Error("want *p == 1")
	}
}

func TestPointer(t *testing.T) {
	var x int = 42
	p := unsafe.Pointer(&x)
	if *(*int)(p) != 42 {
		t.Error("want *(int*)p == 42")
	}

	// An unsafe pointer is stored in an any as is, not by address.
	var a any = p
	if *(*int)(a.(unsafe.Pointer)) != 42 {
		t.Error("want *(int*)a == 42")
	}
}
