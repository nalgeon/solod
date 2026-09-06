package main

type Rect struct {
	width, height int
}

func (r *Rect) Area() int {
	return r.width * r.height
}

type Shape interface {
	Area() int
}

// RectPtr aliases a pointer to a named type.
type RectPtr = *Rect

// Canvas aliases a named interface.
type Canvas = Shape

type Frame struct {
	rect RectPtr
}

func areaOf(s Shape) int {
	return s.Area()
}

func asShape(r RectPtr) Shape {
	return r
}

func main() {
	r := Rect{width: 10, height: 5}
	{
		// An alias to a pointer converts to an interface.
		var s Shape = RectPtr(nil)
		var p RectPtr = &r
		s = p
		if s.Area() != 50 {
			panic("s.Area() != 50")
		}
	}
	{
		// An alias to a pointer passes as an interface argument and result.
		var p RectPtr = &r
		if areaOf(p) != 50 {
			panic("areaOf(p) != 50")
		}
		if asShape(p).Area() != 50 {
			panic("asShape(p).Area() != 50")
		}
	}
	{
		// A struct field of an alias type converts to an interface.
		f := Frame{rect: &r}
		var s Shape = f.rect
		if s.Area() != 50 {
			panic("s.Area() != 50")
		}
	}
	{
		// A type assertion accepts an alias.
		var s Shape = &r
		p := s.(RectPtr)
		if p.Area() != 50 {
			panic("p.Area() != 50")
		}
		if _, ok := s.(RectPtr); !ok {
			panic("want ok == true")
		}
	}
	{
		// An alias to a named interface holds a concrete value.
		var c Canvas = &r
		if c.Area() != 50 {
			panic("c.Area() != 50")
		}
	}
	{
		// A method expression accepts an alias receiver.
		area := RectPtr.Area
		if area(&r) != 50 {
			panic("area(&r) != 50")
		}
	}
	{
		// Two variables of an alias pointer type declare separately.
		r2 := Rect{width: 3, height: 4}
		var p1, p2 RectPtr = &r, &r2
		if p1.Area() != 50 {
			panic("p1.Area() != 50")
		}
		if p2.Area() != 12 {
			panic("p2.Area() != 12")
		}
	}
	{
		// print accepts an alias to a pointer.
		var p RectPtr
		println(p)
	}
}
