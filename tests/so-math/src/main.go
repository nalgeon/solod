package main

import (
	"github.com/nalgeon/solod/so/math"
)

func main() {
	x := math.Sqrt(49)
	if x != 7 {
		panic("want x == 7")
	}
}
