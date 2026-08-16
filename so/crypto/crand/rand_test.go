// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crand

import "testing"

func TestTextPanic(t *testing.T) {
	sizes := []int{0, 1, 25}
	for _, size := range sizes {
		checkTextPanic(t, size)
	}
}

// checkTextPanic checks that Text panics on a buffer of the given size.
func checkTextPanic(t *testing.T, size int) {
	defer func() {
		if recover() == nil {
			t.Errorf("Text(make([]byte, %d)) did not panic", size)
		}
	}()
	Text(make([]byte, size))
}
