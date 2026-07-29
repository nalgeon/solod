package main

type Point struct {
	x, y int
}

func main() {
	// The tag type is rejected even when there is nothing to compare it to.
	p := Point{1, 2}
	switch p {
	default:
		println("default")
	}
}
