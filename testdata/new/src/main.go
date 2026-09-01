package main

type point struct {
	x, y int
}

type buf [4]int
type array [3]int
type grid [2][3]int

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
	{
		// new with an array type. A pointer to an unnamed array
		// type is not supported, so these cases use named types.
		a := new(array)
		if len(a) != 3 || a[0] != 0 || a[2] != 0 {
			panic("expected a == [0 0 0]")
		}
		a[2] = 42
		if a[2] != 42 {
			panic("expected a[2] == 42")
		}
		b := new(buf)
		if len(b) != 4 || b[3] != 0 {
			panic("expected b == [0 0 0 0]")
		}
		m := new(grid)
		if len(m) != 2 || len(m[0]) != 3 || m[1][2] != 0 {
			panic("expected m == [[0 0 0] [0 0 0]]")
		}
	}
	{
		// new with an array variable
		aval := array{5, 6, 7}
		c := new(aval)
		if c[1] != 6 {
			panic("expected c[1] == 6")
		}
		c[1] = 8
		if aval[1] != 8 {
			panic("expected aval[1] == 8")
		}
	}
}
