package main

type shape interface {
	area() int
}

type circle struct {
	radius int
}

func (c circle) area() int {
	return c.radius * c.radius
}

func main() {
	c := circle{radius: 5}
	var s shape = &c
	println(s.area())
}
