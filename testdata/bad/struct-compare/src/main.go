package main

type Point struct {
	x, y int
}

func main() {
	a := Point{1, 2}
	b := Point{1, 2}
	if a == b {
		println("equal")
	}
}
