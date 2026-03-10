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

func main() {
	var v *int = newObj[int]()
	*v = 42
	if *v != 42 {
		panic("unexpected value")
	}
	freeObj(v)
}
