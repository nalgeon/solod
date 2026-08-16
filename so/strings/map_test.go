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

// rot13 moves a letter 13 places along the alphabet. rot13 is its own inverse.
func rot13(r rune) rune {
	const step = 13
	if r >= 'a' && r <= 'z' {
		return ((r - 'a' + step) % 26) + 'a'
	}
	if r >= 'A' && r <= 'Z' {
		return ((r - 'A' + step) % 26) + 'A'
	}
	return r
}

func FuzzCase(f *testing.F) {
	// Compare ToLower, ToUpper, Map and FieldsFunc with the strings package.
	f.Add("")
	f.Add("abc")
	f.Add("AbC123")
	f.Add("longStrinGwitHmixofsmaLLandcAps")
	f.Add("LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS")
	f.Add("ⱭⱭⱭⱭⱭ")
	f.Add("a\U0010FFFF")
	f.Add("a\xffb")
	f.Add("\xed\xa0\x80")

	f.Fuzz(func(t *testing.T, s string) {
		if got, want := ToLower(nil, s), stdstrings.ToLower(s); got != want {
			t.Fatalf("ToLower(%q) = %q, want %q", s, got, want)
		}
		if got, want := ToUpper(nil, s), stdstrings.ToUpper(s); got != want {
			t.Fatalf("ToUpper(%q) = %q, want %q", s, got, want)
		}
		if got, want := Map(nil, rot13, s), stdstrings.Map(rot13, s); got != want {
			t.Fatalf("Map(rot13, %q) = %q, want %q", s, got, want)
		}

		got := FieldsFunc(nil, s, unicode.IsSpace)
		want := stdstrings.FieldsFunc(s, stdunicode.IsSpace)
		if len(got) != len(want) {
			t.Fatalf("FieldsFunc(%q, IsSpace) = %q, want %q", s, got, want)
		}
		for i, p := range got {
			if p != want[i] {
				t.Fatalf("FieldsFunc(%q, IsSpace)[%d] = %q, want %q", s, i, p, want[i])
			}
		}
	})
}
