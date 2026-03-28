// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv_test

import (
	"testing"

	"solod.dev/so/math"
	. "solod.dev/so/strconv"
)

type ftoaTest struct {
	f    float64
	fmt  byte
	prec int
	s    string
}

func fdiv(a, b float64) float64 { return a / b }

const (
	below1e23 = 99999999999999974834176
	above1e23 = 100000000000000008388608
)

var ftoatests = []ftoaTest{
	{1, 'e', 5, "1.00000e+00"},
	{1, 'f', 5, "1.00000"},
	{1, 'g', 5, "1"},
	{1, 'g', -1, "1"},
	{1, 'x', -1, "0x1p+00"},
	{1, 'x', 5, "0x1.00000p+00"},
	{20, 'g', -1, "20"},
	{20, 'x', -1, "0x1.4p+04"},
	{1234567.8, 'g', -1, "1.2345678e+06"},
	{1234567.8, 'x', -1, "0x1.2d687cccccccdp+20"},
	{200000, 'g', -1, "200000"},
	{200000, 'x', -1, "0x1.86ap+17"},
	{200000, 'X', -1, "0X1.86AP+17"},
	{2000000, 'g', -1, "2e+06"},
	{1e10, 'g', -1, "1e+10"},

	// f conversion basic cases
	{12345, 'f', 2, "12345.00"},
	{1234.5, 'f', 2, "1234.50"},
	{123.45, 'f', 2, "123.45"},
	{12.345, 'f', 2, "12.35"},
	{1.2345, 'f', 2, "1.23"},
	{0.12345, 'f', 2, "0.12"},
	{0.12945, 'f', 2, "0.13"},
	{0.012345, 'f', 2, "0.01"},
	{0.015, 'f', 2, "0.01"},
	{0.016, 'f', 2, "0.02"},
	{0.0052345, 'f', 2, "0.01"},
	{0.0012345, 'f', 2, "0.00"},
	{0.00012345, 'f', 2, "0.00"},
	{0.000012345, 'f', 2, "0.00"},

	{0.996644984, 'f', 6, "0.996645"},
	{0.996644984, 'f', 5, "0.99664"},
	{0.996644984, 'f', 4, "0.9966"},
	{0.996644984, 'f', 3, "0.997"},
	{0.996644984, 'f', 2, "1.00"},
	{0.996644984, 'f', 1, "1.0"},

	// g conversion and zero suppression
	{400, 'g', 2, "4e+02"},
	{40, 'g', 2, "40"},
	{4, 'g', 2, "4"},
	{.4, 'g', 2, "0.4"},
	{.04, 'g', 2, "0.04"},
	{.004, 'g', 2, "0.004"},
	{.0004, 'g', 2, "0.0004"},
	{.00004, 'g', 2, "4e-05"},
	{.000004, 'g', 2, "4e-06"},

	{0, 'e', 5, "0.00000e+00"},
	{0, 'f', 5, "0.00000"},
	{0, 'g', 5, "0"},
	{0, 'g', -1, "0"},
	{0, 'x', 5, "0x0.00000p+00"},

	{-1, 'e', 5, "-1.00000e+00"},
	{-1, 'f', 5, "-1.00000"},
	{-1, 'g', 5, "-1"},
	{-1, 'g', -1, "-1"},

	{12, 'e', 5, "1.20000e+01"},
	{12, 'f', 5, "12.00000"},
	{12, 'g', 5, "12"},
	{12, 'g', -1, "12"},

	{123456700, 'e', 5, "1.23457e+08"},
	{123456700, 'f', 5, "123456700.00000"},
	{123456700, 'g', 5, "1.2346e+08"},
	{123456700, 'g', -1, "1.234567e+08"},

	{1.2345e6, 'e', 5, "1.23450e+06"},
	{1.2345e6, 'f', 5, "1234500.00000"},
	{1.2345e6, 'g', 5, "1.2345e+06"},

	// Round to even
	{1.2345e6, 'e', 3, "1.234e+06"},
	{1.2355e6, 'e', 3, "1.236e+06"},
	{1.2345, 'f', 3, "1.234"},
	{1.2355, 'f', 3, "1.236"},
	{1234567890123456.5, 'e', 15, "1.234567890123456e+15"},
	{1234567890123457.5, 'e', 15, "1.234567890123458e+15"},
	{108678236358137.625, 'g', -1, "1.0867823635813762e+14"},

	{1e23, 'e', 17, "9.99999999999999916e+22"},
	{1e23, 'f', 17, "99999999999999991611392.00000000000000000"},
	{1e23, 'g', 17, "9.9999999999999992e+22"},

	{1e23, 'e', -1, "1e+23"},
	{1e23, 'f', -1, "100000000000000000000000"},
	{1e23, 'g', -1, "1e+23"},

	{below1e23, 'e', 17, "9.99999999999999748e+22"},
	{below1e23, 'f', 17, "99999999999999974834176.00000000000000000"},
	{below1e23, 'g', 17, "9.9999999999999975e+22"},

	{below1e23, 'e', -1, "9.999999999999997e+22"},
	{below1e23, 'f', -1, "99999999999999970000000"},
	{below1e23, 'g', -1, "9.999999999999997e+22"},

	{above1e23, 'e', 17, "1.00000000000000008e+23"},
	{above1e23, 'f', 17, "100000000000000008388608.00000000000000000"},
	{above1e23, 'g', 17, "1.0000000000000001e+23"},

	{above1e23, 'e', -1, "1.0000000000000001e+23"},
	{above1e23, 'f', -1, "100000000000000010000000"},
	{above1e23, 'g', -1, "1.0000000000000001e+23"},

	{fdiv(5e-304, 1e20), 'g', -1, "5e-324"},   // avoid constant arithmetic
	{fdiv(-5e-304, 1e20), 'g', -1, "-5e-324"}, // avoid constant arithmetic

	{32, 'g', -1, "32"},
	{32, 'g', 0, "3e+01"},

	{100, 'x', -1, "0x1.9p+06"},
	{100, 'y', -1, "%y"},

	{math.NaN(), 'g', -1, "NaN"},
	{-math.NaN(), 'g', -1, "NaN"},
	{math.Inf(0), 'g', -1, "+Inf"},
	{math.Inf(-1), 'g', -1, "-Inf"},
	{-math.Inf(0), 'g', -1, "-Inf"},

	{-1, 'b', -1, "-4503599627370496p-52"},

	// fixed bugs
	{0.9, 'f', 1, "0.9"},
	{0.09, 'f', 1, "0.1"},
	{0.0999, 'f', 1, "0.1"},
	{0.05, 'f', 1, "0.1"},
	{0.05, 'f', 0, "0"},
	{0.5, 'f', 1, "0.5"},
	{0.5, 'f', 0, "0"},
	{1.5, 'f', 0, "2"},

	// https://www.exploringbinary.com/java-hangs-when-converting-2-2250738585072012e-308/
	{2.2250738585072012e-308, 'g', -1, "2.2250738585072014e-308"},
	// https://www.exploringbinary.com/php-hangs-on-numeric-value-2-2250738585072011e-308/
	{2.2250738585072011e-308, 'g', -1, "2.225073858507201e-308"},

	// Issue 2625.
	{383260575764816448, 'f', 0, "383260575764816448"},
	{383260575764816448, 'g', -1, "3.8326057576481645e+17"},

	// Issue 29491.
	{498484681984085570, 'f', -1, "498484681984085570"},
	{-5.8339553793802237e+23, 'g', -1, "-5.8339553793802237e+23"},

	// Issue 52187
	{123.45, '?', 0, "%?"},
	{123.45, '?', 1, "%?"},
	{123.45, '?', -1, "%?"},

	// rounding
	{2.275555555555555, 'x', -1, "0x1.23456789abcdep+01"},
	{2.275555555555555, 'x', 0, "0x1p+01"},
	{2.275555555555555, 'x', 2, "0x1.23p+01"},
	{2.275555555555555, 'x', 16, "0x1.23456789abcde000p+01"},
	{2.275555555555555, 'x', 21, "0x1.23456789abcde00000000p+01"},
	{2.2755555510520935, 'x', -1, "0x1.2345678p+01"},
	{2.2755555510520935, 'x', 6, "0x1.234568p+01"},
	{2.275555431842804, 'x', -1, "0x1.2345668p+01"},
	{2.275555431842804, 'x', 6, "0x1.234566p+01"},
	{3.999969482421875, 'x', -1, "0x1.ffffp+01"},
	{3.999969482421875, 'x', 4, "0x1.ffffp+01"},
	{3.999969482421875, 'x', 3, "0x1.000p+02"},
	{3.999969482421875, 'x', 2, "0x1.00p+02"},
	{3.999969482421875, 'x', 1, "0x1.0p+02"},
	{3.999969482421875, 'x', 0, "0x1p+02"},

	// Cases that Java once mishandled, from David Chase.
	{1.801439850948199e+16, 'g', -1, "1.801439850948199e+16"},
	{5.960464477539063e-08, 'g', -1, "5.960464477539063e-08"},
	{1.012e-320, 'g', -1, "1.012e-320"},

	// Cases from TestFtoaRandom that caught bugs in fixedFtoa.
	{8177880169308380. * (1 << 1), 'e', 14, "1.63557603386168e+16"},
	{8393378656576888. * (1 << 1), 'e', 15, "1.678675731315378e+16"},
	{8738676561280626. * (1 << 4), 'e', 16, "1.3981882498049002e+17"},
	{8291032395191335. / (1 << 30), 'e', 5, "7.72163e+06"},
	{8880392441509914. / (1 << 80), 'e', 16, "7.3456884594794477e-09"},

	// Exercise divisiblePow5 case in fixedFtoa
	{2384185791015625. * (1 << 12), 'e', 5, "9.76562e+18"},
	{2384185791015625. * (1 << 13), 'e', 5, "1.95312e+19"},

	// Exercise potential mistakes in fixedFtoa.
	// Found by introducing mistakes and running 'go test -testbase'.
	{0x1.000000000005p+71, 'e', 16, "2.3611832414348645e+21"},
	{0x1.0000p-27, 'e', 17, "7.45058059692382812e-09"},
	{0x1.0000p-41, 'e', 17, "4.54747350886464119e-13"},
}

