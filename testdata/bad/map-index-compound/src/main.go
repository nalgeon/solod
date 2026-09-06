package main

func main() {
	m := map[string]int{"a": 1}
	m["a"] += 2
	println(m["a"])
}
