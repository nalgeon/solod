package main

type Inner struct{ v int }

type Outer struct{ in Inner }

var outer Outer

var ptr = &outer

var addr = &ptr.in

func main() {
	println(addr.v)
}
