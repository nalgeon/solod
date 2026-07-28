package main

type Point struct {
	// A struct field cannot be mangled, unlike a local variable.
	double int
}

func main() {
	p := Point{double: 1}
	_ = p
}
