package main

func divmod(a, b int) (int, int) {
	return a / b, a % b
}

func check(d, m int) {
	println(d, m)
}

func main() {
	check(divmod(10, 3))
}
