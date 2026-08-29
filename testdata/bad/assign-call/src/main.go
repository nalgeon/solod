package main

func idx() int {
	return 0
}

func main() {
	a := []int{100, 200}
	n := 3
	a[idx()] /= n
	println(a[0])
}
