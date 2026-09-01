package main

func main() {
	a := [2]int{1, 2}
	m := map[*[2]int]bool{}
	m[&a] = true
	println(len(m))
}
