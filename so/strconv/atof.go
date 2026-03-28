// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

// decimal to binary floating point conversion.
// Algorithm:
//   1) Store input in multiprecision decimal.
//   2) Multiply/divide decimal by powers of two until in range [0.5, 1)
//   3) Multiply by 2^precision and round to get mantissa.

// commonPrefixLenIgnoreCase returns the length of the common
// prefix of s and prefix, with the character case of s ignored.
// The prefix argument must be all lower-case.
func commonPrefixLenIgnoreCase(s, prefix string) int {
	n := min(len(prefix), len(s))
	for i := 0; i < n; i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != prefix[i] {
			return i
		}
	}
	return n
}

type specialFloat struct {
	f  float64
	n  int
	ok bool
}

// special returns the floating-point value for the special,
// possibly signed floating-point representations inf, infinity,
// and NaN. The result is ok if a prefix of s contains one
// of these representations and n is the length of that prefix.
// The character case is ignored.
func special(s string) specialFloat {
	if len(s) == 0 {
		return specialFloat{0, 0, false}
	}
	sign := 1
	nsign := 0
	switch s[0] {
	case '+', '-':
		if s[0] == '-' {
			sign = -1
		}
		nsign = 1
		s = s[1:]
		n := commonPrefixLenIgnoreCase(s, "infinity")
		if 3 < n && n < 8 {
			n = 3
		}
		if n == 3 || n == 8 {
			return specialFloat{floatInf(sign), nsign + n, true}
		}
	case 'i', 'I':
		n := commonPrefixLenIgnoreCase(s, "infinity")
		// Anything longer than "inf" is ok, but if we
		// don't have "infinity", only consume "inf".
		if 3 < n && n < 8 {
			n = 3
		}
		if n == 3 || n == 8 {
			return specialFloat{floatInf(sign), nsign + n, true}
		}
	case 'n', 'N':
		if commonPrefixLenIgnoreCase(s, "nan") == 3 {
			return specialFloat{floatNaN(), 3, true}
		}
	}
	return specialFloat{0, 0, false}
}

func (b *decimal) set(s string) bool {
	i := 0
	b.neg = false
	b.trunc = false

	// optional sign
	if i >= len(s) {
		return false
	}
	switch s[i] {
	case '+':
		i++
	case '-':
		i++
		b.neg = true
	}

	// digits
	sawdot := false
	sawdigits := false
	for ; i < len(s); i++ {
		switch {
		case s[i] == '_':
			// readFloat already checked underscores
			continue
		case s[i] == '.':
			if sawdot {
				return false
			}
			sawdot = true
			b.dp = b.nd
			continue

		case '0' <= s[i] && s[i] <= '9':
			sawdigits = true
			if s[i] == '0' && b.nd == 0 { // ignore leading zeros
				b.dp--
				continue
			}
			if b.nd < len(b.d) {
				b.d[b.nd] = s[i]
				b.nd++
			} else if s[i] != '0' {
				b.trunc = true
			}
			continue
		}
		break
	}
	if !sawdigits {
		return false
	}
	if !sawdot {
		b.dp = b.nd
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 100000).  it doesn't matter if it's
	// not the exact number.
	if i < len(s) && lower(s[i]) == 'e' {
		i++
		if i >= len(s) {
			return false
		}
		esign := 1
		switch s[i] {
		case '+':
			i++
		case '-':
			i++
			esign = -1
		}
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return false
		}
		e := 0
		for ; i < len(s) && (('0' <= s[i] && s[i] <= '9') || s[i] == '_'); i++ {
			if s[i] == '_' {
				// readFloat already checked underscores
				continue
			}
			if e < 10000 {
				e = e*10 + int(s[i]) - '0'
			}
		}
		b.dp += e * esign
	}

	if i != len(s) {
		return false
	}

	return true
}

type readFloatResult struct {
	mantissa uint64
	exp      int
	neg      bool
	trunc    bool
	hex      bool
	n        int
	ok       bool
}

