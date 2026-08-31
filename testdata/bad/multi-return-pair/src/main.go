package main

func f() (int, *int) {
	return 0, nil
}

func main() {
	n, p := f()
	println(n, p != nil)
}
