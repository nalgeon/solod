package main

type array [3]int

func main() {
	a := new(array{1, 2, 3})
	_ = a
}
