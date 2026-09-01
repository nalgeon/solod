package main

type point struct {
	x, y int
}

func main() {
	s := make([]any, 0, 1)
	s = append(s, &point{1, 2})
	_ = s
}
