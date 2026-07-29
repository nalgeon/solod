package main

func main() {
	// The tag goes into a temporary, and C has no array assignment,
	// so an array tag has nowhere to go.
	a := [3]int{1, 2, 3}
	switch a {
	case [3]int{3, 2, 1}:
		println("first")
	case [3]int{1, 2, 3}:
		println("second")
	}
}
