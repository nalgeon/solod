package main

func value() int {
	return 42
}

func main() {
	p := new(value())
	_ = p
}
