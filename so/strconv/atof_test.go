// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

import (
	"math"
	stdconv "strconv"
	"testing"
)

// sameFloat reports whether two float64 values hold the same bits. Two NaN
// values count as the same, whatever their payload.
func sameFloat(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Float64bits(a) == math.Float64bits(b)
}

func FuzzParseFloat(f *testing.F) {
	// Compare ParseFloat with the strconv package.
	f.Add("", 64)
	f.Add("0", 64)
	f.Add("-0", 64)
	f.Add("+1", 64)
	f.Add("1e23", 64)
	f.Add("1e-320", 64)
	f.Add("5e-324", 64)
	f.Add("2e-324", 64)
	f.Add("1.7976931348623157e308", 64)
	f.Add("1.7976931348623159e308", 64)
	f.Add("3.4028236e38", 32)
	f.Add("1e-45", 32)
	f.Add("0x1p-2", 64)
	f.Add("0x1.fffffffffffffp1023", 64)
	f.Add("0x1_ep-1", 64)
	f.Add("0x1e2", 64)
	f.Add("nan", 64)
	f.Add("-Infinity", 64)
	f.Add("+INF", 32)
	f.Add("1_23.50_0_0e+1_2", 64)
	f.Add("123.5e+12_", 64)
	f.Add("1e+18446744073709551616", 64)
	f.Add("2.2250738585072012e-308", 64)
	f.Add("1090544144181609348671888949248", 64)

	f.Fuzz(func(t *testing.T, s string, bitSize int) {
		// The two packages read any other bit size as 64, so a fuzzed value
		// would add no coverage.
		if bitSize != 32 {
			bitSize = 64
		}
		got, gotErr := ParseFloat(s, bitSize)
		want, wantErr := stdconv.ParseFloat(s, bitSize)
		if !sameFloat(got, want) || errKind(gotErr) != errKind(wantErr) {
			t.Fatalf("ParseFloat(%q, %d) = %v, %v; want %v, %v",
				s, bitSize, got, gotErr, want, wantErr)
		}
	})
}