// readFloat reads a decimal or hexadecimal mantissa and exponent from a float
// string representation in s; the number may be followed by other characters.
// readFloat reports the number of bytes consumed (i), and whether the number
// is valid (ok).
func readFloat(s string) readFloatResult {
	var res readFloatResult
	underscores := false

	// optional sign
	if res.n >= len(s) {
		return res
	}
	switch s[res.n] {
	case '+':
		res.n++
	case '-':
		res.n++
		res.neg = true
	}

	// digits
	base := uint64(10)
	maxMantDigits := 19 // 10^19 fits in uint64
	expChar := byte('e')
	if res.n+2 < len(s) && s[res.n] == '0' && lower(s[res.n+1]) == 'x' {
		base = 16
		maxMantDigits = 16 // 16^16 fits in uint64
		res.n += 2
		expChar = 'p'
		res.hex = true
	}
	sawdot := false
	sawdigits := false
	nd := 0
	ndMant := 0
	dp := 0
loop:
	for ; res.n < len(s); res.n++ {
		c := s[res.n]
		switch {
		case c == '_':
			underscores = true
			continue

		case c == '.':
			if sawdot {
				break loop
			}
			sawdot = true
			dp = nd
			continue

		case '0' <= c && c <= '9':
			sawdigits = true
			if c == '0' && nd == 0 { // ignore leading zeros
				dp--
				continue
			}
			nd++
			if ndMant < maxMantDigits {
				res.mantissa *= base
				res.mantissa += uint64(c - '0')
				ndMant++
			} else if c != '0' {
				res.trunc = true
			}
			continue

		case base == 16 && 'a' <= lower(c) && lower(c) <= 'f':
			sawdigits = true
			nd++
			if ndMant < maxMantDigits {
				res.mantissa *= 16
				res.mantissa += uint64(lower(c) - 'a' + 10)
				ndMant++
			} else {
				res.trunc = true
			}
			continue
		}
		break
	}
	if !sawdigits {
		return res
	}
	if !sawdot {
		dp = nd
	}

	if base == 16 {
		dp *= 4
		ndMant *= 4
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 100000).  it doesn't matter if it's
	// not the exact number.
	if res.n < len(s) && lower(s[res.n]) == expChar {
		res.n++
		if res.n >= len(s) {
			return res
		}
		esign := 1
		switch s[res.n] {
		case '+':
			res.n++
		case '-':
			res.n++
			esign = -1
		}
		if res.n >= len(s) || s[res.n] < '0' || s[res.n] > '9' {
			return res
		}
		e := 0
		for ; res.n < len(s) && (('0' <= s[res.n] && s[res.n] <= '9') || s[res.n] == '_'); res.n++ {
			if s[res.n] == '_' {
				underscores = true
				continue
			}
			if e < 10000 {
				e = e*10 + int(s[res.n]-'0')
			}
		}
		dp += e * esign
	} else if base == 16 {
		// Must have exponent.
		return res
	}

	if res.mantissa != 0 {
		res.exp = dp - ndMant
	}

	if underscores && !underscoreOK(s[:res.n]) {
		return res
	}

	res.ok = true
	return res
}

// decimal power of ten to binary power of two.
var powtab = []int{1, 3, 6, 9, 13, 16, 19, 23, 26}

