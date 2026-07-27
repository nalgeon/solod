package main

import "unsafe"

type number int

type point struct {
	x, y int
}

type shape interface {
	area() int
}

type rect struct {
	width, height int
}

func (r *rect) area() int {
	return r.width * r.height
}

func acceptAny(v any) {
	_ = v
}

func acceptByte(v *byte) {
	_ = v
}

func acceptPoint(v *point) {
	_ = v
}

func acceptShape(v shape) {
	_ = v
}

func main() {
	{
		// Nil value.
		var n any
		acceptAny(n)
		acceptAny(any(n))
	}
	{
		// Integer value.
		n := 42
		acceptAny(n)
		acceptAny(any(n))
		acceptByte(any(n).(*byte))
		acceptAny(42)
	}
	{
		// Integer pointer.
		nval := 42
		n := &nval
		acceptAny(n)
		acceptAny(any(n))
		acceptByte(any(n).(*byte))
	}
	{
		// Unsafe pointer.
		nval := 42
		p := unsafe.Pointer(&nval)
		acceptAny(p)
		acceptAny(any(p))
		if any(p).(unsafe.Pointer) != p {
			panic("want any(p).(unsafe.Pointer) == p")
		}
	}
	{
		// String value.
		s := "hello"
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
		acceptAny("hello")
	}
	{
		// String pointer.
		sval := "hello"
		s := &sval
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
	}
	{
		// Slice value.
		s := []int{1, 2, 3}
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
		acceptAny([]int{1, 2, 3})
	}
	{
		// Slice pointer.
		sval := []int{1, 2, 3}
		s := &sval
		acceptAny(s)
		acceptAny(any(s))
		acceptByte(any(s).(*byte))
	}
	{
		// Struct value.
		p := point{1, 2}
		acceptAny(p)
		acceptAny(any(p))
		acceptPoint(any(p).(*point))
		acceptAny(point{1, 2})
	}
	{
		// Struct pointer.
		pval := point{1, 2}
		p := &pval
		acceptAny(p)
		acceptAny(any(p))
		acceptPoint(any(p).(*point))
	}
	{
		// Interface value.
		var s shape = &rect{width: 10, height: 5}
		acceptAny(s)
		acceptAny(any(s))
		acceptShape(any(s).(shape))
	}
	{
		// Any value casts.
		var i int = 42
		var a any = i
		if a.(int) != 42 {
			panic("want a.(int) == 42")
		}
		var n number = 42
		a = n
		if a.(number) != 42 {
			panic("want a.(number) == 42")
		}
		var s string = "hello"
		a = s
		if a.(string) != "hello" {
			panic("want a.(string) == \"hello\"")
		}
		var p point = point{1, 2}
		a = p
		ap := a.(point)
		if ap.x != 1 || ap.y != 2 {
			panic("want a.(point) == point{1, 2}")
		}
	}
	{
		// Any pointer casts.
		var i int = 42
		var a any = &i
		if a.(*int) != &i {
			panic("want a.(*int) == &i")
		}
		var n number = 42
		a = &n
		if a.(*number) != &n {
			panic("want a.(*number) == &n")
		}
		var s string = "hello"
		a = &s
		if a.(*string) != &s {
			panic("want a.(*string) == &s")
		}
		var p1 point = point{1, 2}
		a = &p1
		if a.(*point) != &p1 {
			panic("want a.(*point) == &p1")
		}
	}
	{
		// Any interface casts.
		var a any
		var r rect = rect{width: 10, height: 5}
		sh := shape(&r)
		a = sh
		ashape := a.(shape)
		if ashape.area() != r.area() {
			panic("want a.(shape) == shape(&r)")
		}
	}
}
