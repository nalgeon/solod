package main

var xy string

func add(s string) {
	xy += s
}

func main() {
	add("x")
	println(xy)
}
