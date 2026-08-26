package main

//so:embed main.h
var main_h string

//so:inline
func identity[T any](val T) T {
	return val
}

//so:inline
func setPtr[T any](ptr *T, val T) {
	*ptr = val
}

// increment adds two to n.
//
//so:inline
func increment[T int](n T) T {
	// A line comment in a macro becomes a block comment.
	_n := n
	// A */ in the text does not close the block comment.
	_n = _n + 1
	/* A block comment stays as it is. */
	_n = _n + 1
	return _n
}

//so:inline
func a[T int](n T) T {
	var _some int = 11
	_ = _some
	_x := b(n) + 1
	return _x
}

//so:inline
func b[T int](n T) T {
	var _some float64 = 22.2
	_ = _some
	_x := c(n) + 1
	return _x
}

//so:inline
func c[T int](n T) T {
	var _some string = "33"
	_ = _some
	_x := n + 1
	return _x
}

// Blank parameters.
//
//so:inline
func pickThird[T any](_ T, _ T, c T) T {
	return c
}

//so:inline
func work[T any](v *T) (*T, error) {
	return v, nil
}

//so:extern
type Box[T any] struct {
	val T
}

//so:inline
func (b *Box[T]) set(val T) {
	b.val = val
}

func main() {
	println("lang/macro - start")
	{
		print("lang/macro: Function with return")
		x := identity(42)
		if x != int(42) {
			panic("x != 42")
		}
		println(" - ok")
	}
	{
		print("lang/macro: Function w/o return")
		var y int
		setPtr(&y, 42)
		if y != 42 {
			panic("y != 42")
		}
		println(" - ok")
	}
	{
		print("lang/macro: Pass an expression as an argument")
		x := increment(1 + 1)
		if x != 4 {
			panic("x != 4")
		}
		println(" - ok")
	}
	{
		print("lang/macro: Nested calls with variable shadowing")
		z := a(42)
		if z != 45 {
			panic("z != 45")
		}
		println(" - ok")
	}
	{
		print("lang/macro: Blank parameters")
		x := pickThird(1, 2, 3)
		if x != 3 {
			panic("x != 3")
		}
		println(" - ok")
	}
	{
		print("lang/macro: Generic method")
		var b Box[int]
		b.set(42)
		if b.val != 42 {
			panic("b.val != 42")
		}
		println(" - ok")
	}
	{
		print("lang/macro: Multi-return")
		var v int = 42
		res, err := work(&v)
		if err != nil {
			panic("err != nil")
		}
		if *res != 42 {
			panic("res != 42")
		}
		println(" - ok")
	}
	println("lang/macro - ok")
}
