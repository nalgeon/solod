// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex_test

import (
	"solod.dev/so/encoding/hex"
	"solod.dev/so/mem"
	"solod.dev/so/testing"
)

func TestEncodedLen(t *testing.T) {
	for n := range 100 {
		if got := hex.EncodedLen(n); got != n*2 {
			t.Errorf("EncodedLen(%d) = %d, want %d", n, got, n*2)
			return
		}
	}
}

func TestDecodedLen(t *testing.T) {
	for n := range 100 {
		if got := hex.DecodedLen(n); got != n/2 {
			t.Errorf("DecodedLen(%d) = %d, want %d", n, got, n/2)
			return
		}
	}
}

func TestEncode(t *testing.T) {
	var dst [2 * maxDec]byte
	for i, test := range encDecTests {
		n := hex.Encode(dst[:], []byte(test.dec))
		if n != hex.EncodedLen(len(test.dec)) {
			t.Errorf("#%d: Encode() = %d, want %d", i, n, hex.EncodedLen(len(test.dec)))
			continue
		}
		if string(dst[:n]) != test.enc {
			t.Errorf("#%d: Encode() wrote %s, want %s", i, string(dst[:n]), test.enc)
		}
	}
}

func TestAppendEncode(t *testing.T) {
	alloc := t.Allocator()
	for i, test := range encDecTests {
		dst := mem.AllocSlice[byte](alloc, 4, 4)
		copy(dst, []byte("lead"))
		dst = hex.AppendEncode(alloc, dst, []byte(test.dec))
		if len(dst) != 4+len(test.enc) {
			t.Errorf("#%d: AppendEncode() gave %d bytes, want %d", i, len(dst), 4+len(test.enc))
			continue
		}
		if string(dst[:4]) != "lead" {
			t.Errorf("#%d: AppendEncode() changed the prefix to %s", i, string(dst[:4]))
		}
		if string(dst[4:]) != test.enc {
			t.Errorf("#%d: AppendEncode() appended %s, want %s", i, string(dst[4:]), test.enc)
		}
		mem.FreeSlice(alloc, dst)
	}
}

func TestEncodeToString(t *testing.T) {
	alloc := t.Allocator()
	for i, test := range encDecTests {
		s := hex.EncodeToString(alloc, []byte(test.dec))
		if s != test.enc {
			t.Errorf("#%d: EncodeToString() = %s, want %s", i, s, test.enc)
		}
		mem.FreeString(alloc, s)
	}
}

func TestDecode(t *testing.T) {
	var dst [maxDec]byte
	for i, test := range encDecTests {
		n, err := hex.Decode(dst[:], []byte(test.enc))
		if err != nil {
			t.Errorf("#%d: Decode() = %s, want nil", i, errName(errCode(err)))
			continue
		}
		if n != hex.DecodedLen(len(test.enc)) {
			t.Errorf("#%d: Decode() = %d, want %d", i, n, hex.DecodedLen(len(test.enc)))
			continue
		}
		if string(dst[:n]) != test.dec {
			t.Errorf("#%d: Decode() wrote %s, want %s", i, string(dst[:n]), test.dec)
		}
	}
}

func TestDecodeUpper(t *testing.T) {
	var dst [8]byte
	n, err := hex.Decode(dst[:], []byte("F8F9FAFBFCFDFEFF"))
	if err != nil {
		t.Errorf("Decode() = %s, want nil", errName(errCode(err)))
		return
	}
	if string(dst[:n]) != "\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff" {
		t.Error("Decode() wrote the wrong bytes for an uppercase input")
	}
}

func TestAppendDecode(t *testing.T) {
	alloc := t.Allocator()
	for i, test := range encDecTests {
		dst := mem.AllocSlice[byte](alloc, 4, 4)
		copy(dst, []byte("lead"))
		dst, err := hex.AppendDecode(alloc, dst, []byte(test.enc))
		if err != nil {
			t.Errorf("#%d: AppendDecode() = %s, want nil", i, errName(errCode(err)))
			continue
		}
		if string(dst[:4]) != "lead" {
			t.Errorf("#%d: AppendDecode() changed the prefix to %s", i, string(dst[:4]))
		}
		if string(dst[4:]) != test.dec {
			t.Errorf("#%d: AppendDecode() appended %s, want %s", i, string(dst[4:]), test.dec)
		}
		mem.FreeSlice(alloc, dst)
	}
}

func TestDecodeString(t *testing.T) {
	alloc := t.Allocator()
	for i, test := range encDecTests {
		dst, err := hex.DecodeString(alloc, test.enc)
		if err != nil {
			t.Errorf("#%d: DecodeString() = %s, want nil", i, errName(errCode(err)))
			continue
		}
		if string(dst) != test.dec {
			t.Errorf("#%d: DecodeString() = %s, want %s", i, string(dst), test.dec)
		}
		mem.FreeSlice(alloc, dst)
	}
}

func TestDecodeErr(t *testing.T) {
	var dst [16]byte
	for _, test := range errTests {
		n, err := hex.Decode(dst[:], []byte(test.in))
		if string(dst[:n]) != test.out {
			t.Errorf("Decode(%s) wrote %s, want %s", test.in, string(dst[:n]), test.out)
		}
		if code := errCode(err); code != test.err {
			t.Errorf("Decode(%s) = %s, want %s", test.in, errName(code), errName(test.err))
		}
	}
}

func TestDecodeStringErr(t *testing.T) {
	alloc := t.Allocator()
	for _, test := range errTests {
		out, err := hex.DecodeString(alloc, test.in)
		if string(out) != test.out {
			t.Errorf("DecodeString(%s) = %s, want %s", test.in, string(out), test.out)
		}
		if code := errCode(err); code != test.err {
			t.Errorf("DecodeString(%s) = %s, want %s", test.in, errName(code), errName(test.err))
		}
		mem.FreeSlice(alloc, out)
	}
}

func TestAppendDecodeErr(t *testing.T) {
	alloc := t.Allocator()
	for _, test := range errTests {
		dst := mem.AllocSlice[byte](alloc, 4, 4)
		copy(dst, []byte("lead"))
		dst, err := hex.AppendDecode(alloc, dst, []byte(test.in))
		if string(dst[:4]) != "lead" {
			t.Errorf("AppendDecode(%s) changed the prefix to %s", test.in, string(dst[:4]))
		}
		if string(dst[4:]) != test.out {
			t.Errorf("AppendDecode(%s) appended %s, want %s", test.in, string(dst[4:]), test.out)
		}
		if code := errCode(err); code != test.err {
			t.Errorf("AppendDecode(%s) = %s, want %s", test.in, errName(code), errName(test.err))
		}
		mem.FreeSlice(alloc, dst)
	}
}

func TestDecodeInPlace(t *testing.T) {
	// Check that Decode accepts a destination that overlaps the source.
	// The package writes the byte i after it reads the bytes 2i and 2i+1.
	var buf [16]byte
	copy(buf[:], []byte("0001020304050607"))
	n, err := hex.Decode(buf[:], buf[:])
	if err != nil {
		t.Errorf("Decode() = %s, want nil", errName(errCode(err)))
		return
	}
	if string(buf[:n]) != "\x00\x01\x02\x03\x04\x05\x06\x07" {
		t.Error("Decode() wrote the wrong bytes for an overlapping destination")
	}
}
