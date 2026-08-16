// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	stdstrings "strings"
	"testing"
	"unicode/utf8"
)

// addSearchSeeds adds the seed corpus of the search fuzzers.
func addSearchSeeds(f *testing.F) {
	f.Add("", "")
	f.Add("abc", "")
	f.Add("", "abc")
	f.Add("abc", "b")
	f.Add("abcabc", "abc")
	f.Add("banana", "ana")
	f.Add("1,2,3,4", ",")
	f.Add("1....2....3....4", "...")
	f.Add("☺☻☹", "☻")
	f.Add("\xff-\xff", "\xff")
	f.Add("oxoxoxoxoxoxoxoxoxoxoxoy", "oy")
	f.Add("barfoobarfooyyyzzzyyyzzzyyyzzzyyyxxxzzzyyy", "x")
}

func FuzzSearch(f *testing.F) {
	// Compare the search functions with the strings package.
	addSearchSeeds(f)

	f.Fuzz(func(t *testing.T, s, sep string) {
		if got, want := Index(s, sep), stdstrings.Index(s, sep); got != want {
			t.Fatalf("Index(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if got, want := LastIndex(s, sep), stdstrings.LastIndex(s, sep); got != want {
			t.Fatalf("LastIndex(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if got, want := Count(s, sep), stdstrings.Count(s, sep); got != want {
			t.Fatalf("Count(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if got, want := Contains(s, sep), stdstrings.Contains(s, sep); got != want {
			t.Fatalf("Contains(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := ContainsAny(s, sep), stdstrings.ContainsAny(s, sep); got != want {
			t.Fatalf("ContainsAny(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := IndexAny(s, sep), stdstrings.IndexAny(s, sep); got != want {
			t.Fatalf("IndexAny(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if got, want := HasPrefix(s, sep), stdstrings.HasPrefix(s, sep); got != want {
			t.Fatalf("HasPrefix(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := HasSuffix(s, sep), stdstrings.HasSuffix(s, sep); got != want {
			t.Fatalf("HasSuffix(%q, %q) = %t, want %t", s, sep, got, want)
		}
		if got, want := Compare(s, sep), stdstrings.Compare(s, sep); got != want {
			t.Fatalf("Compare(%q, %q) = %d, want %d", s, sep, got, want)
		}
		if len(sep) > 0 {
			if got, want := IndexByte(s, sep[0]), stdstrings.IndexByte(s, sep[0]); got != want {
				t.Fatalf("IndexByte(%q, %q) = %d, want %d", s, sep[0], got, want)
			}
			got, want := LastIndexByte(s, sep[0]), stdstrings.LastIndexByte(s, sep[0])
			if got != want {
				t.Fatalf("LastIndexByte(%q, %q) = %d, want %d", s, sep[0], got, want)
			}
			r, _ := utf8.DecodeRuneInString(sep)
			if got, want := IndexRune(s, r), stdstrings.IndexRune(s, r); got != want {
				t.Fatalf("IndexRune(%q, %q) = %d, want %d", s, r, got, want)
			}
			if got, want := ContainsRune(s, r), stdstrings.ContainsRune(s, r); got != want {
				t.Fatalf("ContainsRune(%q, %q) = %t, want %t", s, r, got, want)
			}
		}

		gotBefore, gotAfter := Cut(s, sep)
		before, after, _ := stdstrings.Cut(s, sep)
		if gotBefore != before || gotAfter != after {
			t.Fatalf("Cut(%q, %q) = %q, %q; want %q, %q",
				s, sep, gotBefore, gotAfter, before, after)
		}
		gotCut, gotFound := CutPrefix(s, sep)
		wantCut, wantFound := stdstrings.CutPrefix(s, sep)
		if gotCut != wantCut || gotFound != wantFound {
			t.Fatalf("CutPrefix(%q, %q) = %q, %t; want %q, %t",
				s, sep, gotCut, gotFound, wantCut, wantFound)
		}
		gotCut, gotFound = CutSuffix(s, sep)
		wantCut, wantFound = stdstrings.CutSuffix(s, sep)
		if gotCut != wantCut || gotFound != wantFound {
			t.Fatalf("CutSuffix(%q, %q) = %q, %t; want %q, %t",
				s, sep, gotCut, gotFound, wantCut, wantFound)
		}
	})
}

func FuzzTransform(f *testing.F) {
	// Compare the functions that build a new string with the strings package.
	f.Add("", "", "", 0)
	f.Add("hello", "l", "L", -1)
	f.Add("hello", "l", "L", 1)
	f.Add("hello", "", "-", -1)
	f.Add("banana", "a", "", 2)
	f.Add("☺☻☹", "☻", "☺", -1)
	f.Add("\xff-\xff", "\xff", "x", 3)

	f.Fuzz(func(t *testing.T, s, old, new string, n int) {
		got := Replace(nil, s, old, new, n)
		want := stdstrings.Replace(s, old, new, n)
		if got != want {
			t.Fatalf("Replace(%q, %q, %q, %d) = %q, want %q", s, old, new, n, got, want)
		}

		gotAll := ReplaceAll(nil, s, old, new)
		wantAll := stdstrings.ReplaceAll(s, old, new)
		if gotAll != wantAll {
			t.Fatalf("ReplaceAll(%q, %q, %q) = %q, want %q", s, old, new, gotAll, wantAll)
		}

		// A large count would need a large allocation, so the count of the
		// Repeat call stays small.
		count := int(uint(n) % 8)
		gotRep := Repeat(nil, s, count)
		wantRep := stdstrings.Repeat(s, count)
		if gotRep != wantRep {
			t.Fatalf("Repeat(%q, %d) = %q, want %q", s, count, gotRep, wantRep)
		}

		if got := Clone(nil, s); got != s {
			t.Fatalf("Clone(%q) = %q", s, got)
		}
	})
}
