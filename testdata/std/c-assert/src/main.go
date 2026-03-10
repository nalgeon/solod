package main

import "github.com/nalgeon/solod/so/c/assert"

func main() {
	if assert.Enabled {
		assert.Assert(1+1 == 2)
		assert.Assertf(1+1 == 2, "math is broken")
	} else {
		println("assertions are disabled")
	}
}
