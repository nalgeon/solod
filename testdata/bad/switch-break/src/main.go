package main

func main() {
	// The switch becomes an if/else chain, so the emitted break
	// would leave the for loop instead.
	for i := range 3 {
		switch i {
		case 1:
			break
		}
		println(i)
	}
}
