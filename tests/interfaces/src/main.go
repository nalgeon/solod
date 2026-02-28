package main

type Shape interface {
	Area() int
	Perim(n int) int
}

type Line interface {
	Length() int
}

type Rect struct {
	width, height int
}

func (r Rect) Area() int {
	return r.width * r.height
}

func (r Rect) Perim(n int) int {
	return n * (2*r.width + 2*r.height)
}

func (r *Rect) Length() int {
	return 2*r.width + 2*r.height
}

func calc(s Shape) int {
	return s.Perim(2) + s.Area()
}

func shapeIsRect(s Shape) bool {
	_, ok := s.(Rect)
	return ok
}

func shapeAsRect(s Shape) int {
	_, ok := s.(Rect)
	if !ok {
		return 0
	}
	r := s.(Rect)
	return r.Area()
}

func lineIsRect(l Line) bool {
	_, ok := l.(*Rect)
	return ok
}

func lineAsRect(l Line) *Rect {
	_, ok := l.(*Rect)
	if !ok {
		return nil
	}
	r := l.(*Rect)
	return r
}

type reader interface {
	Read(p []byte) (n int, err error)
}

func main() {
	r := Rect{width: 10, height: 5}
	{
		// Shape interface is implemented by Rect value.
		s := Shape(r)
		calc(s)
		calc(Shape(r)) // also works
		calc(r)        // also works

		_ = shapeIsRect(s)
		a := shapeAsRect(s)
		_ = a
	}
	{
		// Line interface is implemented by *Rect pointer.
		l := Line(&r)
		_ = lineIsRect(l)
		rptr := lineAsRect(l)
		_ = rptr
	}
	{
		var rdr reader
		_ = rdr
	}
}
