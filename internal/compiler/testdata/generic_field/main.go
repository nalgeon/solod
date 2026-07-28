package main

// There is no C type to emit for T.
type Pair[T any] struct {
	a, b T
}

func main() {
	var p Pair[int]
	_ = p
}
