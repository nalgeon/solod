package main

const someConst = 7
const SomeConst = 7

var someVar = 42
var SomeVar = 42

func someFunc(x int, y int) bool {
	return x > y+someConst
}

func SomeFunc(x int, y int) bool {
	return x > y+someVar
}

func main() {
	_ = someFunc(1, 2)
	SomeFunc(3, 4)
}
