package main

type point struct {
	x, y int
}

func mkPoint() point {
	return point{1, 2}
}

func main() {
	var a any = mkPoint()
	_ = a
}
