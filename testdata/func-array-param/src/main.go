package main

// A named function type with array parameters.
type SumFn func(a [3]int) int
type SumFn2D func(m [2][3]int) int
type SumFn3D func(m [2][3][4]int) int

// A named array type as a function parameter.
type Matrix [2][3]int
type SumFnNamed func(m Matrix) int

// A struct field of a function type with array parameters.
type Calc struct {
	sum   func(a [3]int) int
	sum2D func(m [2][3]int) int
	sum3D func(m [2][3][4]int) int
}

// An interface method with array parameters.
type Summer interface {
	Sum(a [3]int) int
	Sum2D(m [2][3]int) int
	Sum3D(m [2][3][4]int) int
}

type calc struct{}

func (_ *calc) Sum(a [3]int) int {
	return sum(a)
}

func (_ *calc) Sum2D(m [2][3]int) int {
	return sum2D(m)
}

func (_ *calc) Sum3D(m [2][3][4]int) int {
	return sum3D(m)
}

func sum(a [3]int) int {
	return a[0] + a[1] + a[2]
}

func sum2D(m [2][3]int) int {
	total := 0
	for i := range 2 {
		for j := range 3 {
			total += m[i][j]
		}
	}
	return total
}

func sum3D(m [2][3][4]int) int {
	return m[0][0][0] + m[1][2][3]
}

func sumNamed(m Matrix) int {
	return m[0][0] + m[1][2]
}

// An anonymous function type with array parameters as an argument.
func apply(f func(a [3]int) int, a [3]int) int {
	return f(a)
}

func main() {
	a := [3]int{1, 2, 3}
	m := [2][3]int{{1, 2, 3}, {4, 5, 6}}
	var m3D [2][3][4]int
	m3D[0][0][0] = 10
	m3D[1][2][3] = 11

	{
		// Named function type.
		var f SumFn = sum
		if f(a) != 6 {
			panic("unexpected f")
		}

		var f2D SumFn2D = sum2D
		if f2D(m) != 21 {
			panic("unexpected f2D")
		}

		var f3D SumFn3D = sum3D
		if f3D(m3D) != 21 {
			panic("unexpected f3D")
		}

		var fn SumFnNamed = sumNamed
		if fn(Matrix{{1, 2, 3}, {4, 5, 6}}) != 7 {
			panic("unexpected fn")
		}
	}
	{
		// Anonymous function type variable.
		var g func(a [3]int) int = sum
		if g(a) != 6 {
			panic("unexpected g")
		}

		var g2D func(m [2][3]int) int = sum2D
		if g2D(m) != 21 {
			panic("unexpected g2D")
		}

		var g3D func(m [2][3][4]int) int = sum3D
		if g3D(m3D) != 21 {
			panic("unexpected g3D")
		}
	}
	{
		// Struct field.
		c := Calc{sum: sum, sum2D: sum2D, sum3D: sum3D}
		if c.sum(a) != 6 {
			panic("unexpected c.sum")
		}
		if c.sum2D(m) != 21 {
			panic("unexpected c.sum2D")
		}
		if c.sum3D(m3D) != 21 {
			panic("unexpected c.sum3D")
		}
	}
	{
		// Interface method.
		c := calc{}
		var s Summer = &c
		if s.Sum(a) != 6 {
			panic("unexpected s.Sum")
		}
		if s.Sum2D(m) != 21 {
			panic("unexpected s.Sum2D")
		}
		if s.Sum3D(m3D) != 21 {
			panic("unexpected s.Sum3D")
		}
	}
	{
		// Anonymous function type as an argument.
		if apply(sum, a) != 6 {
			panic("unexpected apply")
		}
	}
}
