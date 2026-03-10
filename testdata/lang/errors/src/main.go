package main

import (
	"github.com/nalgeon/solod/so/errors"
)

var ErrOutOfTea = errors.New("no more tea available")

func makeTea(arg int) error {
	if arg == 42 {
		return ErrOutOfTea
	}
	return nil
}

func work(n int) (int, error) {
	if n == 42 {
		return 0, ErrOutOfTea
	}
	return n, nil
}

func main() {
	{
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
	}
	{
		// Variable of type error.
		var err error
		if err != nil {
			panic("err != nil")
		}
		err = makeTea(42)
		if err == nil {
			panic("err == nil")
		}
	}
	{
		// Multiple returns with error.
		r1, err := work(11)
		if r1 != 11 {
			panic("unexpected result")
		}
		if err != nil {
			panic("unexpected error")
		}
		_ = r1

		r2, err := work(42)
		if r2 != 0 {
			panic("unexpected result")
		}
		if err != ErrOutOfTea {
			panic("expected ErrOutOfTea")
		}
		_ = r2
	}
	{
		// Printing errors.
		err := makeTea(42)
		println("err =", err)
	}

	// Not supported: errors can only be defined at package level.
	// errNotSupported := errors.New("operation not supported")

	// Dynamic errors are also not supported.
	// errNotSupported := fmt.Errorf("not supported: %d", 42)
}
