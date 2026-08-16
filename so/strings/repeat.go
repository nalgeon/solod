// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"solod.dev/so/math/bits"
	"solod.dev/so/mem"
)

// maxInt is the maximum value of an int.
const maxInt = int(uint64(^uint(0)) >> 1)

// According to static analysis, spaces, dashes, zeros, equals, and tabs
// are the most commonly repeated string literal,
// often used for display on fixed-width terminal windows.
// Pre-declare constants for these for O(1) repetition in the common-case.
const (
	repeatedSpaces string = "" +
		"                                                                " +
		"                                                                "
	repeatedDashes string = "" +
		"----------------------------------------------------------------" +
		"----------------------------------------------------------------"
	repeatedZeroes string = "" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	repeatedEquals string = "" +
		"================================================================" +
		"================================================================"
	repeatedTabs string = "" +
		"\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t" +
		"\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
)

// Repeat returns a new string consisting of count copies of the string s.
//
// It panics if count is negative or if the result of (len(s) * count)
// overflows.
//
// If the allocator is nil, uses the system allocator.
// The returned string is allocated; the caller owns it.
func Repeat(a mem.Allocator, s string, count int) string {
	if count == 0 {
		return ""
	} else if count == 1 {
		return Clone(a, s)
	}

	// Since we cannot return an error on overflow,
	// we should panic if the repeat will generate an overflow.
	// See golang.org/issue/16237.
	if count < 0 {
		panic("strings: negative repeat count")
	}
	hi, lo := bits.Mul(uint(len(s)), uint(count))
	if hi > 0 || lo > uint(maxInt) {
		panic("strings: repeat overflow")
	}
	n := int(lo) // lo = len(s) * count

	if len(s) == 0 {
		return ""
	}

	// Optimize for commonly repeated strings of relatively short length.
	if s[0] == ' ' || s[0] == '-' || s[0] == '0' || s[0] == '=' || s[0] == '\t' {
		if n <= len(repeatedSpaces) && HasPrefix(repeatedSpaces, s) {
			return Clone(a, repeatedSpaces[:n])
		} else if n <= len(repeatedDashes) && HasPrefix(repeatedDashes, s) {
			return Clone(a, repeatedDashes[:n])
		} else if n <= len(repeatedZeroes) && HasPrefix(repeatedZeroes, s) {
			return Clone(a, repeatedZeroes[:n])
		} else if n <= len(repeatedEquals) && HasPrefix(repeatedEquals, s) {
			return Clone(a, repeatedEquals[:n])
		} else if n <= len(repeatedTabs) && HasPrefix(repeatedTabs, s) {
			return Clone(a, repeatedTabs[:n])
		}
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
	if n > chunkLimit {
		chunkMax = chunkLimit / len(s) * len(s)
		if chunkMax == 0 {
			chunkMax = len(s)
		}
	}

	b := Builder{a: a}
	b.Grow(n)
	b.WriteString(s)
	for b.Len() < n {
		chunk := min(b.Len(), n-b.Len())
		if chunkMax < chunk {
			chunk = chunkMax
		}
		b.WriteString(b.String()[:chunk])
	}
	return b.String()
}
