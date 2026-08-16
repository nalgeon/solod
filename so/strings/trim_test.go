// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	stdstrings "strings"
	"testing"
	stdunicode "unicode"

	"solod.dev/so/unicode"
)

func FuzzTrim(f *testing.F) {
	// Compare the trim functions with the strings package.
	f.Add("", "")
	f.Add("abba", "a")
	f.Add("abba", "ab")
	f.Add("<tag>", "<>")
	f.Add("ⱯⱯɐɐⱯⱯ", "Ɐ")
	f.Add("\x80test\xff", "\xff")
	f.Add("☺\xc0", "☺")
	f.Add("\t\v\r\f\n hello \t\v\r\f\n", " ")
	f.Add("x \xc0\xc0 ", "　")

	f.Fuzz(func(t *testing.T, s, cutset string) {
		if got, want := Trim(s, cutset), stdstrings.Trim(s, cutset); got != want {
			t.Fatalf("Trim(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimLeft(s, cutset), stdstrings.TrimLeft(s, cutset); got != want {
			t.Fatalf("TrimLeft(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimRight(s, cutset), stdstrings.TrimRight(s, cutset); got != want {
			t.Fatalf("TrimRight(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimPrefix(s, cutset), stdstrings.TrimPrefix(s, cutset); got != want {
			t.Fatalf("TrimPrefix(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimSuffix(s, cutset), stdstrings.TrimSuffix(s, cutset); got != want {
			t.Fatalf("TrimSuffix(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimSpace(s), stdstrings.TrimSpace(s); got != want {
			t.Fatalf("TrimSpace(%q) = %q, want %q", s, got, want)
		}

		got := TrimFunc(s, unicode.IsSpace)
		want := stdstrings.TrimFunc(s, stdunicode.IsSpace)
		if got != want {
			t.Fatalf("TrimFunc(%q, IsSpace) = %q, want %q", s, got, want)
		}
		gotIdx := IndexFunc(s, unicode.IsSpace)
		wantIdx := stdstrings.IndexFunc(s, stdunicode.IsSpace)
		if gotIdx != wantIdx {
			t.Fatalf("IndexFunc(%q, IsSpace) = %d, want %d", s, gotIdx, wantIdx)
		}
		gotHas := ContainsFunc(s, unicode.IsSpace)
		wantHas := stdstrings.ContainsFunc(s, stdunicode.IsSpace)
		if gotHas != wantHas {
			t.Fatalf("ContainsFunc(%q, IsSpace) = %t, want %t", s, gotHas, wantHas)
		}
	})
}
