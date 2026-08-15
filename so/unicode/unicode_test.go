// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode_test

import (
	"testing"

	. "solod.dev/so/unicode"
)

func FuzzRune(f *testing.F) {
	// Checks the invariants of the package over arbitrary runes,
	// which include negative runes and runes above MaxRune.
	f.Add(int32(0))
	f.Add(int32('a'))
	f.Add(int32(MaxASCII))
	f.Add(int32(MaxLatin1))
	f.Add(int32(MaxRune))
	f.Add(int32(MaxRune + 1))
	f.Add(int32(-1))
	f.Add(int32(1 << 30))

	f.Fuzz(func(t *testing.T, v int32) {
		r := rune(v)

		// The fast path of each Is function must agree with its table.
		// The So test checks this over Latin-1 only.
		checkIs(t, r, "Letter", IsLetter(r), Is(Letter, r))
		checkIs(t, r, "Upper", IsUpper(r), Is(Upper, r))
		checkIs(t, r, "Lower", IsLower(r), Is(Lower, r))
		checkIs(t, r, "Title", IsTitle(r), Is(Title, r))
		checkIs(t, r, "Space", IsSpace(r), Is(White_Space, r))
		checkIs(t, r, "Digit", IsDigit(r), Is(Digit, r))

		// The fast path of each To function must agree with To.
		checkTo(t, r, "ToUpper", ToUpper(r), To(UpperCase, r))
		checkTo(t, r, "ToLower", ToLower(r), To(LowerCase, r))
		checkTo(t, r, "ToTitle", ToTitle(r), To(TitleCase, r))

		// A rune the package reports as cased must map to itself in that case.
		if IsUpper(r) && ToUpper(r) != r {
			t.Errorf("IsUpper(%U) = true, but ToUpper(%U) = %U", r, r, ToUpper(r))
		}
		if IsLower(r) && ToLower(r) != r {
			t.Errorf("IsLower(%U) = true, but ToLower(%U) = %U", r, r, ToLower(r))
		}

		// No mapping may leave the Unicode range for a valid rune.
		if r >= 0 && r <= MaxRune {
			if got := ToUpper(r); got < 0 || got > MaxRune {
				t.Errorf("ToUpper(%U) = %U, outside the Unicode range", r, got)
			}
			if got := ToLower(r); got < 0 || got > MaxRune {
				t.Errorf("ToLower(%U) = %U, outside the Unicode range", r, got)
			}
		}

		// An invalid case index must give the replacement character.
		if got := To(MaxCase, r); got != ReplacementChar {
			t.Errorf("To(MaxCase, %U) = %U, want %U", r, got, ReplacementChar)
		}
	})
}

func checkIs(t *testing.T, r rune, name string, fast, table bool) {
	t.Helper()
	if fast != table {
		t.Errorf("Is%s(%U) = %t, but Is(%s, %U) = %t", name, r, fast, name, r, table)
	}
}

func checkTo(t *testing.T, r rune, name string, fast, general rune) {
	t.Helper()
	if fast != general {
		t.Errorf("%s(%U) = %U, but To(%U) = %U", name, r, fast, r, general)
	}
}
