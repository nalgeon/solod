// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bytes implements functions for the manipulation of byte slices.
// It is analogous to the facilities of the [strings] package.
//
// Based on the [bytes] package, with the following modifications:
//   - Buffer is not implemented (planned).
//   - Cut returns CutResult instead of multiple return values.
//   - EqualFold is not implemented.
//   - Iterators are not implemented.
//   - Reader is not implemented (planned).
//   - Title is not implemented.
//   - ToUpperSpecial, ToLowerSpecial, and ToTitleSpecial are not implemented.
//
// [bytes]: https://github.com/golang/go/blob/go1.26.1/src/bytes/bytes.go
package bytes

import (
	"github.com/nalgeon/solod/so"
	"github.com/nalgeon/solod/so/bytealg"
	"github.com/nalgeon/solod/so/math/bits"
	"github.com/nalgeon/solod/so/mem"
	"github.com/nalgeon/solod/so/slices"
	"github.com/nalgeon/solod/so/unicode"
	"github.com/nalgeon/solod/so/unicode/utf8"
)

const maxInt = int(so.MaxInt64)

// RunePredicate reports whether the rune satisfies a condition.
type RunePredicate func(rune) bool

// RuneFunc maps a rune to another rune. If mapping returns
// a negative value, the rune is dropped from the result.
type RuneFunc func(rune) rune

// asciiSet is a 32-byte value, where each bit represents the presence of a
// given ASCII character in the set. The 128-bits of the lower 16 bytes,
// starting with the least-significant bit of the lowest word to the
// most-significant bit of the highest word, map to the full range of all
// 128 ASCII characters. The 128-bits of the upper 16 bytes will be zeroed,
// ensuring that any non-ASCII character will be reported as not in the set.
// This allocates a total of 32 bytes even though the upper half
// is unused to avoid bounds checks in asciiSet.contains.
type asciiSet struct {
	val [8]uint32
	ok  bool
}

// Equal reports whether a and b
// are the same length and contain the same bytes.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []byte) bool {
	// Neither cmd/compile nor gccgo allocates for these string conversions.
	return string(a) == string(b)
}

// Compare returns an integer comparing two byte slices lexicographically.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
// A nil argument is equivalent to an empty slice.
func Compare(a, b []byte) int {
	return bytealg.Compare(a, b)
}

// Count counts the number of non-overlapping instances of sep in s.
// If sep is an empty slice, Count returns 1 + the number of UTF-8-encoded code points in s.
func Count(s, sep []byte) int {
	// special case
	if len(sep) == 0 {
		return utf8.RuneCount(s) + 1
	}
	if len(sep) == 1 {
		return bytealg.Count(s, sep[0])
	}
	n := 0
	for {
		i := Index(s, sep)
		if i == -1 {
			return n
		}
		n++
		s = s[i+len(sep):]
	}
}

// Contains reports whether subslice is within b.
func Contains(b, subslice []byte) bool {
	return Index(b, subslice) != -1
}

// ContainsAny reports whether any of the UTF-8-encoded code points in chars are within b.
func ContainsAny(b []byte, chars string) bool {
	return IndexAny(b, chars) >= 0
}

// ContainsRune reports whether the rune is contained in the UTF-8-encoded byte slice b.
func ContainsRune(b []byte, r rune) bool {
	return IndexRune(b, r) >= 0
}

// ContainsFunc reports whether any of the UTF-8-encoded code points r within b satisfy f(r).
func ContainsFunc(b []byte, f RunePredicate) bool {
	return IndexFunc(b, f) >= 0
}

// IndexByte returns the index of the first instance of c in b, or -1 if c is not present in b.
func IndexByte(b []byte, c byte) int {
	return bytealg.IndexByte(b, c)
}

// LastIndex returns the index of the last instance of sep in s, or -1 if sep is not present in s.
func LastIndex(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return len(s)
	} else if n == 1 {
		return bytealg.LastIndexByte(s, sep[0])
	} else if n == len(s) {
		if Equal(s, sep) {
			return 0
		}
		return -1
	} else if n > len(s) {
		return -1
	}
	return bytealg.LastIndexRabinKarp(s, sep)
}

// LastIndexByte returns the index of the last instance of c in s, or -1 if c is not present in s.
func LastIndexByte(s []byte, c byte) int {
	return bytealg.LastIndexByte(s, c)
}

