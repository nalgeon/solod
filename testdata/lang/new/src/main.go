package main

type point struct {
	x, y int
}

func main() {
	{
		// new with type
		n := new(int)
		if n == nil || *n != 0 {
			panic("expected n == 0")
		}
		p := new(point)
		if p == nil || p.x != 0 || p.y != 0 {
			panic("expected p.x == 0 && p.y == 0")
		}
	}
	{
		// new with value
		n := new(42)
		if n == nil || *n != 42 {
			panic("expected n == 42")
		}
		p1 := new(point{1, 2})
		if p1 == nil || p1.x != 1 || p1.y != 2 {
			panic("expected p1.x == 1 && p1.y == 2")
		}
		pval := point{3, 4}
		_ = pval
		p2 := new(pval)
		if p2 == nil || p2.x != 3 || p2.y != 4 {
			panic("expected p2.x == 3 && p2.y == 4")
		}
	}
}
