package main

func main() {
	x := 1
	if x > 0 {
		// Same as `x := x + 1`, but through a var declaration.
		var x = x + 1
		_ = x
	}
}
