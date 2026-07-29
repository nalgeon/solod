package main

type Shape interface {
	Area() int
}

type Square struct {
	side int
}

func (s *Square) Area() int {
	return s.side * s.side
}

func main() {
	sq := &Square{2}
	var sh Shape = sq
	// The interface holds a pointer to the value, so comparing it with
	// the value itself would compare an address with a value.
	if sh == sq {
		println("same")
	}
}
