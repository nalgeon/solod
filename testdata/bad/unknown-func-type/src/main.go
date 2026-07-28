package main

func twice(x int) int {
	return x * 2
}

// A return type needs a C type name, so it must be a named function type.
// An anonymous one is fine as a parameter (see the func-fields case).
func pick() func(int) int {
	return twice
}

func main() {
	_ = pick()
}
