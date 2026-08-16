package strconv_test

import (
	"solod.dev/so/math"
	"solod.dev/so/strconv"
)

// The errors of the package, by number. A table holds the number instead of
// the error, because a package level table must not depend on the
// initialization order of the error values.
//
// This file is named to sort before the other test files. A package level
// table must not name a constant declared in an alphabetically later file.
const (
	errNone = iota
	errSyntax
	errRange
	errBase
	errBitSize
	errOther
)

// errCode returns the number of the error.
func errCode(err error) int {
	if err == nil {
		return errNone
	}
	if err == strconv.ErrSyntax {
		return errSyntax
	}
	if err == strconv.ErrRange {
		return errRange
	}
	if err == strconv.ErrBase {
		return errBase
	}
	if err == strconv.ErrBitSize {
		return errBitSize
	}
	return errOther
}

// errName returns the name of the error with the number.
func errName(code int) string {
	switch code {
	case errNone:
		return "nil"
	case errSyntax:
		return "ErrSyntax"
	case errRange:
		return "ErrRange"
	case errBase:
		return "ErrBase"
	case errBitSize:
		return "ErrBitSize"
	}
	return "other"
}

// bufLen is the length of a scratch buffer. It holds the longest string that
// the package writes, plus the prefix that an append test puts before it.
const bufLen = 128

// prefix goes before an appended value, to check that an Append function keeps
// the bytes that the destination already holds.
const prefix = "abc"

// appended returns the text that an Append function put after the prefix,
// or an empty string if the prefix is gone.
func appended(dst []byte) string {
	if len(dst) < len(prefix) || string(dst[:len(prefix)]) != prefix {
		return ""
	}
	return string(dst[len(prefix):])
}

// fdiv divides two float64 values. A call keeps the division out of the
// constant arithmetic of the compiler.
func fdiv(a, b float64) float64 { return a / b }

// pow2 returns two raised to the power exp. pow2 builds the bit pattern of the
// value, because math.Ldexp needs libm and a freestanding target has none.
func pow2(exp int) float64 {
	const (
		bias     = 1023  // the exponent of the value one
		minExp   = -1074 // the exponent of the smallest denormal
		mantBits = 52
	)
	switch {
	case exp < minExp:
		return 0
	case exp < -bias+1:
		return math.Float64frombits(1 << uint(exp-minExp))
	case exp <= bias:
		return math.Float64frombits(uint64(exp+bias) << mantBits)
	}
	return math.Inf(0)
}
