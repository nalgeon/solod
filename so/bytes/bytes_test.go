// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	stdbytes "bytes"
	"testing"
)

// addSearchSeeds adds the seed corpus of the search fuzzers.
func addSearchSeeds(f *testing.F) {
	f.Add([]byte(""), []byte(""))
	f.Add([]byte("abc"), []byte(""))
	f.Add([]byte(""), []byte("abc"))
	f.Add([]byte("abc"), []byte("b"))
	f.Add([]byte("abcabc"), []byte("abc"))
	f.Add([]byte("banana"), []byte("ana"))
	f.Add([]byte("1,2,3,4"), []byte(","))
	f.Add([]byte("1....2....3....4"), []byte("..."))
	f.Add([]byte("☺☻☹"), []byte("☻"))
	f.Add([]byte("\xff-\xff"), []byte("\xff"))
	f.Add([]byte("oxoxoxoxoxoxoxoxoxoxoxoy"), []byte("oy"))
	f.Add([]byte("barfoobarfooyyyzzzyyyzzzyyyzzzyyyxxxzzzyyy"), []byte("x"))
}

func FuzzSearch(f *testing.F) {
	// Compare the search functions with the bytes package.
	addSearchSeeds(f)

	f.Fuzz(func(t *testing.T, s, sep []byte) {
		if got, want := Index(s, sep), stdbytes.Index(s, sep); got != want {
			t.Fatalf("Index(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if got, want := Count(s, sep), stdbytes.Count(s, sep); got != want {
			t.Fatalf("Count(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if got, want := Contains(s, sep), stdbytes.Contains(s, sep); got != want {
			t.Fatalf("Contains(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := HasPrefix(s, sep), stdbytes.HasPrefix(s, sep); got != want {
			t.Fatalf("HasPrefix(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := HasSuffix(s, sep), stdbytes.HasSuffix(s, sep); got != want {
			t.Fatalf("HasSuffix(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := Equal(s, sep), stdbytes.Equal(s, sep); got != want {
			t.Fatalf("Equal(%q, %q) = %t, want %t", s, sep, got, want)
		}
		got, want := Compare(s, sep), stdbytes.Compare(s, sep)
		if got != want {
			t.Fatalf("Compare(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if len(sep) > 0 {
			got, want := IndexByte(s, sep[0]), stdbytes.IndexByte(s, sep[0])
			if got != want {
				t.Fatalf("IndexByte(%q, %q) = %d, want %d", s, sep[0], got, want)
			}
		}
		cut := Cut(s, sep)
		before, after, found := stdbytes.Cut(s, sep)
		if !Equal(cut.Before, before) || !Equal(cut.After, after) || cut.Found != found {
			t.Fatalf("Cut(%q, %q) = %q, %q, %t; want %q, %q, %t",
				s, sep, cut.Before, cut.After, cut.Found, before, after, found)
		}
	})
}

func FuzzTransform(f *testing.F) {
	// Compare the functions that build a new slice with the bytes package.
	f.Add([]byte(""), []byte(""), []byte(""), 0)
	f.Add([]byte("hello"), []byte("l"), []byte("L"), -1)
	f.Add([]byte("hello"), []byte("l"), []byte("L"), 1)
	f.Add([]byte("hello"), []byte(""), []byte("-"), -1)
	f.Add([]byte("banana"), []byte("a"), []byte(""), 2)
	f.Add([]byte("☺☻☹"), []byte("☻"), []byte("☺"), -1)
	f.Add([]byte("\xff-\xff"), []byte("\xff"), []byte("x"), 3)

	f.Fuzz(func(t *testing.T, s, old, new []byte, n int) {
		got := Replace(nil, s, old, new, n)
		want := stdbytes.Replace(s, old, new, n)
		if !Equal(got, want) {
			t.Fatalf("Replace(%q, %q, %q, %d) = %q, want %q", s, old, new, n, got, want)
		}

		// A large count would need a large allocation, so the count of the
		// Repeat call stays small.
		count := int(uint(n) % 8)
		gotRep := Repeat(nil, s, count)
		wantRep := stdbytes.Repeat(s, count)
		if !Equal(gotRep, wantRep) {
			t.Fatalf("Repeat(%q, %d) = %q, want %q", s, count, gotRep, wantRep)
		}

		gotRunes := Runes(nil, s)
		wantRunes := stdbytes.Runes(s)
		if len(gotRunes) != len(wantRunes) {
			t.Fatalf("Runes(%q) = %v, want %v", s, gotRunes, wantRunes)
		}
		for i, r := range gotRunes {
			if r != wantRunes[i] {
				t.Fatalf("Runes(%q)[%d] = %U, want %U", s, i, r, wantRunes[i])
			}
		}

		if got := String(nil, s); got != string(s) {
			t.Fatalf("String(%q) = %q, want %q", s, got, string(s))
		}
		if got := Clone(nil, s); !Equal(got, s) || (len(s) == 0) != (len(got) == 0) {
			t.Fatalf("Clone(%q) = %q", s, got)
		}
	})
}