// IndexRune interprets s as a sequence of UTF-8-encoded code points.
// It returns the byte index of the first occurrence in s of the given rune.
// It returns -1 if rune is not present in s.
// If r is [utf8.RuneError], it returns the first instance of any
// invalid UTF-8 byte sequence.
func IndexRune(s []byte, r rune) int {
	if 0 <= r && r < utf8.RuneSelf {
		return IndexByte(s, byte(r))
	} else if r == utf8.RuneError {
		for i := 0; i < len(s); {
			r1, n := utf8.DecodeRune(s[i:])
			if r1 == utf8.RuneError {
				return i
			}
			i += n
		}
		return -1
	} else if !utf8.ValidRune(r) {
		return -1
	} else {
		// Search for rune r using the last byte of its UTF-8 encoded form.
		// The distribution of the last byte is more uniform compared to the
		// first byte which has a 78% chance of being [240, 243, 244].
		var b [utf8.UTFMax]byte
		n := utf8.EncodeRune(b[:], r)
		last := n - 1
		i := last
		fails := 0
		for i < len(s) {
			if s[i] != b[last] {
				o := IndexByte(s[i+1:], b[last])
				if o < 0 {
					return -1
				}
				i += o + 1
			}
			// Step backwards comparing bytes.
			for j := 1; j < n; j++ {
				if s[i-j] != b[last-j] {
					goto next
				}
			}
			return i - last
		next:
			fails++
			i++
			if fails >= 4+(i>>4) && i < len(s) {
				goto fallback
			}
		}
		return -1

	fallback:
		// A brute force search when IndexByte returns too many false positives.
		// A brute force search is ~1.5-3x faster than Rabin-Karp since n is small.
		c0 := b[last]
		c1 := b[last-1] // There are at least 2 chars to match
		for ; i < len(s); i++ {
			if s[i] == c0 && s[i-1] == c1 {
				for k := 2; k < n; k++ {
					if s[i-k] != b[last-k] {
						goto loop
					}
				}
				return i - last
			}
		loop:
			continue
		}

		return -1
	}
}

// IndexAny interprets s as a sequence of UTF-8-encoded Unicode code points.
// It returns the byte index of the first occurrence in s of any of the Unicode
// code points in chars. It returns -1 if chars is empty or if there is no code
// point in common.
func IndexAny(s []byte, chars string) int {
	if chars == "" {
		// Avoid scanning all of s.
		return -1
	}
	if len(s) == 1 {
		r := rune(s[0])
		if r >= utf8.RuneSelf {
			// search utf8.RuneError.
			for _, r := range chars {
				if r == utf8.RuneError {
					return 0
				}
			}
			return -1
		}
		if bytealg.IndexByteString(chars, s[0]) >= 0 {
			return 0
		}
		return -1
	}
	if len(chars) == 1 {
		r := rune(chars[0])
		if r >= utf8.RuneSelf {
			r = utf8.RuneError
		}
		return IndexRune(s, r)
	}
	if len(s) > 8 {
		if as := makeASCIISet(chars); as.ok {
			for i, c := range s {
				if as.contains(c) {
					return i
				}
			}
			return -1
		}
	}
	var width int
	for i := 0; i < len(s); i += width {
		r := rune(s[i])
		if r < utf8.RuneSelf {
			if bytealg.IndexByteString(chars, s[i]) >= 0 {
				return i
			}
			width = 1
			continue
		}
		r, width = utf8.DecodeRune(s[i:])
		if r != utf8.RuneError {
			// r is 2 to 4 bytes
			if len(chars) == width {
				if chars == string(r) {
					return i
				}
				continue
			}
		}
		for _, ch := range chars {
			if r == ch {
				return i
			}
		}
	}
	return -1
}

// LastIndexAny interprets s as a sequence of UTF-8-encoded Unicode code
// points. It returns the byte index of the last occurrence in s of any of
// the Unicode code points in chars. It returns -1 if chars is empty or if
// there is no code point in common.
func LastIndexAny(s []byte, chars string) int {
	if chars == "" {
		// Avoid scanning all of s.
		return -1
	}
	if len(s) > 8 {
		if as := makeASCIISet(chars); as.ok {
			for i := len(s) - 1; i >= 0; i-- {
				if as.contains(s[i]) {
					return i
				}
			}
			return -1
		}
	}
	if len(s) == 1 {
		r := rune(s[0])
		if r >= utf8.RuneSelf {
			for _, r := range chars {
				if r == utf8.RuneError {
					return 0
				}
			}
			return -1
		}
		if bytealg.IndexByteString(chars, s[0]) >= 0 {
			return 0
		}
		return -1
	}
	if len(chars) == 1 {
		cr := rune(chars[0])
		if cr >= utf8.RuneSelf {
			cr = utf8.RuneError
		}
		for i := len(s); i > 0; {
			r, size := utf8.DecodeLastRune(s[:i])
			i -= size
			if r == cr {
				return i
			}
		}
		return -1
	}
	for i := len(s); i > 0; {
		r := rune(s[i-1])
		if r < utf8.RuneSelf {
			if bytealg.IndexByteString(chars, s[i-1]) >= 0 {
				return i - 1
			}
			i--
			continue
		}
		r, size := utf8.DecodeLastRune(s[:i])
		i -= size
		if r != utf8.RuneError {
			// r is 2 to 4 bytes
			if len(chars) == size {
				if chars == string(r) {
					return i
				}
				continue
			}
		}
		for _, ch := range chars {
			if r == ch {
				return i
			}
		}
	}
	return -1
}

