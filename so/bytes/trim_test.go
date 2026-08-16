// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	stdbytes "bytes"
	"testing"
	stdunicode "unicode"

	"solod.dev/so/unicode"
)

func FuzzTrim(f *testing.F) {
	// Compare the trim functions with the bytes package.
	f.Add([]byte(""), "")
	f.Add([]byte("abba"), "a")
	f.Add([]byte("abba"), "ab")
	f.Add([]byte("<tag>"), "<>")
	f.Add([]byte("ⱯⱯɐɐⱯⱯ"), "Ɐ")
	f.Add([]byte("\x80test\xff"), "\xff")
	f.Add([]byte("☺\xc0"), "☺")
	f.Add([]byte("\t\v\r\f\n  hello \t\v\r\f\n "), " ")
	f.Add([]byte("x \xc0\xc0 "), "　")

	f.Fuzz(func(t *testing.T, s []byte, cutset string) {
		if got, want := Trim(s, cutset), stdbytes.Trim(s, cutset); !Equal(got, want) {
			t.Fatalf("Trim(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimLeft(s, cutset), stdbytes.TrimLeft(s, cutset); !Equal(got, want) {
			t.Fatalf("TrimLeft(%q, %q) = %q, want %q", s, cutset, got, want)
		}
		if got, want := TrimRight(s, cutset), stdbytes.TrimRight(s, cutset); !Equal(got, want) {
			t.Fatalf("TrimRight(%q, %q) = %q, want %q", s, cutset, got, want)
		}

		cut := []byte(cutset)
		if got, want := TrimPrefix(s, cut), stdbytes.TrimPrefix(s, cut); !Equal(got, want) {
			t.Fatalf("TrimPrefix(%q, %q) = %q, want %q", s, cut, got, want)
		}
		if got, want := TrimSuffix(s, cut), stdbytes.TrimSuffix(s, cut); !Equal(got, want) {
			t.Fatalf("TrimSuffix(%q, %q) = %q, want %q", s, cut, got, want)
		}

		if got, want := TrimSpace(s), stdbytes.TrimSpace(s); !Equal(got, want) {
			t.Fatalf("TrimSpace(%q) = %q, want %q", s, got, want)
		}
		got := TrimFunc(s, unicode.IsSpace)
		want := stdbytes.TrimFunc(s, stdunicode.IsSpace)
		if !Equal(got, want) {
			t.Fatalf("TrimFunc(%q, IsSpace) = %q, want %q", s, got, want)
		}
	})
}
