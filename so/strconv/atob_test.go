// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

import (
	stdconv "strconv"
	"testing"
)

func FuzzParseBool(f *testing.F) {
	// Compare ParseBool with the strconv package.
	f.Add("")
	f.Add("0")
	f.Add("1")
	f.Add("t")
	f.Add("T")
	f.Add("TRUE")
	f.Add("true")
	f.Add("True")
	f.Add("f")
	f.Add("F")
	f.Add("FALSE")
	f.Add("false")
	f.Add("False")
	f.Add("asdf")
	f.Add("tRuE")

	f.Fuzz(func(t *testing.T, s string) {
		got, gotErr := ParseBool(s)
		want, wantErr := stdconv.ParseBool(s)
		if got != want || errKind(gotErr) != errKind(wantErr) {
			t.Fatalf("ParseBool(%q) = %t, %v; want %t, %v", s, got, gotErr, want, wantErr)
		}
	})
}
