package main

const huge = 1e200 * 1e200

func main() {
	var x float64 = huge / 1e200
	println(x)
}
