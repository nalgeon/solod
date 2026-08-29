package main

type box struct {
	s string
}

func (b *box) set(a, c string) {
	b.s = a + c
}

func main() {
	var v box
	v.set("x", "y")
	println(v.s)
}
