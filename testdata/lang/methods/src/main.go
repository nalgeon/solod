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

type circle struct {
	radius int
}

func (c *circle) area() int {
	return 3 * c.radius * c.radius
}

// Methods on primitive types are not supported.
// type HttpStatus int

// func (s HttpStatus) String() string {
// 	switch s {
// 	case 200:
// 		return "OK"
// 	case 404:
// 		return "Not Found"
// 	case 500:
// 		return "Error"
// 	default:
// 		return "Other"
// 	}
// }

func main() {
	r := Rect{width: 10, height: 5}

	rArea := r.Area()
	_ = rArea
	rPerim := r.perim(2)
	_ = rPerim

	rp := &r
	rpArea := rp.Area()
	_ = rpArea
	rpPerim := rp.perim(2)
	_ = rpPerim

	c := circle{radius: 7}
	cArea := c.area()
	_ = cArea
}
