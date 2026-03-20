package main

// Methods on struct types.
type Rect struct {
	width, height int
}

func (r *Rect) Area() int {
	return r.width * r.height
}

func (r *Rect) perim(n int) int {
	return n * (2*r.width + 2*r.height)
}

func (r Rect) resize(x int) Rect {
	r.height *= x
	r.width *= x
	return r
}

type circle struct {
	radius int
}

func (c *circle) area() int {
	return 3 * c.radius * c.radius
}

// Methods on primitive types are also supported.
type HttpStatus int

func (s HttpStatus) String() string {
	if s == 200 {
		return "OK"
	} else if s == 404 {
		return "Not Found"
	} else if s == 500 {
		return "Error"
	} else {
		return "Other"
	}
}

func main() {
	r := Rect{width: 10, height: 5}
	{
		// Value + pointer receiver.
		rArea := r.Area()
		if rArea != 50 {
			panic("unexpected area")
		}
		rPerim := r.perim(2)
		if rPerim != 60 {
			panic("unexpected perimeter")
		}
	}
	{
		// Pointer + pointer receiver.
		rp := &r
		rpArea := rp.Area()
		if rpArea != 50 {
			panic("unexpected area")
		}
		rpPerim := rp.perim(2)
		if rpPerim != 60 {
			panic("unexpected perimeter")
		}
	}
	{
		// Value + value receiver.
		rResized := r.resize(2)
		if r.width != 10 || r.height != 5 {
			panic("unexpected original rect")
		}
		if rResized.width != 20 || rResized.height != 10 {
			panic("unexpected resized rect")
		}
	}
	{
		// Pointer + value receiver.
		rp := &r
		rResized := rp.resize(2)
		if r.width != 10 || r.height != 5 {
			panic("unexpected original rect")
		}
		if rResized.width != 20 || rResized.height != 10 {
			panic("unexpected resized rect")
		}
	}
	{
		// Unexported type and method.
		c := circle{radius: 7}
		cArea := c.area()
		if cArea != 147 {
			panic("unexpected area")
		}
	}
	{
		// Method on primitive type.
		var s HttpStatus = 200
		if s.String() != "OK" {
			panic("unexpected string")
		}
		s = 404
		if s.String() != "Not Found" {
			panic("unexpected string")
		}
	}
}
