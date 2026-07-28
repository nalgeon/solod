package main

func main() {
	x := 1
	if x > 0 {
		// In C the right-hand x would refer to the new variable.
		x := x + 1
		_ = x
	}
}