// explode splits s into a slice of UTF-8 sequences, one per Unicode code point
// (still slices of bytes), up to a maximum of n byte slices. Invalid UTF-8
// sequences are chopped into individual bytes.
//
// The returned slice is allocated; the caller owns it.
func explode(a mem.Allocator, s []byte, n int) [][]byte {
	if n <= 0 || n > len(s) {
		n = len(s)
	}
	res := mem.AllocSlice[[]byte](a, n, n)
	var size int
	na := 0
	for len(s) > 0 {
		if na+1 >= n {
			res[na] = s
			na++
			break
		}
		_, size = utf8.DecodeRune(s)
		res[na] = s[0:size:size]
		s = s[size:]
		na++
	}
	return res[0:na]
}

// Generic split: splits after each instance of sep,
// including sepSave bytes of sep in the subslices.
//
// The returned slice is allocated; the caller owns it.
func genSplit(a mem.Allocator, s, sep []byte, sepSave, n int) [][]byte {
	if n == 0 {
		return [][]byte{}
	}
	if len(sep) == 0 {
		return explode(a, s, n)
	}
	if n < 0 {
		n = Count(s, sep) + 1
	}
	if n > len(s)+1 {
		n = len(s) + 1
	}

	res := mem.AllocSlice[[]byte](a, n, n)
	n--
	i := 0
	for i < n {
		m := Index(s, sep)
		if m < 0 {
			break
		}
		res[i] = s[: m+sepSave : m+sepSave]
		s = s[m+len(sep):]
		i++
	}
	res[i] = s
	return res[:i+1]
}

// SplitN slices s into subslices separated by sep and returns a slice of
// the subslices between those separators.
// If sep is empty, SplitN splits after each UTF-8 sequence.
// The count determines the number of subslices to return:
//   - n > 0: at most n subslices; the last subslice will be the unsplit remainder;
//   - n == 0: the result is nil (zero subslices);
//   - n < 0: all subslices.
//
// To split around the first instance of a separator, see [Cut].
func SplitN(a mem.Allocator, s, sep []byte, n int) [][]byte {
	return genSplit(a, s, sep, 0, n)
}

// SplitAfterN slices s into subslices after each instance of sep and
// returns a slice of those subslices.
// If sep is empty, SplitAfterN splits after each UTF-8 sequence.
// The count determines the number of subslices to return:
//   - n > 0: at most n subslices; the last subslice will be the unsplit remainder;
//   - n == 0: the result is nil (zero subslices);
//   - n < 0: all subslices.
func SplitAfterN(a mem.Allocator, s, sep []byte, n int) [][]byte {
	return genSplit(a, s, sep, len(sep), n)
}

// Split slices s into all subslices separated by sep and returns a slice of
// the subslices between those separators.
// If sep is empty, Split splits after each UTF-8 sequence.
// It is equivalent to SplitN with a count of -1.
//
// To split around the first instance of a separator, see [Cut].
func Split(a mem.Allocator, s, sep []byte) [][]byte {
	return genSplit(a, s, sep, 0, -1)
}

