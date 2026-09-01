package main

func main() {
	a := [2]int{1, 2}
	s := []*[2]int{&a}
	println(s[0][1])
}
