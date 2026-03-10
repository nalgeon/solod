package main

// Same-type pair.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// Mixed types.
func check(n int) (bool, int) {
	return n > 0, n * 2
}

// String pair.
func greet(name string) (string, string) {
	return "hello", name
}

// Forwarding.
func forwardDivmod() (int, int) {
	return divmod(10, 3)
}

func main() {
	{
		// Destructure into new variables.
		q, r := divmod(10, 3)
		_ = q
		_ = r

		// Blank identifiers.
		_, r2 := divmod(10, 3)
		_ = r2
		q3, _ := divmod(10, 3)
		_ = q3

		// Partial reassignment.
		q4, r2 := divmod(20, 7)
		_ = q4

		// Assign to existing variables.
		q = 0
		r = 0
		q, r = divmod(20, 7)
	}

	{
		// Mixed types.
		ok, doubled := check(5)
		_ = ok
		_ = doubled
	}

	{
		// String pair.
		greeting, name := greet("world")
		_ = greeting
		_ = name
	}

	{
		// If-init with multi-return.
		if q, r := divmod(10, 3); r > 0 {
			_ = q
		}
	}

	{
		// Forwarding.
		q, r := forwardDivmod()
		_ = q
		_ = r
	}
}
