package main

import (
	"github.com/nalgeon/solod/so/errors"
)

var ErrNotFound = errors.New("not found")

func panicLiteral() {
	panic("something went wrong")
}

func panicString() {
	msg := "runtime error"
	panic(msg)
}

func panicError() {
	err := ErrNotFound
	panic(err)
}

func main() {
	if false {
		panicLiteral()
	}
	if false {
		panicString()
	}
	if false {
		panicError()
	}
}
