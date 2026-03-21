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
}
