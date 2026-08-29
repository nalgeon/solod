package main

var first = &second
var second = 5

func main() {
	println(*first, second)
}
