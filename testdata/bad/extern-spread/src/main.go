package main

//so:extern nodecay
func measure(kinds string, args ...any) int

func report(args ...any) int {
	// A C variadic takes one value per argument, so a slice cannot spread.
	return measure("ii", args...)
}

func main() {
	println(report(1, 2))
}
