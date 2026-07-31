package main

const huge = 1 << 200
const small = huge >> 190

func main() {
	var x uint64 = small
	println(x)
}
