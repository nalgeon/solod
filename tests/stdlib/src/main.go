package main

import (
	"github.com/nalgeon/solod/so"
	"github.com/nalgeon/solod/so/errors"
	"github.com/nalgeon/solod/so/math"
)

var Err42 = errors.New("42")

func work(n int) so.ResInt {
	if n == 42 {
		return so.ResInt{Err: Err42}
	}
	return so.ResInt{Val: 42}
}

func main() {
	x := math.Sqrt(4.0)
	_ = x

	r1 := work(11)
	if r1.Err != nil {
		panic("unexpected error")
	}

	r2 := work(42)
	if r2.Err != Err42 {
		panic("expected Err42")
	}
}
