package main

type point struct {
	x, y int
}

// Origin is exported, so point would leak into the header.
func Origin() point {
	return point{}
}

func main() {
	_ = Origin()
}
