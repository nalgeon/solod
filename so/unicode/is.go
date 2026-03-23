// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode

// In reports whether the rune is a member of one of the ranges.
func In(r rune, ranges ...*RangeTable) bool {
	for _, inside := range ranges {
		if Is(inside, r) {
			return true
		}
	}
	return false
}

// Is reports whether the rune is in the specified table of ranges.
func Is(rangeTab *RangeTable, r rune) bool {
	r16 := rangeTab.R16
	// Compare as uint32 to correctly handle negative runes.
	if len(r16) > 0 && uint32(r) <= uint32(r16[len(r16)-1].Hi) {
		return is16(r16, uint16(r))
	}
	r32 := rangeTab.R32
	if len(r32) > 0 && r >= rune(r32[0].Lo) {
		return is32(r32, uint32(r))
	}
	return false
}

// IsUpper reports whether the rune is an upper case letter.
func IsUpper(r rune) bool {
	// We convert to uint32 to avoid the extra test for negative,
	// and in the index we convert to uint8 to avoid the range check.
	if uint32(r) <= uint32(MaxLatin1) {
		return (properties[uint8(r)] & pLmask) == pLu
	}
	return isExcludingLatin(Upper, r)
}

// IsLower reports whether the rune is a lower case letter.
func IsLower(r rune) bool {
	// We convert to uint32 to avoid the extra test for negative,
	// and in the index we convert to uint8 to avoid the range check.
	if uint32(r) <= uint32(MaxLatin1) {
		return (properties[uint8(r)] & pLmask) == pLl
	}
	return isExcludingLatin(Lower, r)
}

// IsTitle reports whether the rune is a title case letter.
func IsTitle(r rune) bool {
	if r <= MaxLatin1 {
		return false
	}
	return isExcludingLatin(Title, r)
}

// IsControl reports whether the rune is a control character.
func IsControl(r rune) bool {
	if uint32(r) <= uint32(MaxLatin1) {
		return (properties[uint8(r)] & pC) != 0
	}
	// All control characters are < MaxLatin1.
	return false
}

// IsLetter reports whether the rune is a letter (category [L]).
func IsLetter(r rune) bool {
	if uint32(r) <= uint32(MaxLatin1) {
		return (properties[uint8(r)] & pLmask) != 0
	}
	return isExcludingLatin(Letter, r)
}

// IsDigit reports whether the rune is a decimal digit.
func IsDigit(r rune) bool {
	if r <= MaxLatin1 {
		return '0' <= uint32(r) && uint32(r) <= '9'
	}
	return isExcludingLatin(Digit, r)
}

// IsSpace reports whether the rune is a space character as defined
// by Unicode's White Space property; in the Latin-1 space
// this is
//
//	'\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP).
//
// Other spacing characters are defined by [White_Space].
func IsSpace(r rune) bool {
	// This property isn't the same as Z; special-case it.
	if r <= MaxLatin1 {
		if r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' || r == ' ' || r == 0x85 || r == 0xA0 {
			return true
		}
		return false
	}
	return isExcludingLatin(White_Space, r)
}

// linearMax is the maximum size table for linear search for non-Latin1 rune.
// Derived by running 'go test -calibrate'.
const linearMax = 18

func isExcludingLatin(rangeTab *RangeTable, r rune) bool {
	r16 := rangeTab.R16
	// Compare as uint32 to correctly handle negative runes.
	if off := rangeTab.LatinOffset; len(r16) > off && uint32(r) <= uint32(r16[len(r16)-1].Hi) {
		return is16(r16[off:], uint16(r))
	}
	r32 := rangeTab.R32
	if len(r32) > 0 && r >= rune(r32[0].Lo) {
		return is32(r32, uint32(r))
	}
	return false
}

// is16 reports whether r is in the sorted slice of 16-bit ranges.
func is16(ranges []Range16, r uint16) bool {
	if len(ranges) <= linearMax || r <= MaxLatin1 {
		for i := range ranges {
			range_ := &ranges[i]
			if r < range_.Lo {
				return false
			}
			if r <= range_.Hi {
				return range_.Stride == 1 || (r-range_.Lo)%range_.Stride == 0
			}
		}
		return false
	}

	// binary search over ranges
	lo := 0
	hi := len(ranges)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		range_ := &ranges[m]
		if range_.Lo <= r && r <= range_.Hi {
			return range_.Stride == 1 || (r-range_.Lo)%range_.Stride == 0
		}
		if r < range_.Lo {
			hi = m
		} else {
			lo = m + 1
		}
	}
	return false
}

// is32 reports whether r is in the sorted slice of 32-bit ranges.
func is32(ranges []Range32, r uint32) bool {
	if len(ranges) <= linearMax {
		for i := range ranges {
			range_ := &ranges[i]
			if r < range_.Lo {
				return false
			}
			if r <= range_.Hi {
				return range_.Stride == 1 || (r-range_.Lo)%range_.Stride == 0
			}
		}
		return false
	}

	// binary search over ranges
	lo := 0
	hi := len(ranges)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		range_ := ranges[m]
		if range_.Lo <= r && r <= range_.Hi {
			return range_.Stride == 1 || (r-range_.Lo)%range_.Stride == 0
		}
		if r < range_.Lo {
			hi = m
		} else {
			lo = m + 1
		}
	}
	return false
}
