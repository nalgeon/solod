package main

type Shape interface {
	Area() int
	Perim(n int) int
}

type Canvas struct {
	name  string
	shape Shape
}

type Rect struct {
	width, height int
}

func (r *Rect) Area() int {
	return r.width * r.height
}

func (r *Rect) Perim(n int) int {
	return n * (2*r.width + 2*r.height)
}

func calcShape(s Shape) int {
	return s.Perim(2) + s.Area()
}

func shapeIsRect(s Shape) bool {
	_, ok := s.(*Rect)
	return ok
}

func shapeAsRect(s Shape) *Rect {
	if _, ok := s.(*Rect); !ok {
		return nil
	}
	r := s.(*Rect)
	return r
}

func rectAsShape(r *Rect) Shape {
	return r
}

func shapeCheckAssign(s Shape) bool {
	var ok bool
	_, ok = s.(*Rect)
	return ok
}

func nilShape() Shape {
	return nil
}

func main() {
	r := Rect{width: 10, height: 5}
	{
		// Shape interface is implemented by *Rect pointer.
		s := Shape(&r)
		var s2 Shape = &r // also works
		_ = s2

		calcShape(s)
		calcShape(Shape(&r)) // also works
		calcShape(&r)        // also works

		_ = shapeIsRect(s)
		_ = shapeCheckAssign(s)
		rval := shapeAsRect(s)
		_ = rval
	}
	{
		// Wrap Rect value into Shape via function.
		s := rectAsShape(&r)
		_ = s
	}
	{
		// Nil interface. Nil works on either side.
		var s1 Shape
		if s1 != nil {
			panic("want nil interface")
		}
		if nil != s1 {
			panic("want nil interface")
		}
		var s2 Shape = nil
		if s2 != nil {
			panic("want nil interface")
		}
		s3 := nilShape()
		if s3 != nil {
			panic("want nil interface")
		}
		isRect := shapeIsRect(nil)
		if isRect {
			panic("want isRect == false")
		}
		var r Rect
		var s4 Shape = &r
		if s4 == nil {
			panic("want non-nil interface")
		}
		if nil == s4 {
			panic("want non-nil interface")
		}
	}
	{
		// Interface field in struct.
		c1 := Canvas{name: "c1", shape: &r}
		if c1.shape.Area() != 50 {
			panic("c1.shape.Area() != 50")
		}
		c2 := Canvas{name: "c2", shape: &Rect{5, 4}}
		if c2.shape.Area() != 20 {
			panic("c2.shape.Area() != 20")
		}
		c3 := Canvas{name: "c3", shape: nil}
		if c3.shape != nil {
			panic("c3.shape != nil")
		}
	}
	{
		// Interface field assignment.
		var c Canvas
		c.shape = &r
		if c.shape.Area() != 50 {
			panic("c.shape.Area() != 50")
		}
	}
	{
		// Existing interface in struct literal.
		s := Shape(&r)
		c := Canvas{name: "wrap", shape: s}
		if c.shape.Area() != 50 {
			panic("c.shape.Area() != 50")
		}
	}
	{
		// Multi-var interface declaration.
		r2 := Rect{width: 3, height: 4}
		var s1, s2 Shape = &r, &r2
		if s1.Area() != 50 {
			panic("s1.Area() != 50")
		}
		if s2.Area() != 12 {
			panic("s2.Area() != 12")
		}
	}
	{
		// Redeclared interface variable.
		var s Shape
		s, n := &r, 42
		_ = n
		if s.Area() != 50 {
			panic("s.Area() != 50")
		}
	}
}