// SplitAfter slices s into all subslices after each instance of sep and
// returns a slice of those subslices.
// If sep is empty, SplitAfter splits after each UTF-8 sequence.
// It is equivalent to SplitAfterN with a count of -1.
func SplitAfter(a mem.Allocator, s, sep []byte) [][]byte {
	return genSplit(a, s, sep, len(sep), -1)
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

// Fields interprets s as a sequence of UTF-8-encoded code points.
// It splits the slice s around each instance of one or more consecutive white space
// characters, as defined by [unicode.IsSpace], returning a slice of subslices of s or an
// empty slice if s contains only white space. Every element of the returned slice is
// non-empty. Unlike [Split], leading and trailing runs of white space characters
// are discarded.
//
// The returned slice is allocated; the caller owns it.
func Fields(a mem.Allocator, s []byte) [][]byte {
	// First count the fields.
	// This is an exact count if s is ASCII, otherwise it is an approximation.
	n := 0
	wasSpace := 1
	// setBits is used to track which bits are set in the bytes of s.
	setBits := uint8(0)
	for i := 0; i < len(s); i++ {
		r := s[i]
		setBits |= r
		isSpace := int(asciiSpace[r])
		n += wasSpace & ^isSpace
		wasSpace = isSpace
	}

	if setBits >= utf8.RuneSelf {
		// Some runes in the input slice are not ASCII.
		return FieldsFunc(a, s, unicode.IsSpace)
	}

	// ASCII fast path
	res := mem.AllocSlice[[]byte](a, n, n)
	na := 0
	fieldStart := 0
	i := 0
	// Skip spaces in the front of the input.
	for i < len(s) && asciiSpace[s[i]] != 0 {
		i++
	}
	fieldStart = i
	for i < len(s) {
		if asciiSpace[s[i]] == 0 {
			i++
			continue
		}
		res[na] = s[fieldStart:i:i]
		na++
		i++
		// Skip spaces in between fields.
		for i < len(s) && asciiSpace[s[i]] != 0 {
			i++
		}
		fieldStart = i
	}
	if fieldStart < len(s) { // Last field might end at EOF.
		res[na] = s[fieldStart:len(s):len(s)]
	}
	return res
}

// FieldsFunc interprets s as a sequence of UTF-8-encoded code points.
// It splits the slice s at each run of code points c satisfying f(c) and
// returns a slice of subslices of s. If all code points in s satisfy f(c), or
// len(s) == 0, an empty slice is returned. Every element of the returned slice is
// non-empty. Unlike [Split], leading and trailing runs of code points
// satisfying f(c) are discarded.
//
// FieldsFunc makes no guarantees about the order in which it calls f(c)
// and assumes that f always returns the same value for a given c.
//
// The returned slice is allocated; the caller owns it.
func FieldsFunc(a mem.Allocator, s []byte, f RunePredicate) [][]byte {
	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		start int
		end   int
	}
	spans := mem.AllocSlice[span](a, 0, 32)

	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	start := -1 // valid span start if >= 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRune(s[i:])
		if f(r) {
			if start >= 0 {
				spans = slices.Append(a, spans, span{start, i})
				start = -1
			}
		} else {
			if start < 0 {
				start = i
			}
		}
		i += size
	}

	// Last field might end at EOF.
	if start >= 0 {
		spans = slices.Append(a, spans, span{start, len(s)})
	}

	// Create subslices from recorded field indices.
	res := mem.AllocSlice[[]byte](a, len(spans), len(spans))
	for i, sp := range spans {
		res[i] = s[sp.start:sp.end:sp.end]
	}

	mem.FreeSlice(a, spans)
	return res
}

// Join concatenates the elements of s to create a new byte slice. The separator
// sep is placed between elements in the resulting slice.
//
// The returned slice is allocated; the caller owns it.
func Join(a mem.Allocator, s [][]byte, sep []byte) []byte {
	if len(s) == 0 {
		return []byte{}
	}
	if len(s) == 1 {
		// Just return a copy.
		return slices.Clone(a, s[0])
	}

	var n int
	if len(sep) > 0 {
		if len(sep) >= maxInt/(len(s)-1) {
			panic("bytes: Join output length overflow")
		}
		n += len(sep) * (len(s) - 1)
	}
	for _, v := range s {
		if len(v) > maxInt-n {
			panic("bytes: Join output length overflow")
		}
		n += len(v)
	}

	b := mem.AllocSlice[byte](a, n, n)
	bp := copy(b, s[0])
	for _, v := range s[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], v)
	}
	return b
}

// HasPrefix reports whether the byte slice s begins with prefix.
func HasPrefix(s, prefix []byte) bool {
	return len(s) >= len(prefix) && Equal(s[:len(prefix)], prefix)
}

// HasSuffix reports whether the byte slice s ends with suffix.
func HasSuffix(s, suffix []byte) bool {
	return len(s) >= len(suffix) && Equal(s[len(s)-len(suffix):], suffix)
}

