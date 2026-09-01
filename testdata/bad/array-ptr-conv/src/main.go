package main

func main() {
	s := []int{1, 2, 3}
	p := (*[3]int)(s)
	_ = p
}
