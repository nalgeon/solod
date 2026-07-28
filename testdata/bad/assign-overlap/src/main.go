package main

func main() {
	a, b := 1, 2
	// C assigns left to right, so a swap would set both to the old b.
	a, b = b, a
	_ = a
	_ = b
}
