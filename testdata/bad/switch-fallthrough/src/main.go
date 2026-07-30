package main

func main() {
	// The switch becomes an if/else chain, which has no case to fall into.
	i := 1
	switch i {
	case 1:
		println("one")
		fallthrough
	case 2:
		println("two")
	}
}
