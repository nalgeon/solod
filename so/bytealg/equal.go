// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

//so:include <string.h>

// Equal reports whether a and b
// are the same length and contain the same bytes.
// A nil argument is equivalent to an empty slice.
//
// Equal is equivalent to bytes.Equal.
// It is provided here for convenience,
// because some packages cannot depend on bytes.
func Equal(a, b []byte) bool {
	return string(a) == string(b)
}