func (d *decimal) floatBits(flt *floatInfo) (uint64, bool) {
	var overflowed bool
	var exp int
	var mant uint64

	// Zero is always a special case.
	if d.nd == 0 {
		mant = 0
		exp = flt.bias
		goto out
	}

	// Obvious overflow/underflow.
	// These bounds are for 64-bit floats.
	// Will have to change if we want to support 80-bit floats in the future.
	if d.dp > 310 {
		goto overflow
	}
	if d.dp < -330 {
		// zero
		mant = 0
		exp = flt.bias
		goto out
	}

	// Scale by powers of two until in range [0.5, 1.0)
	exp = 0
	for d.dp > 0 {
		var n int
		if d.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[d.dp]
		}
		d.Shift(-n)
		exp += n
	}
	for d.dp < 0 || (d.dp == 0 && d.d[0] < '5') {
		var n int
		if -d.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[-d.dp]
		}
		d.Shift(n)
		exp -= n
	}

	// Our range is [0.5,1) but floating point range is [1,2).
	exp--

	// Minimum representable exponent is flt.bias+1.
	// If the exponent is smaller, move it up and
	// adjust d accordingly.
	if exp < flt.bias+1 {
		n := flt.bias + 1 - exp
		d.Shift(-n)
		exp += n
	}

	if exp-flt.bias >= 1<<flt.expbits-1 {
		goto overflow
	}

	// Extract 1+flt.mantbits bits.
	d.Shift(int(1 + flt.mantbits))
	mant = d.RoundedInteger()

	// Rounding might have added a bit; shift down.
	if mant == 2<<flt.mantbits {
		mant >>= 1
		exp++
		if exp-flt.bias >= 1<<flt.expbits-1 {
			goto overflow
		}
	}

	// Denormalized?
	if mant&(1<<flt.mantbits) == 0 {
		exp = flt.bias
	}
	goto out

overflow:
	// ±Inf
	mant = 0
	exp = 1<<flt.expbits - 1 + flt.bias
	overflowed = true

out:
	// Assemble bits.
	bits := mant & (uint64(1)<<flt.mantbits - 1)
	bits |= uint64((exp-flt.bias)&(1<<flt.expbits-1)) << flt.mantbits
	if d.neg {
		bits |= 1 << flt.mantbits << flt.expbits
	}
	return bits, overflowed
}

// Exact powers of 10.
var float64pow10 = []float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
	1e20, 1e21, 1e22,
}
var float32pow10 = []float32{1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10}

// If possible to convert decimal representation to 64-bit float f exactly,
// entirely in floating-point math, do so, avoiding the expense of decimalToFloatBits.
// Three common cases:
//
//	value is exact integer
//	value is exact integer * exact power of ten
//	value is exact integer / exact power of ten
//
// These all produce potentially inexact but correctly rounded answers.
func atof64exact(mantissa uint64, exp int, neg bool) (float64, bool) {
	if mantissa>>float64info.mantbits != 0 {
		return 0, false
	}
	f := float64(mantissa)
	if neg {
		f = -f
	}
	switch {
	case exp == 0:
		// an integer.
		return f, true
	// Exact integers are <= 10^15.
	// Exact powers of ten are <= 10^22.
	case exp > 0 && exp <= 15+22: // int * 10^k
		// If exponent is big but number of digits is not,
		// can move a few zeros into the integer part.
		if exp > 22 {
			f *= float64pow10[exp-22]
			exp = 22
		}
		if f > 1e15 || f < -1e15 {
			// the exponent was really too large.
			return 0, false
		}
		return f * float64pow10[exp], true
	case exp < 0 && exp >= -22: // int / 10^k
		return f / float64pow10[-exp], true
	}
	return 0, false
}

// If possible to compute mantissa*10^exp to 32-bit float f exactly,
// entirely in floating-point math, do so, avoiding the machinery above.
func atof32exact(mantissa uint64, exp int, neg bool) (float32, bool) {
	if mantissa>>float32MantBits != 0 {
		return 0, false
	}
	f := float32(mantissa)
	if neg {
		f = -f
	}
	switch {
	case exp == 0:
		return f, true
	// Exact integers are <= 10^7.
	// Exact powers of ten are <= 10^10.
	case exp > 0 && exp <= 7+10: // int * 10^k
		// If exponent is big but number of digits is not,
		// can move a few zeros into the integer part.
		if exp > 10 {
			f *= float32pow10[exp-10]
			exp = 10
		}
		if f > 1e7 || f < -1e7 {
			// the exponent was really too large.
			return 0, false
		}
		return f * float32pow10[exp], true
	case exp < 0 && exp >= -10: // int / 10^k
		return f / float32pow10[-exp], true
	}
	return 0, false
}

