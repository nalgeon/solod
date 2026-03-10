package main

type Sum3Fn func(int, int, int) int
type sum3Fn func(int, int, int) int

func sum3(a, b, c int) int {
	return a + b + c
}

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
}
