// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

// Implementations to avoid importing other dependencies.

// package math

// isFinite reports whether f is neither NaN nor an infinity.
// f-f is NaN if f is NaN or an infinity. f-f is 0 in every other case.
func isFinite(f float64) bool {
	d := f - f
	return d == d
}
