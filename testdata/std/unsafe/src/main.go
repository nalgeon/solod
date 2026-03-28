package main

import "unsafe"

type point struct {
	x, y int
}

func main() {
	{
		// Sizeof.
		var x int = 42
		size := unsafe.Sizeof(x)
		if size != 8 {
			panic("want size == 8")
		}

		var p = point{1, 2}
		size = unsafe.Sizeof(p)
		if size != 16 {
			panic("want size == 16")
		}
	}
	{
		// Alignof.
		var x int = 42
		align := unsafe.Alignof(x)
		if align != 8 {
			panic("want align == 8")
		}

		var p = point{1, 2}
		align = unsafe.Alignof(p)
		if align != 8 {
			panic("want align == 8")
		}
	}
	// {
	// 	// Offsetof is not supported.
	// 	var p = point{1, 2}
	// 	offsetX := unsafe.Offsetof(p.x)
	// 	if offsetX != 0 {
	// 		panic("want offsetX == 0")
	// 	}
	// 	offsetY := unsafe.Offsetof(p.y)
	// 	if offsetY != 8 {
	// 		panic("want offsetY == 8")
	// 	}
	// }
	{
		// String.
		var b = []byte("hello")
		s := unsafe.String(&b[0], len(b))
		if s != "hello" {
			panic("want s == 'hello'")
		}
	}
	{
		// StringData.
		var s = "hello"
		b := unsafe.StringData(s)
		if *b != 'h' {
			panic("want *b == 'h'")
		}
	}
	{
		// Slice.
		var a = [5]int{1, 2, 3, 4, 5}
		slice := unsafe.Slice(&a[0], len(a))
		if len(slice) != 5 {
			panic("want len(slice) == 5")
		}
		if slice[0] != 1 || slice[4] != 5 {
			panic("want slice[0] == 1 and slice[4] == 5")
		}
	}
	{
		// SliceData.
		var s = []int{1, 2, 3, 4, 5}
		p := unsafe.SliceData(s)
		if *p != 1 {
			panic("want *p == 1")
		}
	}
	{
		// Pointer.
		var x int = 42
		p := unsafe.Pointer(&x)
		if *(*int)(p) != 42 {
			panic("want *(int*)p == 42")
		}
	}
}
