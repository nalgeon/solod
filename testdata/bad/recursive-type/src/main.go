package main

// Only a struct can close a cycle through a forward declaration.
type StateFn func() StateFn

func main() {
	var s StateFn
	_ = s
}
