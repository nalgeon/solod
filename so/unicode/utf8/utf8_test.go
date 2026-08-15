// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utf8

import "testing"

// FuzzDecode checks the invariants of the decoders over arbitrary bytes.
func FuzzDecode(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("abcd"))
	f.Add([]byte("☺☻☹"))
	f.Add([]byte("日a本b語ç"))
	f.Add([]byte("\x80\x80\x80\x80"))
	f.Add([]byte("\xed\xa0\x80"))     // a surrogate half
	f.Add([]byte("\xc0\x80"))         // U+0000 encoded in two bytes
	f.Add([]byte("\xf4\x90\x80\x80")) // above MaxRune
	f.Add([]byte("\xf4\x8f\xbf\xbf")) // MaxRune

	f.Fuzz(func(t *testing.T, b []byte) {
		s := string(b)

		// The byte and the string decoder must agree.
		r1, size1 := DecodeRune(b)
		r2, size2 := DecodeRuneInString(s)
		if r1 != r2 || size1 != size2 {
			t.Fatalf("DecodeRune(%q) = %U, %d; DecodeRuneInString = %U, %d", b, r1, size1, r2, size2)
		}
		r3, last1 := DecodeLastRune(b)
		r4, last2 := DecodeLastRuneInString(s)
		if r3 != r4 || last1 != last2 {
			t.Fatalf("DecodeLastRune(%q) = %U, %d; DecodeLastRuneInString = %U, %d", b, r3, last1, r4, last2)
		}
		if FullRune(b) != FullRuneInString(s) {
			t.Fatalf("FullRune(%q) = %t; FullRuneInString = %t", b, FullRune(b), FullRuneInString(s))
		}
		if Valid(b) != ValidString(s) {
			t.Fatalf("Valid(%q) = %t; ValidString = %t", b, Valid(b), ValidString(s))
		}

		// A forward walk must reach the end, count the runes that
		// RuneCount counts, and report an error exactly when Valid does.
		count, ok := 0, true
		for i := 0; i < len(b); {
			r, size := DecodeRune(b[i:])
			if size == 0 {
				t.Fatalf("DecodeRune(%q) returned size 0 at %d", b, i)
			}
			if r == RuneError && size == 1 {
				ok = false
			} else if !ValidRune(r) {
				t.Fatalf("DecodeRune(%q) at %d = %U, not a valid rune", b, i, r)
			} else if n := RuneLen(r); n != size {
				t.Fatalf("DecodeRune(%q) at %d = %U, %d; RuneLen = %d", b, i, r, size, n)
			}
			i += size
			count++
		}
		if got := RuneCount(b); got != count {
			t.Fatalf("RuneCount(%q) = %d; the walk counted %d", b, got, count)
		}
		if got := RuneCountInString(s); got != count {
			t.Fatalf("RuneCountInString(%q) = %d; the walk counted %d", b, got, count)
		}
		if Valid(b) != ok {
			t.Fatalf("Valid(%q) = %t; the walk says %t", b, Valid(b), ok)
		}

		// A backward walk must take the same number of steps.
		back := 0
		for i := len(b); i > 0; back++ {
			_, size := DecodeLastRune(b[:i])
			if size == 0 {
				t.Fatalf("DecodeLastRune(%q) returned size 0 at %d", b, i)
			}
			i -= size
		}
		if back != count {
			t.Fatalf("the backward walk over %q took %d steps; the forward walk took %d", b, back, count)
		}
	})
}

// FuzzEncode checks that a valid rune survives an encode and a decode,
// and that an invalid rune encodes as RuneError.
func FuzzEncode(f *testing.F) {
	f.Add(int32(0))
	f.Add(int32('a'))
	f.Add(int32('☺'))
	f.Add(int32(RuneError))
	f.Add(int32(MaxRune))
	f.Add(int32(MaxRune + 1))
	f.Add(int32(0xD800))
	f.Add(int32(-1))

	f.Fuzz(func(t *testing.T, v int32) {
		r := rune(v)

		buf := make([]byte, UTFMax)
		n := EncodeRune(buf, r)
		buf = buf[:n]

		if !ValidRune(r) {
			// An invalid rune must encode as RuneError.
			if string(buf) != string(RuneError) {
				t.Fatalf("EncodeRune(%U) = %q, want the encoding of RuneError", r, buf)
			}
			if RuneLen(r) != -1 {
				t.Fatalf("RuneLen(%U) = %d, want -1", r, RuneLen(r))
			}
			return
		}

		if n != RuneLen(r) {
			t.Fatalf("EncodeRune(%U) wrote %d bytes; RuneLen = %d", r, n, RuneLen(r))
		}
		if !Valid(buf) {
			t.Fatalf("EncodeRune(%U) = %q, which Valid rejects", r, buf)
		}
		if got, size := DecodeRune(buf); got != r || size != n {
			t.Fatalf("DecodeRune(EncodeRune(%U)) = %U, %d, want %U, %d", r, got, size, r, n)
		}
		if got, size := DecodeLastRune(buf); got != r || size != n {
			t.Fatalf("DecodeLastRune(EncodeRune(%U)) = %U, %d, want %U, %d", r, got, size, r, n)
		}

		// AppendRune must write the same bytes as EncodeRune.
		app := AppendRune(make([]byte, 0, UTFMax), r)
		if string(app) != string(buf) {
			t.Fatalf("AppendRune(%U) = %q, EncodeRune = %q", r, app, buf)
		}
	})
}
