package main

func main() {
	m := make(map[string]int, 2)
	for i := 0; i < 1; m["k"] = i {
		i++
	}
}
