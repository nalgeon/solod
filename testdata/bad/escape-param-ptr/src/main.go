package main

func concat(out *string, a, b string) {
	*out = a + b
}

func main() {
	var s string
	concat(&s, "x", "y")
	println(s)
}
