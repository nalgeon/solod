// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmp_test

import (
	"solod.dev/so/cmp"
	"solod.dev/so/fmt"
)

func ExampleLess() {
	fmt.Printf("%t\n", cmp.Less(1, 2))
	fmt.Printf("%t\n", cmp.Less("a", "aa"))
	// Output:
	// true
	// true
}

func ExampleCompare() {
	fmt.Printf("%d\n", cmp.Compare(1, 2))
	fmt.Printf("%d\n", cmp.Compare("a", "aa"))
	fmt.Printf("%d\n", cmp.Compare(1.5, 1.5))
	// Output:
	// -1
	// -1
	// 0
}

func ExampleEqual() {
	fmt.Printf("%t\n", cmp.Equal(1, 1))
	fmt.Printf("%t\n", cmp.Equal("a", "aa"))
	fmt.Printf("%t\n", cmp.Equal(1.5, 1.5))
	// Output:
	// true
	// false
	// true
}
