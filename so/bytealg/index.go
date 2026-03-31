// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

//so:extern nodecay
func IndexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

func IndexByteString(s string, c byte) int {
	return IndexByte([]byte(s), c)
}

func LastIndexByte(s []byte, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func LastIndexByteString(s string, c byte) int {
	return LastIndexByte([]byte(s), c)
}
