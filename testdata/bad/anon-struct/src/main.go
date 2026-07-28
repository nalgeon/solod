package main

func main() {
	// An array element needs a C type name, which an anonymous struct lacks.
	var items [2]struct{ x int }
	_ = items
}
