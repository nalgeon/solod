package main

import _ "embed"

//so:embed main.h
var header string

//so:extern
func newObj[T any]() *T {
	return nil
}

//so:extern
func freeObj[T any](ptr *T) {
}

//so:extern
type Map[K comparable, V any] struct {
	len int
}

//so:extern
func (m *Map[K, V]) Len() int {
	return m.len
}

//so:extern
func newMap[K comparable, V any](size int) Map[K, V] {
	return Map[K, V]{len: size}
}

type Stringer interface {
	String() string
}

// Constraints are not emitted in C.
type Number interface {
	~int | ~float64
}

type ordered interface {
	comparable
	~int | ~string
}

type named interface {
	Stringer
	~int
}

//so:inline
func add[T Number](a, b T) T {
	return a + b
}

//so:inline
func first[T ordered](a, b T) T {
	return a
}

//so:inline
func same[T named](v T) T {
	return v
}

type step int

func (s step) String() string {
	if s == 0 {
		return "zero"
	}
	return "step"
}

func main() {
	{
		// Generic extern function (single type parameter).
		var v *int = newObj[int]()
		*v = 42
		if *v != 42 {
			panic("unexpected value")
		}
		freeObj(v)
	}
	{
		// Generic extern function (multiple type parameters),
		// generic extern type, generic extern method.
		m := newMap[string, int](10)
		if m.Len() != 10 {
			panic("unexpected map size")
		}
	}
	{
		// Generic inline functions with named constraints.
		if add(2, 3) != 5 {
			panic("unexpected sum")
		}
		if first("a", "b") != "a" {
			panic("unexpected value")
		}
		s := same(step(1))
		if s != 1 || s.String() != "step" {
			panic("unexpected step")
		}
	}
	{
		// A constraint declared inside a function body is not emitted.
		type small interface {
			~int8 | ~int16
		}
		if add(1, 1) != 2 {
			panic("unexpected sum")
		}
	}
}
