package main

type Rect struct {
	width, height int
}

func (r *Rect) Area() int { return r.width * r.height }

type Shape interface{ Area() int }
type Sized interface{ Area() int }

func main() {
	var s Shape = &Rect{2, 4}
	other := s.(Sized)
	_ = other
}
