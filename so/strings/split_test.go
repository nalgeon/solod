// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	stdstrings "strings"
	"testing"
)

func FuzzSplit(f *testing.F) {
	// Compare Split, SplitN, SplitAfter and Join with the strings package.
	f.Add("", "", -1)
	f.Add("abcd", "", 2)
	f.Add("abcd", "a", 0)
	f.Add("1,2,3,4", ",", -1)
	f.Add("1 2 3 4", " ", 3)
	f.Add("1....2....3....4", "...", -1)
	f.Add("☺☻☹", "☹", -1)
	f.Add("\xff-\xff", "", -1)

	f.Fuzz(func(t *testing.T, s, sep string, n int) {
		got := SplitN(nil, s, sep, n)
		want := stdstrings.SplitN(s, sep, n)
		if len(got) != len(want) {
			t.Fatalf("SplitN(%q, %q, %d) = %q, want %q", s, sep, n, got, want)
		}
		for i, p := range got {
			if p != want[i] {
				t.Fatalf("SplitN(%q, %q, %d)[%d] = %q, want %q", s, sep, n, i, p, want[i])
			}
		}

		parts := Split(nil, s, sep)
		wantAll := stdstrings.Split(s, sep)
		if len(parts) != len(wantAll) {
			t.Fatalf("Split(%q, %q) = %q, want %q", s, sep, parts, wantAll)
		}
		for i, p := range parts {
			if p != wantAll[i] {
				t.Fatalf("Split(%q, %q)[%d] = %q, want %q", s, sep, i, p, wantAll[i])
			}
		}

		after := SplitAfter(nil, s, sep)
		wantAfter := stdstrings.SplitAfter(s, sep)
		if len(after) != len(wantAfter) {
			t.Fatalf("SplitAfter(%q, %q) = %q, want %q", s, sep, after, wantAfter)
		}
		for i, p := range after {
			if p != wantAfter[i] {
				t.Fatalf("SplitAfter(%q, %q)[%d] = %q, want %q", s, sep, i, p, wantAfter[i])
			}
		}

		joined := Join(nil, parts, sep)
		wantJoined := stdstrings.Join(wantAll, sep)
		if joined != wantJoined {
			t.Fatalf("Join(Split(%q, %q), %q) = %q, want %q", s, sep, sep, joined, wantJoined)
		}

		fields := Fields(nil, s)
		wantFields := stdstrings.Fields(s)
		if len(fields) != len(wantFields) {
			t.Fatalf("Fields(%q) = %q, want %q", s, fields, wantFields)
		}
		for i, p := range fields {
			if p != wantFields[i] {
				t.Fatalf("Fields(%q)[%d] = %q, want %q", s, i, p, wantFields[i])
			}
		}
	})
}