// Map returns a copy of the byte slice s with all its characters modified
// according to the mapping function. If mapping returns a negative value, the character is
// dropped from the byte slice with no replacement. The characters in s and the
// output are interpreted as UTF-8-encoded code points.
//
// The returned slice is allocated; the caller owns it.
func Map(a mem.Allocator, mapping RuneFunc, s []byte) []byte {
	// In the worst case, the slice can grow when mapped, making
	// things unpleasant. But it's so rare we barge in assuming it's
	// fine. It could also shrink but that falls out naturally.
	b := mem.AllocSlice[byte](a, 0, len(s))
	for i := 0; i < len(s); {
		r, wid := utf8.DecodeRune(s[i:])
		r = mapping(r)
		if r >= 0 {
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			b = slices.Extend(a, b, buf[:n])
		}
		i += wid
	}
	return b
}

// Repeat returns a new byte slice consisting of count copies of b.
//
// It panics if count is negative or if the result of (len(b) * count)
// overflows.
//
// The returned slice is allocated; the caller owns it.
func Repeat(a mem.Allocator, b []byte, count int) []byte {
	if count == 0 {
		return []byte{}
	}

	// Since we cannot return an error on overflow,
	// we should panic if the repeat will generate an overflow.
	// See golang.org/issue/16237.
	if count < 0 {
		panic("bytes: negative Repeat count")
	}
	hi, lo := bits.Mul(uint(len(b)), uint(count))
	if hi > 0 || lo > uint(maxInt) {
		panic("bytes: Repeat output length overflow")
	}
	n := int(lo) // lo = len(b) * count

	if len(b) == 0 {
		return []byte{}
	}

	// Past a certain chunk size it is counterproductive to use
	// larger chunks as the source of the write, as when the source
	// is too large we are basically just thrashing the CPU D-cache.
	// So if the result length is larger than an empirically-found
	// limit (8KB), we stop growing the source string once the limit
	// is reached and keep reusing the same source string - that
	// should therefore be always resident in the L1 cache - until we
	// have completed the construction of the result.
	// This yields significant speedups (up to +100%) in cases where
	// the result length is large (roughly, over L2 cache size).
	const chunkLimit = 8 * 1024
	chunkMax := n
	if chunkMax > chunkLimit {
		chunkMax = chunkLimit / len(b) * len(b)
		if chunkMax == 0 {
			chunkMax = len(b)
		}
	}
	nb := mem.AllocSlice[byte](a, n, n)
	bp := copy(nb, b)
	for bp < n {
		chunk := min(bp, chunkMax)
		bp += copy(nb[bp:], nb[:chunk])
	}
	return nb
}

// ToUpper returns a copy of the byte slice s with all Unicode letters mapped to
// their upper case.
//
// The returned slice is allocated; the caller owns it.
func ToUpper(a mem.Allocator, s []byte) []byte {
	isASCII, hasLower := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasLower = hasLower || ('a' <= c && c <= 'z')
	}

	if isASCII { // optimize for ASCII-only byte slices.
		if !hasLower {
			// Just return a copy.
			return slices.Clone(a, s)
		}
		b := mem.AllocSlice[byte](a, len(s), len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if 'a' <= c && c <= 'z' {
				c -= 'a' - 'A'
			}
			b[i] = c
		}
		return b
	}
	return Map(a, unicode.ToUpper, s)
}

// ToLower returns a copy of the byte slice s with all Unicode letters mapped to
// their lower case.
//
// The returned slice is allocated; the caller owns it.
func ToLower(a mem.Allocator, s []byte) []byte {
	isASCII, hasUpper := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasUpper = hasUpper || ('A' <= c && c <= 'Z')
	}

	if isASCII { // optimize for ASCII-only byte slices.
		if !hasUpper {
			return slices.Clone(a, s)
		}
		b := mem.AllocSlice[byte](a, len(s), len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			b[i] = c
		}
		return b
	}
	return Map(a, unicode.ToLower, s)
}

// ToTitle treats s as UTF-8-encoded bytes and returns a copy with all the Unicode letters mapped to their title case.
//
// The returned slice is allocated; the caller owns it.
func ToTitle(a mem.Allocator, s []byte) []byte {
	return Map(a, unicode.ToTitle, s)
}

