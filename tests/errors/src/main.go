package main

import (
	"errors"
)

var ErrOutOfTea = errors.New("no more tea available")

func makeTea(arg int) error {
	if arg == 42 {
		return ErrOutOfTea
	}
	return nil
}

func main() {
	// Nil and non-nil errors.
	err := makeTea(7)
	if err != nil {
		panic("err != nil")
	}

	err = makeTea(42)
	if err == nil {
		panic("err == nil")
	}
	if err != ErrOutOfTea {
		panic("err != ErrOutOfTea")
	}

	// Not supported: errors can only be defined at package level.
	// errNotSupported := errors.New("operation not supported")

	// Dynamic errors are also not supported.
	// errNotSupported := fmt.Errorf("not supported: %d", 42)
}
