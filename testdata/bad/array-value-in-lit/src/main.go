package main

type Grid struct {
	cells [2]int
}

func main() {
	cells := [2]int{1, 2}
	// C only allows a brace initializer for an array field.
	g := Grid{cells: cells}
	_ = g
}
