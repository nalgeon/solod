// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

// Implementations to avoid importing other dependencies.

// package math

// floatAbs returns the absolute value of x.
// floatAbs returns NaN if x is NaN.
// The name avoids a collision with abs from the C standard library.
func floatAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
