// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

import (
	stdconv "strconv"
	"testing"
)

// floatFmts are the format bytes of the float fuzzer. The last byte is not a
// format, so the fuzzer also covers the text that an unknown format writes.
const floatFmts = "beEfgGxX?"

// maxFuzzPrec is the highest precision of the float fuzzer.
const maxFuzzPrec = 100

// floatBufLen is the length of the scratch buffer of the float fuzzer. It
// holds the longest text that FormatFloat writes at maxFuzzPrec.
const floatBufLen = 512

func FuzzFormatFloat(f *testing.F) {
	// Compare FormatFloat with the strconv package.
	f.Add(1.0, byte(0), -1, true)
	f.Add(0.0, byte(3), 5, true)
	f.Add(-1.0, byte(4), 5, true)
	f.Add(1e23, byte(2), 17, true)
	f.Add(1234567.8, byte(6), -1, true)
	f.Add(0.996644984, byte(3), 2, true)
	f.Add(2.275555555555555, byte(6), 21, true)
	f.Add(3.999969482421875, byte(6), 3, true)
	f.Add(2.2250738585072012e-308, byte(4), -1, true)
	f.Add(5.960464477539063e-08, byte(4), -1, false)
	f.Add(3.4028234663852886e38, byte(4), -1, false)
	f.Add(1e-45, byte(4), -1, false)
	f.Add(123.45, byte(8), 0, true)

	f.Fuzz(func(t *testing.T, val float64, fmtc byte, prec int, wide bool) {
		c := floatFmts[int(fmtc)%len(floatFmts)]
		prec = (prec&0x7f)%(maxFuzzPrec+2) - 1
		bitSize := 32
		if wide {
			bitSize = 64
		} else {
			val = float64(float32(val))
		}
		buf := make([]byte, floatBufLen)
		got := FormatFloat(buf, val, c, prec, bitSize)
		want := stdconv.FormatFloat(val, c, prec, bitSize)
		if got != want {
			t.Fatalf("FormatFloat(%v, %c, %d, %d) = %q, want %q",
				val, rune(c), prec, bitSize, got, want)
		}
		gotApp := AppendFloat([]byte("abc"), val, c, prec, bitSize)
		wantApp := stdconv.AppendFloat([]byte("abc"), val, c, prec, bitSize)
		if string(gotApp) != string(wantApp) {
			t.Fatalf("AppendFloat(%q, %v, %c, %d, %d) = %q, want %q",
				"abc", val, rune(c), prec, bitSize, gotApp, wantApp)
		}
	})
}

func TestFormatFloatInvalidBitSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to invalid bitSize")
		}
	}()
	buf := make([]byte, floatBufLen)
	_ = FormatFloat(buf, 3.14, 'g', -1, 100)
}