// atofHex converts the hex floating-point string s
// to a rounded float32 or float64 value (depending on flt==&float32info or flt==&float64info)
// and returns it as a float64.
// The string s has already been parsed into a mantissa, exponent, and sign (neg==true for negative).
// If trunc is true, trailing non-zero bits have been omitted from the mantissa.
func atofHex(s string, flt *floatInfo, mantissa uint64, exp int, neg, trunc bool) (float64, error) {
	_ = s
	maxExp := 1<<flt.expbits + flt.bias - 2
	minExp := flt.bias + 1
	exp += int(flt.mantbits) // mantissa now implicitly divided by 2^mantbits.

	// Shift mantissa and exponent to bring representation into float range.
	// Eventually we want a mantissa with a leading 1-bit followed by mantbits other bits.
	// For rounding, we need two more, where the bottom bit represents
	// whether that bit or any later bit was non-zero.
	// (If the mantissa has already lost non-zero bits, trunc is true,
	// and we OR in a 1 below after shifting left appropriately.)
	for mantissa != 0 && mantissa>>(flt.mantbits+2) == 0 {
		mantissa <<= 1
		exp--
	}
	if trunc {
		mantissa |= 1
	}
	for mantissa>>(1+flt.mantbits+2) != 0 {
		mantissa = mantissa>>1 | mantissa&1
		exp++
	}

	// If exponent is too negative,
	// denormalize in hopes of making it representable.
	// (The -2 is for the rounding bits.)
	for mantissa > 1 && exp < minExp-2 {
		mantissa = mantissa>>1 | mantissa&1
		exp++
	}

	// Round using two bottom bits.
	round := mantissa & 3
	mantissa >>= 2
	round |= mantissa & 1 // round to even (round up if mantissa is odd)
	exp += 2
	if round == 3 {
		mantissa++
		if mantissa == 1<<(1+flt.mantbits) {
			mantissa >>= 1
			exp++
		}
	}

	if mantissa>>flt.mantbits == 0 { // Denormal or zero.
		exp = flt.bias
	}
	var err error
	if exp > maxExp { // infinity and range error
		mantissa = 1 << flt.mantbits
		exp = maxExp + 1
		err = ErrRange
	}

	bits := mantissa & (1<<flt.mantbits - 1)
	bits |= uint64((exp-flt.bias)&(1<<flt.expbits-1)) << flt.mantbits
	if neg {
		bits |= 1 << flt.mantbits << flt.expbits
	}
	if flt == &float32info {
		return float64(float32frombits(uint32(bits))), err
	}
	return float64frombits(bits), err
}

type atof32Result struct {
	f   float32
	n   int
	err error
}

func atof32(s string) atof32Result {
	var err error
	if spec := special(s); spec.ok {
		return atof32Result{f: float32(spec.f), n: spec.n, err: nil}
	}

	flo := readFloat(s)
	if !flo.ok {
		return atof32Result{f: 0, n: flo.n, err: ErrSyntax}
	}

	if flo.hex {
		f, err := atofHex(s[:flo.n], &float32info, flo.mantissa, flo.exp, flo.neg, flo.trunc)
		return atof32Result{f: float32(f), n: flo.n, err: err}
	}

	// Try pure floating-point arithmetic conversion, and if that fails,
	// the Eisel-Lemire algorithm.
	if !flo.trunc {
		if f, ok := atof32exact(flo.mantissa, flo.exp, flo.neg); ok {
			return atof32Result{f: f, n: flo.n, err: nil}
		}
	}
	f, ok := eiselLemire32(flo.mantissa, flo.exp, flo.neg)
	if ok {
		if !flo.trunc {
			return atof32Result{f: f, n: flo.n, err: nil}
		}
		// Even if the mantissa was truncated, we may
		// have found the correct result. Confirm by
		// converting the upper mantissa bound.
		fUp, ok := eiselLemire32(flo.mantissa+1, flo.exp, flo.neg)
		if ok && f == fUp {
			return atof32Result{f: f, n: flo.n, err: nil}
		}
	}

	// Slow fallback.
	var d decimal
	if !d.set(s[:flo.n]) {
		return atof32Result{f: 0, n: flo.n, err: ErrSyntax}
	}
	b, ovf := d.floatBits(&float32info)
	f = float32frombits(uint32(b))
	if ovf {
		err = ErrRange
	}
	return atof32Result{f: f, n: flo.n, err: err}
}

