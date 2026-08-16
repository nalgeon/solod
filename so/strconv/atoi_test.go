// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

import (
	"errors"
	stdconv "strconv"
	"testing"
)

// The error classes that both packages report. Compare the class, not the
// error: this package returns a bare error, and the standard library wraps its
// own error in a NumError.
const (
	kindNone = iota
	kindSyntax
	kindRange
	kindOther
)

// errKind returns the class of an error.
func errKind(err error) int {
	switch {
	case err == nil:
		return kindNone
	case errors.Is(err, ErrSyntax) || errors.Is(err, stdconv.ErrSyntax):
		return kindSyntax
	case errors.Is(err, ErrRange) || errors.Is(err, stdconv.ErrRange):
		return kindRange
	}
	return kindOther
}

// addNumberSeeds adds the seed corpus of the number parsing fuzzers.
func addNumberSeeds(f *testing.F) {
	f.Add("", 10, 64)
	f.Add("0", 10, 64)
	f.Add("-0", 10, 64)
	f.Add("+42", 10, 64)
	f.Add("12345", 10, 32)
	f.Add("012345", 0, 64)
	f.Add("0x12345", 0, 64)
	f.Add("0b101", 0, 64)
	f.Add("0o17", 0, 64)
	f.Add("1_2_3", 0, 64)
	f.Add("1__2", 0, 64)
	f.Add("9223372036854775807", 10, 64)
	f.Add("-9223372036854775808", 10, 64)
	f.Add("9223372036854775808", 10, 64)
	f.Add("18446744073709551615", 10, 64)
	f.Add("18446744073709551616", 10, 64)
	f.Add("2147483648", 10, 32)
	f.Add("-2147483649", 10, 32)
	f.Add("holycow", 36, 64)
	f.Add("7fffffff", 16, 32)
	f.Add("123%45", 10, 64)
	f.Add("1", 1, 64)
	f.Add("1", 37, 64)
	f.Add("1", 10, 65)
	f.Add("1", 10, -1)
	f.Add("1", 10, 0)
}

func FuzzParseInt(f *testing.F) {
	// Compare ParseInt with the strconv package.
	addNumberSeeds(f)

	f.Fuzz(func(t *testing.T, s string, base, bitSize int) {
		got, gotErr := ParseInt(s, base, bitSize)
		want, wantErr := stdconv.ParseInt(s, base, bitSize)
		if got != want || errKind(gotErr) != errKind(wantErr) {
			t.Fatalf("ParseInt(%q, %d, %d) = %d, %v; want %d, %v",
				s, base, bitSize, got, gotErr, want, wantErr)
		}
	})
}

func FuzzParseUint(f *testing.F) {
	// Compare ParseUint with the strconv package.
	addNumberSeeds(f)

	f.Fuzz(func(t *testing.T, s string, base, bitSize int) {
		got, gotErr := ParseUint(s, base, bitSize)
		want, wantErr := stdconv.ParseUint(s, base, bitSize)
		if got != want || errKind(gotErr) != errKind(wantErr) {
			t.Fatalf("ParseUint(%q, %d, %d) = %d, %v; want %d, %v",
				s, base, bitSize, got, gotErr, want, wantErr)
		}
	})
}

func FuzzAtoi(f *testing.F) {
	// Compare Atoi with the strconv package.
	f.Add("")
	f.Add("0")
	f.Add("-0")
	f.Add("+42")
	f.Add("12345")
	f.Add("0x12345")
	f.Add("1_2_3")
	f.Add("9223372036854775807")
	f.Add("-9223372036854775808")
	f.Add("9223372036854775808")

	f.Fuzz(func(t *testing.T, s string) {
		got, gotErr := Atoi(s)
		want, wantErr := stdconv.Atoi(s)
		if got != want || errKind(gotErr) != errKind(wantErr) {
			t.Fatalf("Atoi(%q) = %d, %v; want %d, %v", s, got, gotErr, want, wantErr)
		}
	})
}
