package main

type base struct{ n int }

func (b *base) square() int { return b.n * b.n }

type number struct {
	base
}

func main() {
	var n number
	n.n = 5
	println(n.square())
}
