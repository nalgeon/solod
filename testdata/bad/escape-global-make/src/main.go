package main

var buf []int

func setup() {
	buf = make([]int, 3)
}

func main() {
	setup()
	println(len(buf))
}
