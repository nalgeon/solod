package main

func show(p struct{ x int }) {
	_ = p.x
}

func main() {
	show(struct{ x int }{1})
}
