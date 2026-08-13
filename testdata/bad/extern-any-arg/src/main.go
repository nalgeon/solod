package main

//so:extern nodecay
func measure(kinds string, args ...any) int

func report(a any) int {
	// The callee reads one C type per argument, and an any carries none.
	return measure("i", a)
}

func main() {
	println(report(1))
}
