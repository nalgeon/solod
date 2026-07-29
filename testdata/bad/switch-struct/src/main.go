package main

type Point struct {
	x, y int
}

func main() {
	// Case comparisons follow the rules of ==, which rejects structs.
	p := Point{1, 2}
	switch p {
	case Point{1, 2}:
		println("origin")
	}
}