// ToValidUTF8 treats s as UTF-8-encoded bytes and returns a copy with each run of bytes
// representing invalid UTF-8 replaced with the bytes in replacement, which may be empty.
//
// The returned slice is allocated; the caller owns it.
func ToValidUTF8(a mem.Allocator, s, replacement []byte) []byte {
	b := mem.AllocSlice[byte](a, 0, len(s)+len(replacement))
	invalid := false // previous byte was from an invalid UTF-8 sequence
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			i++
			invalid = false
			b = slices.Append(a, b, c)
			continue
		}
		_, wid := utf8.DecodeRune(s[i:])
		if wid == 1 {
			i++
			if !invalid {
				invalid = true
				b = slices.Extend(a, b, replacement)
			}
			continue
		}
		invalid = false
		b = slices.Extend(a, b, s[i:i+wid])
		i += wid
	}
	return b
}

// TrimLeftFunc treats s as UTF-8-encoded bytes and returns a subslice of s by slicing off
// all leading UTF-8-encoded code points c that satisfy f(c).
func TrimLeftFunc(s []byte, f RunePredicate) []byte {
	i := indexFunc(s, f, false)
	if i == -1 {
		return []byte{}
	}
	return s[i:]
}

// TrimRightFunc returns a subslice of s by slicing off all trailing
// UTF-8-encoded code points c that satisfy f(c).
func TrimRightFunc(s []byte, f RunePredicate) []byte {
	i := lastIndexFunc(s, f, false)
	if i >= 0 && s[i] >= utf8.RuneSelf {
		_, wid := utf8.DecodeRune(s[i:])
		i += wid
	} else {
		i++
	}
	return s[0:i]
}

// TrimFunc returns a subslice of s by slicing off all leading and trailing
// UTF-8-encoded code points c that satisfy f(c).
func TrimFunc(s []byte, f RunePredicate) []byte {
	return TrimRightFunc(TrimLeftFunc(s, f), f)
}

// TrimPrefix returns s without the provided leading prefix string.
// If s doesn't start with prefix, s is returned unchanged.
func TrimPrefix(s, prefix []byte) []byte {
	if HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// TrimSuffix returns s without the provided trailing suffix string.
// If s doesn't end with suffix, s is returned unchanged.
func TrimSuffix(s, suffix []byte) []byte {
	if HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// IndexFunc interprets s as a sequence of UTF-8-encoded code points.
// It returns the byte index in s of the first Unicode
// code point satisfying f(c), or -1 if none do.
func IndexFunc(s []byte, f RunePredicate) int {
	return indexFunc(s, f, true)
}

// LastIndexFunc interprets s as a sequence of UTF-8-encoded code points.
// It returns the byte index in s of the last Unicode
// code point satisfying f(c), or -1 if none do.
func LastIndexFunc(s []byte, f RunePredicate) int {
	return lastIndexFunc(s, f, true)
}

// indexFunc is the same as IndexFunc except that if
// truth==false, the sense of the predicate function is
// inverted.
func indexFunc(s []byte, f RunePredicate, truth bool) int {
	start := 0
	for start < len(s) {
		r, wid := utf8.DecodeRune(s[start:])
		if f(r) == truth {
			return start
		}
		start += wid
	}
	return -1
}

// lastIndexFunc is the same as LastIndexFunc except that if
// truth==false, the sense of the predicate function is
// inverted.
func lastIndexFunc(s []byte, f RunePredicate, truth bool) int {
	for i := len(s); i > 0; {
		r, size := rune(s[i-1]), 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeLastRune(s[0:i])
		}
		i -= size
		if f(r) == truth {
			return i
		}
	}
	return -1
}

// makeASCIISet creates a set of ASCII characters and reports whether all
// characters in chars are ASCII.
func makeASCIISet(chars string) asciiSet {
	var as asciiSet
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c >= utf8.RuneSelf {
			return as
		}
		as.val[c/32] |= 1 << (c % 32)
	}
	as.ok = true
	return as
}

// contains reports whether c is inside the set.
func (as *asciiSet) contains(c byte) bool {
	return (as.val[c/32] & (1 << (c % 32))) != 0
}

// containsRune is a simplified version of strings.ContainsRune
// to avoid importing the strings package.
// We avoid bytes.ContainsRune to avoid allocating a temporary copy of s.
func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

// Trim returns a subslice of s by slicing off all leading and
// trailing UTF-8-encoded code points contained in cutset.
func Trim(s []byte, cutset string) []byte {
	if len(s) == 0 {
		return s
	}
	if cutset == "" {
		return s
	}
	if len(cutset) == 1 && cutset[0] < utf8.RuneSelf {
		return trimLeftByte(trimRightByte(s, cutset[0]), cutset[0])
	}
	if as := makeASCIISet(cutset); as.ok {
		return trimLeftASCII(trimRightASCII(s, &as), &as)
	}
	return trimLeftUnicode(trimRightUnicode(s, cutset), cutset)
}

