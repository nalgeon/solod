// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex_test

import (
	"solod.dev/so/encoding/hex"
	"solod.dev/so/testing"
)

func TestEncodeSweep(t *testing.T) {
	// Encode every byte value and compares the result against the reference encoder.
	var src [256]byte
	for i := range 256 {
		src[i] = byte(i)
	}
	var got, want [512]byte

	// Every prefix of the byte range, so the sweep covers every length too.
	for length := range len(src) + 1 {
		gotN := hex.Encode(got[:], src[:length])
		wantN := encodeBrute(want[:], src[:length])
		if gotN != wantN {
			t.Errorf("length %d: Encode() = %d, want %d", length, gotN, wantN)
			return
		}
		if string(got[:gotN]) != string(want[:wantN]) {
			t.Errorf("length %d: Encode() wrote %s, want %s",
				length, string(got[:gotN]), string(want[:wantN]))
			return
		}
	}
}

func TestPairSweep(t *testing.T) {
	// Encode and decode every byte pair. The decoded value must give the value back.
	var src, back [2]byte
	var enc [4]byte

	for i := range 65536 {
		src[0] = byte(i >> 8)
		src[1] = byte(i)

		hex.Encode(enc[:], src[:])
		n, err := hex.Decode(back[:], enc[:])
		if err != nil {
			t.Errorf("pair %x: Decode() = %s, want nil", i, errName(errCode(err)))
			return
		}
		if n != 2 || back[0] != src[0] || back[1] != src[1] {
			t.Errorf("pair %x: the round trip gave %x %x", i, back[0], back[1])
			return
		}
	}
}

func TestDecodePairSweep(t *testing.T) {
	// Decode every two character input and compare the
	// result against the reference decoder.
	var src [2]byte
	var got, want [1]byte

	for i := range 65536 {
		src[0] = byte(i >> 8)
		src[1] = byte(i)

		gotN, gotErr := hex.Decode(got[:], src[:])
		wantN, wantErr := decodeBrute(want[:], string(src[:]))
		if gotN != wantN || errCode(gotErr) != wantErr {
			t.Errorf("input %x: Decode() = %d, %s, want %d, %s",
				i, gotN, errName(errCode(gotErr)), wantN, errName(wantErr))
			return
		}
		if gotN == 1 && got[0] != want[0] {
			t.Errorf("input %x: Decode() wrote %x, want %x", i, got[0], want[0])
			return
		}
	}
}

func TestDecodeWordSweep(t *testing.T) {
	// Decode every short word of the sweep alphabet and compare the result
	// against the reference decoder. The words cover the odd lengths and
	// the invalid characters the pair sweep cannot reach.
	var wbuf [maxDecWord]byte
	var got, want [maxDecWord / 2]byte

	words := wordTotal(decAlpha, maxDecWord)
	for i := range words {
		word := wordAt(wbuf[:], decAlpha, maxDecWord, i)

		gotN, gotErr := hex.Decode(got[:], []byte(word))
		wantN, wantErr := decodeBrute(want[:], word)
		if gotN != wantN || errCode(gotErr) != wantErr {
			t.Errorf("word %d: Decode() = %d, %s, want %d, %s",
				i, gotN, errName(errCode(gotErr)), wantN, errName(wantErr))
			return
		}
		if string(got[:gotN]) != string(want[:wantN]) {
			t.Errorf("word %d: Decode() wrote %s, want %s",
				i, string(got[:gotN]), string(want[:wantN]))
			return
		}
	}
}

func TestCaseSweep(t *testing.T) {
	// Check that Decode accepts both letter cases and that the two
	// cases give the same value.
	var lower, upper [2]byte
	var gotLower, gotUpper [1]byte

	for i := range 256 {
		lower[0] = hexDigit(byte(i >> 4))
		lower[1] = hexDigit(byte(i) & 0x0f)
		upper[0] = upperHex(lower[0])
		upper[1] = upperHex(lower[1])

		if _, err := hex.Decode(gotLower[:], lower[:]); err != nil {
			t.Errorf("byte %x: Decode() = %s, want nil", i, errName(errCode(err)))
			return
		}
		if _, err := hex.Decode(gotUpper[:], upper[:]); err != nil {
			t.Errorf("byte %x: Decode() = %s, want nil", i, errName(errCode(err)))
			return
		}
		if gotLower[0] != byte(i) || gotUpper[0] != byte(i) {
			t.Errorf("byte %x: Decode() gave %x and %x", i, gotLower[0], gotUpper[0])
			return
		}
	}
}

// upperHex returns the uppercase form of a hexadecimal character.
func upperHex(c byte) byte {
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 'A'
	}
	return c
}
