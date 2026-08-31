package main

import "example/functions/src/sub"

type Sum3Fn func(int, int, int) int
type sum3Fn func(int, int, int) int

func sum3(a, b, c int) int {
	return a + b + c
}

// Blank parameters.
func pickThird(_ int, _ float32, c int) int {
	return c
}

// Unnamed parameters.
func countUp(int, float32) int {
	return 1
}

// stop guards an early return in main.
var stop = false

func main() {
	s0 := sum3(1, 2, 3)
	_ = s0

	fn1 := sum3
	s1 := fn1(4, 5, 6)
	_ = s1

	var fn2 Sum3Fn = sum3
	s2 := fn2(7, 8, 9)
	_ = s2

	var fn3 sum3Fn = sum3
	s3 := fn3(3, 3, 3)
	_ = s3

	// Function literals (anonymous functions) are not supported.
	// fn4 := func(a, b, c int) int {
	// 	return a * b * c
	// }
	// s4 := fn4(2, 3, 4)
	// _ = s4

	var fn5 Sum3Fn = sub.Sum
	s5 := fn5(10, 20, 30)
	_ = s5

	s6 := pickThird(1, 2, 3)
	_ = s6

	s7 := countUp(1, 2)
	_ = s7

	// A bare return in main translate to return 0 in C.
	if stop {
		return
	}
	println("done")
	return
}