// TrimLeft returns a subslice of s by slicing off all leading
// UTF-8-encoded code points contained in cutset.
func TrimLeft(s []byte, cutset string) []byte {
	if len(s) == 0 {
		return s
	}
	if cutset == "" {
		return s
	}
	if len(cutset) == 1 && cutset[0] < utf8.RuneSelf {
		return trimLeftByte(s, cutset[0])
	}
	if as := makeASCIISet(cutset); as.ok {
		return trimLeftASCII(s, &as)
	}
	return trimLeftUnicode(s, cutset)
}

func trimLeftByte(s []byte, c byte) []byte {
	for len(s) > 0 && s[0] == c {
		s = s[1:]
	}
	if len(s) == 0 {
		return []byte{}
	}
	return s
}

func trimLeftASCII(s []byte, as *asciiSet) []byte {
	for len(s) > 0 {
		if !as.contains(s[0]) {
			break
		}
		s = s[1:]
	}
	if len(s) == 0 {
		return []byte{}
	}
	return s
}

func trimLeftUnicode(s []byte, cutset string) []byte {
	for len(s) > 0 {
		r, n := utf8.DecodeRune(s)
		if !containsRune(cutset, r) {
			break
		}
		s = s[n:]
	}
	if len(s) == 0 {
		return []byte{}
	}
	return s
}

// TrimRight returns a subslice of s by slicing off all trailing
// UTF-8-encoded code points that are contained in cutset.
func TrimRight(s []byte, cutset string) []byte {
	if len(s) == 0 || cutset == "" {
		return s
	}
	if len(cutset) == 1 && cutset[0] < utf8.RuneSelf {
		return trimRightByte(s, cutset[0])
	}
	if as := makeASCIISet(cutset); as.ok {
		return trimRightASCII(s, &as)
	}
	return trimRightUnicode(s, cutset)
}

func trimRightByte(s []byte, c byte) []byte {
	for len(s) > 0 && s[len(s)-1] == c {
		s = s[:len(s)-1]
	}
	return s
}

