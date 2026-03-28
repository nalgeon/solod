// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

type Uint128 = uint128

const (
	Pow10Min = pow10Min
	Pow10Max = pow10Max
)

var (
	MulLog10_2       = mulLog10_2
	MulLog2_10       = mulLog2_10
	ParseFloatPrefix = parseFloatPrefix
	Pow10            = intPow10
	Umul128          = umul128
	Umul192          = umul192
	Div5Tab          = div5Tab
	DivisiblePow5    = divisiblePow5
	TrimZeros        = trimZeros
)

func NewDecimal(i uint64) *decimal {
	d := new(decimal)
	d.Assign(i)
	return d
}

func (a *decimal) String(buf []byte) string {
	w := 0
	switch {
	case a.nd == 0:
		return "0"

	case a.dp <= 0:
		// zeros fill space between decimal point and digits
		buf[w] = '0'
		w++
		buf[w] = '.'
		w++
		w += digitZero(buf[w : w+-a.dp])
		w += copy(buf[w:], a.d[0:a.nd])

	case a.dp < a.nd:
		// decimal point in middle of digits
		w += copy(buf[w:], a.d[0:a.dp])
		buf[w] = '.'
		w++
		w += copy(buf[w:], a.d[a.dp:a.nd])

	default:
		// zeros fill space between digits and decimal point
		w += copy(buf[w:], a.d[0:a.nd])
		w += digitZero(buf[w : w+a.dp-a.nd])
	}
	return string(buf[0:w])
}

func digitZero(dst []byte) int {
	for i := range dst {
		dst[i] = '0'
	}
	return len(dst)
}
