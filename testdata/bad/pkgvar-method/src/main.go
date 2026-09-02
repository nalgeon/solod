package main

type Rect struct {
	w, h int
}

func (r Rect) Area() int { return r.w * r.h }

var unit = Rect{2, 3}

var area = unit.Area()

func main() {
	println(area)
}
