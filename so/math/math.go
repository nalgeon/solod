// Package math provides basic constants and mathematical functions.
//
// CAUTION: This package is under development, do not use yet.
package math

//so:include <math.h>

func Sqrt(x float64) float64 {
	return sqrt(x)
}

func Floor(x float64) float64 {
	return floor(x)
}

func Ceil(x float64) float64 {
	return ceil(x)
}

//so:extern
func sqrt(x float64) float64 { return x }

//so:extern
func floor(x float64) float64 { return x }

//so:extern
func ceil(x float64) float64 { return x }
