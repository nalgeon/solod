package main

type Shape interface {
	self() int
}

func main() {
	var s Shape
	_ = s
}
