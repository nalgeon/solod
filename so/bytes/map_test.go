// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	stdbytes "bytes"
	"testing"
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
	// Compare ToLower, ToUpper and Map with the bytes package.
	f.Add([]byte(""))
	f.Add([]byte("abc"))
	f.Add([]byte("AbC123"))
	f.Add([]byte("longStrinGwitHmixofsmaLLandcAps"))
	f.Add([]byte("LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS"))
	f.Add([]byte("ⱭⱭⱭⱭⱭ"))
	f.Add([]byte("a\U0010FFFF"))
	f.Add([]byte("a\xffb"))
	f.Add([]byte("\xed\xa0\x80"))

	f.Fuzz(func(t *testing.T, s []byte) {
		if got, want := ToLower(nil, s), stdbytes.ToLower(s); !Equal(got, want) {
			t.Fatalf("ToLower(%q) = %q, want %q", s, got, want)
		}
		if got, want := ToUpper(nil, s), stdbytes.ToUpper(s); !Equal(got, want) {
			t.Fatalf("ToUpper(%q) = %q, want %q", s, got, want)
		}
		got := Map(nil, rot13, s)
		want := stdbytes.Map(rot13, s)
		if !Equal(got, want) {
			t.Fatalf("Map(rot13, %q) = %q, want %q", s, got, want)
		}
	})
}
