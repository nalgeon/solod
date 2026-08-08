// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices_test

import (
	"testing"

	"solod.dev/so/c"
	. "solod.dev/so/slices"
)

func TestMinMaxPanics(t *testing.T) {
	intCmp := func(a, b any) int {
		va := *c.PtrAs[int](a)
		vb := *c.PtrAs[int](b)
		return va - vb
	}
	emptySlice := []int{}

	if !panics(func() { _ = Min(emptySlice) }) {
		t.Errorf("Min([]): got no panic, want panic")
	}

	if !panics(func() { _ = Max(emptySlice) }) {
		t.Errorf("Max([]): got no panic, want panic")
	}

	if !panics(func() { _ = MinFunc(emptySlice, intCmp) }) {
		t.Errorf("MinFunc([]): got no panic, want panic")
	}

	if !panics(func() { _ = MaxFunc(emptySlice, intCmp) }) {
		t.Errorf("MaxFunc([]): got no panic, want panic")
	}
}

func panics(f func()) (b bool) {
	defer func() {
		if x := recover(); x != nil {
			b = true
		}
	}()
	f()
	return false
}
