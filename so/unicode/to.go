// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode

// ToUpper maps the rune to upper case.
func ToUpper(r rune) rune {
	if r <= MaxASCII {
		if 'a' <= uint32(r) && uint32(r) <= 'z' {
			r -= 'a' - 'A'
		}
		return r
	}
	return To(UpperCase, r)
}

// ToLower maps the rune to lower case.
func ToLower(r rune) rune {
	if r <= MaxASCII {
		if 'A' <= uint32(r) && uint32(r) <= 'Z' {
			r += 'a' - 'A'
		}
		return r
	}
	return To(LowerCase, r)
}

// ToTitle maps the rune to title case.
func ToTitle(r rune) rune {
	if r <= MaxASCII {
		if 'a' <= uint32(r) && uint32(r) <= 'z' { // title case is upper case for ASCII
			r -= 'a' - 'A'
		}
		return r
	}
	return To(TitleCase, r)
}

// To maps the rune to the specified case: [UpperCase], [LowerCase], or [TitleCase].
func To(_case int, r rune) rune {
	r, _ = to(_case, r, CaseRanges)
	return r
}

// to maps the rune using the specified case mapping.
// It additionally reports whether caseRange contained a mapping for r.
func to(_case int, r rune, caseRange []CaseRange) (rune, bool) {
	if _case < 0 || MaxCase <= _case {
		return ReplacementChar, false // as reasonable an error as any
	}
	if cr := lookupCaseRange(r, caseRange); cr != nil {
		return convertCase(_case, r, cr), true
	}
	return r, false
}

// lookupCaseRange returns the CaseRange mapping for rune r or nil if no
// mapping exists for r.
func lookupCaseRange(r rune, caseRange []CaseRange) *CaseRange {
	// binary search over ranges
	lo := 0
	hi := len(caseRange)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		cr := &caseRange[m]
		if rune(cr.Lo) <= r && r <= rune(cr.Hi) {
			return cr
		}
		if r < rune(cr.Lo) {
			hi = m
		} else {
			lo = m + 1
		}
	}
	return nil
}

// convertCase converts r to _case using CaseRange cr.
func convertCase(_case int, r rune, cr *CaseRange) rune {
	delta := cr.Delta[_case]
	if delta > MaxRune {
		// In an Upper-Lower sequence, which always starts with
		// an UpperCase letter, the real deltas always look like:
		//	{0, 1, 0}    UpperCase (Lower is next)
		//	{-1, 0, -1}  LowerCase (Upper, Title are previous)
		// The characters at even offsets from the beginning of the
		// sequence are upper case; the ones at odd offsets are lower.
		// The correct mapping can be done by clearing or setting the low
		// bit in the sequence offset.
		// The constants UpperCase and TitleCase are even while LowerCase
		// is odd so we take the low bit from _case.
		return rune(cr.Lo) + (((r - rune(cr.Lo)) &^ 1) | rune(_case&1))
	}
	return r + delta
}