func TestFtoa(t *testing.T) {
	buf := make([]byte, 64)
	for i := 0; i < len(ftoatests); i++ {
		test := &ftoatests[i]
		s := FormatFloat(buf, test.f, test.fmt, test.prec, 64)
		if s != test.s {
			t.Error("testN=64", test.f, string(test.fmt), test.prec, "want", test.s, "got", s)
		}
		x := AppendFloat([]byte("abc"), test.f, test.fmt, test.prec, 64)
		if string(x) != "abc"+test.s {
			t.Error("AppendFloat testN=64", test.f, string(test.fmt), test.prec, "want", "abc"+test.s, "got", string(x))
		}
		if float64(float32(test.f)) == test.f && test.fmt != 'b' {
			test_s := test.s
			if test.f == 5.960464477539063e-08 {
				// This test is an exact float32 but asking for float64 precision in the string.
				// (All our other float64-only tests fail to exactness check above.)
				test_s = "5.9604645e-08"
				continue
			}
			s := FormatFloat(buf, test.f, test.fmt, test.prec, 32)
			if s != test.s {
				t.Error("testN=32", test.f, string(test.fmt), test.prec, "want", test_s, "got", s)
			}
			x := AppendFloat([]byte("abc"), test.f, test.fmt, test.prec, 32)
			if string(x) != "abc"+test_s {
				t.Error("AppendFloat testN=32", test.f, string(test.fmt), test.prec, "want", "abc"+test_s, "got", string(x))
			}
		}
	}
}

func TestFtoaPowersOfTwo(t *testing.T) {
	buf := make([]byte, 64)
	for exp := -2048; exp <= 2048; exp++ {
		f := math.Ldexp(1, exp)
		if !math.IsInf(f, 0) {
			s := FormatFloat(buf, f, 'e', -1, 64)
			if x, _ := ParseFloat(s, 64); x != f {
				t.Errorf("failed roundtrip %v => %s => %v", f, s, x)
			}
		}
		f32 := float32(f)
		if !math.IsInf(float64(f32), 0) {
			s := FormatFloat(buf, float64(f32), 'e', -1, 32)
			if x, _ := ParseFloat(s, 32); float32(x) != f32 {
				t.Errorf("failed roundtrip %v => %s => %v", f32, s, float32(x))
			}
		}
	}
}

func TestFormatFloatInvalidBitSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to invalid bitSize")
		}
	}()
	buf := make([]byte, 64)
	_ = FormatFloat(buf, 3.14, 'g', -1, 100)
}
