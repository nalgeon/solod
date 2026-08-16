// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The shared test data and helpers of the strings tests.
// This file is named to sort before the other test files
// so its definitions are available to the other files.
package strings_test

import (
	"solod.dev/so/io"
	"solod.dev/so/strings"
	"solod.dev/so/unicode"
	"solod.dev/so/unicode/utf8"
)

// The strings of several test cases.
const (
	abcd   = "abcd"
	faces  = "☺☻☹"
	commas = "1,2,3,4"
	dots   = "1....2....3....4"
)

// space holds one space code point of every width.
const space = "\t\v\r\f\n\u0085\u00a0\u2000\u3000"

// The strings that hold another string several times.
const (
	// faces + faces
	facesTwice = "☺☻☹☺☻☹"
	// dots + dots + dots
	dotsThrice = "1....2....3....41....2....3....41....2....3....4"
	// space + "abc" + space
	spaceAbc = "\t\v\r\f\n\u0085\u00a0\u2000\u3000abc\t\v\r\f\n\u0085\u00a0\u2000\u3000"
	// space + " hello " + space
	spaceHello = "\t\v\r\f\n\u0085\u00a0\u2000\u3000 hello \t\v\r\f\n\u0085\u00a0\u2000\u3000"
	// "hello" + space + "hello"
	helloSpace = "hello\t\v\r\f\n\u0085\u00a0\u2000\u3000hello"
)

// hexDigits holds the digits of a hexadecimal number.
const hexDigits = "0123456789abcdef"

// dump writes s into buf as hexadecimal and returns the result.
func dump(buf []byte, s string) string {
	for i := 0; i < len(s); i++ {
		buf[2*i] = hexDigits[s[i]>>4]
		buf[2*i+1] = hexDigits[s[i]&0xf]
	}
	return string(buf[:2*len(s)])
}

// The errors of the reader tests, by number.
const (
	errNone = iota
	errEOF
	errOffset
	errWhence
	errUnread
	errOther
)

// errCode returns the number of the error.
func errCode(err error) int {
	if err == nil {
		return errNone
	}
	if err == io.EOF {
		return errEOF
	}
	if err == io.ErrOffset {
		return errOffset
	}
	if err == io.ErrWhence {
		return errWhence
	}
	if err == io.ErrUnread {
		return errUnread
	}
	return errOther
}

// errName returns the name of the error with the number.
func errName(code int) string {
	switch code {
	case errNone:
		return "nil"
	case errEOF:
		return "EOF"
	case errOffset:
		return "ErrOffset"
	case errWhence:
		return "ErrWhence"
	case errUnread:
		return "ErrUnread"
	}
	return "other"
}

// maxWord is the length of the longest word of a search sweep.
const maxWord = 8

// maxSep is the length of the longest separator of a search sweep.
const maxSep = 4

// A sweep enumerates every word of an alphabet up to a length.
type sweep struct {
	alpha   string // the letters of the words
	maxWord int    // the length of the longest word
	maxSep  int    // the length of the longest separator
}

// The sweeps of the tests. The first alphabet has two letters, so its words
// repeat and overlap in every way. The second alphabet has a NUL byte and a
// high byte, which the functions must treat as ordinary bytes.
var sweeps = []sweep{
	{"ab", maxWord, maxSep},
	{"\x00a\xff", 5, 3},
}

// The sweeps of the tests that allocate.
var allocSweeps = []sweep{
	{"ab", 5, 2},
	{"\x00a\xff", 3, 2},
}

// wordCount returns the number of words of alpha with the length n.
func wordCount(alpha string, n int) int {
	count := 1
	for range n {
		count *= len(alpha)
	}
	return count
}

// wordTotal returns the number of words of alpha with a length up to max.
func wordTotal(alpha string, max int) int {
	total := 0
	for n := 0; n <= max; n++ {
		total += wordCount(alpha, n)
	}
	return total
}

// wordAt writes the word number i of alpha into buf and returns the result.
// The shorter words come first, and every word with a length up to max appears
// once. The caller must keep i below wordTotal(alpha, max).
func wordAt(buf []byte, alpha string, max, i int) string {
	for n := 0; n <= max; n++ {
		count := wordCount(alpha, n)
		if i < count {
			for k := 0; k < n; k++ {
				buf[k] = alpha[i%len(alpha)]
				i /= len(alpha)
			}
			return string(buf[:n])
		}
		i -= count
	}
	return ""
}

// The reference implementations. Every sweep checks the package against the
// simplest code that gives the wanted result.

// indexBrute returns the index of the first sep in s, or -1.
func indexBrute(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// lastIndexBrute returns the index of the last sep in s, or -1.
func lastIndexBrute(s, sep string) int {
	for i := len(s) - len(sep); i >= 0; i-- {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// countBrute returns the number of separate sep in s. The separator must not
// be empty, because an empty separator counts the code points of s.
func countBrute(s, sep string) int {
	n := 0
	for i := 0; i+len(sep) <= len(s); {
		if s[i:i+len(sep)] == sep {
			n++
			i += len(sep)
			continue
		}
		i++
	}
	return n
}

// nextRand returns the next value of a pseudo random sequence.
func nextRand(x uint32) uint32 {
	return x*1664525 + 1013904223
}

// partSep separates the wanted substrings of a split case.
// No case holds the byte 0x02.
const partSep = "\x02"

// partAt returns the substring with the index i of a joined want string.
func partAt(want string, i int) string {
	start := 0
	for range i {
		for start < len(want) && want[start] != partSep[0] {
			start++
		}
		start++
	}
	end := start
	for end < len(want) && want[end] != partSep[0] {
		end++
	}
	if start > len(want) {
		return ""
	}
	return want[start:end]
}

// The predicates of the index and trim cases, by number.
const (
	predSpace = iota
	predDigit
	predUpper
	predValidRune
	predNotSpace
	predNotDigit
	predNotValidRune
)

// isValidRune reports whether the decoding of the code point succeeded.
func isValidRune(r rune) bool {
	return r != utf8.RuneError
}

// notSpace reports whether the code point is not a space.
func notSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

// notDigit reports whether the code point is not a decimal digit.
func notDigit(r rune) bool {
	return !unicode.IsDigit(r)
}

// notValidRune reports whether the decoding of the code point failed.
func notValidRune(r rune) bool {
	return r == utf8.RuneError
}

// predicate returns the predicate with the number.
func predicate(pred int) strings.RunePredicate {
	switch pred {
	case predSpace:
		return unicode.IsSpace
	case predDigit:
		return unicode.IsDigit
	case predUpper:
		return unicode.IsUpper
	case predValidRune:
		return isValidRune
	case predNotSpace:
		return notSpace
	case predNotDigit:
		return notDigit
	}
	return notValidRune
}

// predName returns the name of the predicate with the number.
func predName(pred int) string {
	switch pred {
	case predSpace:
		return "IsSpace"
	case predDigit:
		return "IsDigit"
	case predUpper:
		return "IsUpper"
	case predValidRune:
		return "IsValidRune"
	case predNotSpace:
		return "not IsSpace"
	case predNotDigit:
		return "not IsDigit"
	}
	return "not IsValidRune"
}
