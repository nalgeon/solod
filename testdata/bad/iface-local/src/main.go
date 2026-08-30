package main

type rect struct{ w, h int }

func (r *rect) area() int { return r.w * r.h }

func main() {
	type shape interface {
		area() int
	}
	var s shape = &rect{2, 3}
	_ = s.area()
}
