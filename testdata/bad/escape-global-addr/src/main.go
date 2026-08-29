package main

var p *int

func setup() {
	n := 42
	p = &n
}

func main() {
	setup()
	println(*p)
}