type atof64Result struct {
	f   float64
	n   int
	err error
}

func atof64(s string) atof64Result {
	var err error
	if spec := special(s); spec.ok {
		return atof64Result{f: spec.f, n: spec.n, err: nil}
	}

	flo := readFloat(s)
	if !flo.ok {
		return atof64Result{f: 0, n: flo.n, err: ErrSyntax}
	}

	if flo.hex {
		f, err := atofHex(s[:flo.n], &float64info, flo.mantissa, flo.exp, flo.neg, flo.trunc)
		return atof64Result{f: f, n: flo.n, err: err}
	}

	// Try pure floating-point arithmetic conversion, and if that fails,
	// the Eisel-Lemire algorithm.
	if !flo.trunc {
		if f, ok := atof64exact(flo.mantissa, flo.exp, flo.neg); ok {
			return atof64Result{f: f, n: flo.n, err: nil}
		}
	}
	f, ok := eiselLemire64(flo.mantissa, flo.exp, flo.neg)
	if ok {
		if !flo.trunc {
			return atof64Result{f: f, n: flo.n, err: nil}
		}
		// Even if the mantissa was truncated, we may
		// have found the correct result. Confirm by
		// converting the upper mantissa bound.
		fUp, ok := eiselLemire64(flo.mantissa+1, flo.exp, flo.neg)
		if ok && f == fUp {
			return atof64Result{f: f, n: flo.n, err: nil}
		}
	}

	// Slow fallback.
	var d decimal
	if !d.set(s[:flo.n]) {
		return atof64Result{f: 0, n: flo.n, err: ErrSyntax}
	}
	b, ovf := d.floatBits(&float64info)
	f = float64frombits(b)
	if ovf {
		err = ErrRange
	}
	return atof64Result{f: f, n: flo.n, err: err}
}

// ParseFloat converts the string s to a floating-point number
// with the precision specified by bitSize: 32 for float32, or 64 for float64.
// When bitSize=32, the result still has type float64, but it will be
// convertible to float32 without changing its value.
//
// ParseFloat accepts decimal and hexadecimal floating-point numbers
// as defined by the Go syntax for [floating-point literals].
// If s is well-formed and near a valid floating-point number,
// ParseFloat returns the nearest floating-point number rounded
// using IEEE754 unbiased rounding.
// (Parsing a hexadecimal floating-point value only rounds when
// there are more bits in the hexadecimal representation than
// will fit in the mantissa.)
//
// The errors that ParseFloat returns have concrete type *NumError
// and include err.Num = s.
//
// If s is not syntactically well-formed, ParseFloat returns err.Err = ErrSyntax.
//
// If s is syntactically well-formed but is more than 1/2 ULP
// away from the largest floating point number of the given size,
// ParseFloat returns f = ±Inf, err.Err = ErrRange.
//
// ParseFloat recognizes the string "NaN", and the (possibly signed) strings "Inf" and "Infinity"
// as their respective special floating point values. It ignores case when matching.
//
// [floating-point literals]: https://go.dev/ref/spec#Floating-point_literals
func ParseFloat(s string, bitSize int) (float64, error) {
	res := parseFloatPrefix(s, bitSize)
	if res.n != len(s) {
		return 0, ErrSyntax
	}
	return res.f, res.err
}

func parseFloatPrefix(s string, bitSize int) atof64Result {
	if bitSize == 32 {
		res := atof32(s)
		return atof64Result{f: float64(res.f), n: res.n, err: res.err}
	}
	return atof64(s)
}
