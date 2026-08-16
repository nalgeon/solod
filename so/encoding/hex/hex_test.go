// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex

import (
	stdhex "encoding/hex"
	"testing"

	"solod.dev/so/bytes"
	"solod.dev/so/io"
	"solod.dev/so/mem"
)

// The kinds of a decode error.
const (
	kindNone = iota
	kindLength
	kindInvalidByte
	kindOther
)

// errKind returns the kind of an error of this package.
func errKind(err error) int {
	switch err {
	case nil:
		return kindNone
	case ErrLength:
		return kindLength
	case ErrInvalidByte:
		return kindInvalidByte
	}
	return kindOther
}

// stdErrKind returns the kind of an error of the Go encoding/hex package.
func stdErrKind(err error) int {
	if err == nil {
		return kindNone
	}
	if err == stdhex.ErrLength {
		return kindLength
	}
	if _, ok := err.(stdhex.InvalidByteError); ok {
		return kindInvalidByte
	}
	return kindOther
}

// hexSeeds are the encoded seeds of the fuzzers.
var hexSeeds = []string{
	"", "0", "00", "0001020304050607", "08090a0b0c0d0e0f",
	"f8f9fafbfcfdfeff", "F8F9FAFBFCFDFEFF", "67", "e3a1",
	"zd4aa", "d4aaz", "30313", "0g", "00gg", "0\x01", "ffeed",
	"48656c6c6f20476f7068657221",
}

// binSeeds are the decoded seeds of the fuzzers.
var binSeeds = [][]byte{
	nil, {}, {0}, {0xff}, {0, 1, 2, 3, 4, 5, 6, 7},
	[]byte("Hello Gopher!"),
	[]byte("Go is an open source programming language."),
}

func FuzzEncode(f *testing.F) {
	for _, seed := range binSeeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src []byte) {
		want := stdhex.EncodeToString(src)

		if got := EncodedLen(len(src)); got != len(want) {
			t.Fatalf("EncodedLen(%d) = %d, want %d", len(src), got, len(want))
		}

		dst := make([]byte, EncodedLen(len(src)))
		n := Encode(dst, src)
		if n != len(want) || string(dst[:n]) != want {
			t.Fatalf("Encode(%x) = %q, want %q", src, dst[:n], want)
		}

		got := EncodeToString(mem.System, src)
		if got != want {
			t.Fatalf("EncodeToString(%x) = %q, want %q", src, got, want)
		}
		mem.FreeString(mem.System, got)

		// The prefix must come from the allocator AppendEncode reallocates
		// with, so it is not a plain []byte("lead").
		lead := mem.AllocSlice[byte](mem.System, 4, 4)
		copy(lead, "lead")
		app := AppendEncode(mem.System, lead, src)
		if string(app) != "lead"+want {
			t.Fatalf("AppendEncode(%x) = %q, want %q", src, app, "lead"+want)
		}
		mem.FreeSlice(mem.System, app)

		// The encoded form decodes back into the input.
		back, err := DecodeString(mem.System, want)
		if err != nil {
			t.Fatalf("DecodeString(%q) = %v, want nil", want, err)
		}
		if !bytes.Equal(back, src) {
			t.Fatalf("the round trip of %x gave %x", src, back)
		}
		mem.FreeSlice(mem.System, back)
	})
}

func FuzzDecode(f *testing.F) {
	for _, seed := range hexSeeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		wantDst := make([]byte, stdhex.DecodedLen(len(s)))
		wantN, wantErr := stdhex.Decode(wantDst, []byte(s))
		want := string(wantDst[:wantN])
		wantKind := stdErrKind(wantErr)

		if got := DecodedLen(len(s)); got != stdhex.DecodedLen(len(s)) {
			t.Fatalf("DecodedLen(%d) = %d, want %d", len(s), got, stdhex.DecodedLen(len(s)))
		}

		dst := make([]byte, DecodedLen(len(s)))
		n, err := Decode(dst, []byte(s))
		if string(dst[:n]) != want || errKind(err) != wantKind {
			t.Fatalf("Decode(%q) = %q, %v, want %q, %v", s, dst[:n], err, want, wantErr)
		}

		got, err := DecodeString(mem.System, s)
		if string(got) != want || errKind(err) != wantKind {
			t.Fatalf("DecodeString(%q) = %q, %v, want %q, %v", s, got, err, want, wantErr)
		}
		mem.FreeSlice(mem.System, got)

		// The prefix must come from the allocator AppendDecode reallocates
		// with, so it is not a plain []byte("lead").
		lead := mem.AllocSlice[byte](mem.System, 4, 4)
		copy(lead, "lead")
		app, err := AppendDecode(mem.System, lead, []byte(s))
		if string(app) != "lead"+want || errKind(err) != wantKind {
			t.Fatalf("AppendDecode(%q) = %q, %v, want %q, %v", s, app, err, "lead"+want, wantErr)
		}
		mem.FreeSlice(mem.System, app)
	})
}

func FuzzDump(f *testing.F) {
	for _, seed := range binSeeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		want := stdhex.Dump(data)
		got := Dump(mem.System, data)
		if got != want {
			t.Fatalf("Dump(%x) = \n%s\nwant\n%s", data, got, want)
		}
		mem.FreeString(mem.System, got)
	})
}

func FuzzStream(f *testing.F) {
	// Compare the Encoder and the Decoder against the Go encoding/hex
	// package. The chunk of the encoder is 512 bytes, so a long input takes
	// the path a short one does not.
	for _, seed := range binSeeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		want := stdhex.EncodeToString(data)

		var buf bytes.Buffer
		defer buf.Free()
		enc := NewEncoder(&buf)
		n, err := enc.Write(data)
		if n != len(data) || err != nil {
			t.Fatalf("Encoder.Write(%x) = %d, %v, want %d, nil", data, n, err, len(data))
		}
		if buf.String() != want {
			t.Fatalf("the encoder wrote %s, want %s", buf.String(), want)
		}

		dec := NewDecoder(&buf)
		back, err := io.ReadAll(mem.System, &dec)
		if err != nil {
			t.Fatalf("Decoder read of %q = %v, want nil", want, err)
		}
		if !bytes.Equal(back, data) {
			t.Fatalf("the round trip of %x gave %x", data, back)
		}
		mem.FreeSlice(mem.System, back)
	})
}
