package main

func fill(out []string, a, b string) {
	out[0] = a + b
}

func main() {
	s := make([]string, 1)
	fill(s, "x", "y")
	println(s[0])
}
