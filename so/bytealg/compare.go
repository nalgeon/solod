// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

// Compare returns an integer comparing two byte slices lexicographically.
// The result is 0 if a and b are the same, -1 if a is less than b, and +1 if a
// is greater than b. A nil argument is the same as an empty slice.
//
//so:extern nodecay
func Compare(a, b []byte) int {
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	if l == 0 || &a[0] == &b[0] {
		if len(a) < len(b) {
			return -1
		}
		if len(a) > len(b) {
			return +1
		}
		return 0
	}
	for i := 0; i < l; i++ {
		c1, c2 := a[i], b[i]
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return +1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return +1
	}
	return 0
}