func trimRightASCII(s []byte, as *asciiSet) []byte {
	for len(s) > 0 {
		if !as.contains(s[len(s)-1]) {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}

func trimRightUnicode(s []byte, cutset string) []byte {
	for len(s) > 0 {
		r, n := rune(s[len(s)-1]), 1
		if r >= utf8.RuneSelf {
			r, n = utf8.DecodeLastRune(s)
		}
		if !containsRune(cutset, r) {
			break
		}
		s = s[:len(s)-n]
	}
	return s
}

// TrimSpace returns a subslice of s by slicing off all leading and
// trailing white space, as defined by Unicode.
func TrimSpace(s []byte) []byte {
	// Fast path for ASCII: look for the first ASCII non-space byte.
	for lo, c := range s {
		if c >= utf8.RuneSelf {
			// If we run into a non-ASCII byte, fall back to the
			// slower unicode-aware method on the remaining bytes.
			return TrimFunc(s[lo:], unicode.IsSpace)
		}
		if asciiSpace[c] != 0 {
			continue
		}
		s = s[lo:]
		// Now look for the first ASCII non-space byte from the end.
		for hi := len(s) - 1; hi >= 0; hi-- {
			c := s[hi]
			if c >= utf8.RuneSelf {
				return TrimFunc(s[:hi+1], unicode.IsSpace)
			}
			if asciiSpace[c] == 0 {
				// At this point, s[:hi+1] starts and ends with ASCII
				// non-space bytes, so we're done. Non-ASCII cases have
				// already been handled above.
				return s[:hi+1]
			}
		}
	}
	return []byte{}
}

// Runes interprets s as a sequence of UTF-8-encoded code points.
// It returns a slice of runes (Unicode code points) equivalent to s.
//
// The returned slice is allocated; the caller owns it.
func Runes(a mem.Allocator, s []byte) []rune {
	tlen := utf8.RuneCount(s)
	t := mem.AllocSlice[rune](a, tlen, tlen)
	i := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		t[i] = r
		i++
		s = s[l:]
	}
	return t
}

// Replace returns a copy of the slice s with the first n
// non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the slice
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune slice.
// If n < 0, there is no limit on the number of replacements.
//
// The returned slice is allocated; the caller owns it.
func Replace(a mem.Allocator, s, old, new []byte, n int) []byte {
	m := 0
	if n != 0 {
		// Compute number of replacements.
		m = Count(s, old)
	}
	if m == 0 {
		// Just return a copy.
		return slices.Clone(a, s)
	}
	if n < 0 || m < n {
		n = m
	}

	// Apply replacements to buffer.
	tlen := len(s) + n*(len(new)-len(old))
	t := mem.AllocSlice[byte](a, tlen, tlen)
	w := 0
	start := 0
	if len(old) > 0 {
		for range n {
			j := start + Index(s[start:], old)
			w += copy(t[w:], s[start:j])
			w += copy(t[w:], new)
			start = j + len(old)
		}
	} else { // len(old) == 0
		w += copy(t[w:], new)
		for range n - 1 {
			_, wid := utf8.DecodeRune(s[start:])
			j := start + wid
			w += copy(t[w:], s[start:j])
			w += copy(t[w:], new)
			start = j
		}
	}
	w += copy(t[w:], s[start:])
	return t[0:w]
}

// ReplaceAll returns a copy of the slice s with all
// non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the slice
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune slice.
//
// The returned slice is allocated; the caller owns it.
func ReplaceAll(a mem.Allocator, s, old, new []byte) []byte {
	return Replace(a, s, old, new, -1)
}

// Index returns the index of the first instance of sep in s, or -1 if sep is not present in s.
func Index(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return 0
	} else if n == 1 {
		return IndexByte(s, sep[0])
	} else if n == len(s) {
		if Equal(sep, s) {
			return 0
		}
		return -1
	} else if n > len(s) {
		return -1
	} else if n <= bytealg.MaxLen {
		c0 := sep[0]
		c1 := sep[1]
		i := 0
		t := len(s) - n + 1
		for i < t {
			if s[i] != c0 {
				// IndexByte is faster than bytealg.Index, so use it as long as
				// we're not getting lots of false positives.
				o := IndexByte(s[i+1:t], c0)
				if o < 0 {
					return -1
				}
				i += o + 1
			}
			if s[i+1] == c1 && Equal(s[i:i+n], sep) {
				return i
			}
			i++
		}
		return -1
	}
	c0 := sep[0]
	c1 := sep[1]
	i := 0
	fails := 0
	t := len(s) - n + 1
	for i < t {
		if s[i] != c0 {
			o := IndexByte(s[i+1:t], c0)
			if o < 0 {
				break
			}
			i += o + 1
		}
		if s[i+1] == c1 && Equal(s[i:i+n], sep) {
			return i
		}
		i++
		fails++
		if fails >= (4+(i>>4)) && i < t {
			// Give up on IndexByte, it isn't skipping ahead
			// far enough to be better than Rabin-Karp.
			// Experiments (using IndexPeriodic) suggest
			// the cutover is about 16 byte skips.
			// TODO: if large prefixes of sep are matching
			// we should cutover at even larger average skips,
			// because Equal becomes that much more expensive.
			// This code does not take that effect into account.
			j := bytealg.IndexRabinKarp(s[i:], sep)
			if j < 0 {
				return -1
			}
			return i + j
		}
	}
	return -1
}

type CutResult struct {
	Before []byte
	After  []byte
	Found  bool
}

// Cut slices s around the first instance of sep,
// returning the text before and after sep.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, nil, false.
//
// Cut returns slices of the original slice s, not copies.
func Cut(s, sep []byte) CutResult {
	var res CutResult
	if i := Index(s, sep); i >= 0 {
		res.Before = s[:i]
		res.After = s[i+len(sep):]
		res.Found = true
		return res
	}
	res.Before = s
	return res
}

// Clone returns a copy of b[:len(b)].
// The returned slice is allocated; the caller owns it.
func Clone(a mem.Allocator, b []byte) []byte {
	return slices.Clone(a, b)
}

// CutPrefix returns s without the provided leading prefix byte slice
// and reports whether it found the prefix.
// If s doesn't start with prefix, CutPrefix returns s, false.
// If prefix is the empty byte slice, CutPrefix returns s, true.
//
// CutPrefix returns slices of the original slice s, not copies.
func CutPrefix(s, prefix []byte) ([]byte, bool) {
	if !HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

// CutSuffix returns s without the provided ending suffix byte slice
// and reports whether it found the suffix.
// If s doesn't end with suffix, CutSuffix returns s, false.
// If suffix is the empty byte slice, CutSuffix returns s, true.
//
// CutSuffix returns slices of the original slice s, not copies.
func CutSuffix(s, suffix []byte) ([]byte, bool) {
	if !HasSuffix(s, suffix) {
		return s, false
	}
	return s[:len(s)-len(suffix)], true
}
