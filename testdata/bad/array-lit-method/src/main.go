package main

type T [3]int

func (t T) first() int { return t[0] }

func main() {
	x := T{1, 2, 3}.first()
	_ = x
}
