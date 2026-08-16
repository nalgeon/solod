// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex_test

import (
	"solod.dev/so/bytes"
	"solod.dev/so/encoding/hex"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// maxMul is the largest repeat count of the encoder and decoder test. The
// count 192 gives an input above the 512 byte chunk of the encoder.
const maxMul = 192

// The buffers of the encoder and decoder test. The test writes into them
// through a bytes.Buffer with the mem.NoAlloc allocator, so the sizes bound
// the whole test.
var (
	inArr  [maxMul * maxDec]byte
	outArr [2 * maxMul * maxDec]byte
	encArr [2 * maxMul * maxDec]byte
	decArr [maxMul * maxDec]byte
)

// repeatInto writes count copies of s into buf and returns the written part.
func repeatInto(buf []byte, s string, count int) []byte {
	for i := range count {
		copy(buf[i*len(s):], []byte(s))
	}
	return buf[:len(s)*count]
}

func TestEncoderDecoder(t *testing.T) {
	var copyBuf [7]byte
	muls := [3]int{1, 128, maxMul}
	for _, mul := range muls {
		for _, test := range encDecTests {
			input := repeatInto(inArr[:], test.dec, mul)
			output := string(repeatInto(outArr[:], test.enc, mul))

			// The encoder writes the hexadecimal form of the input.
			buf := bytes.NewBuffer(mem.NoAlloc, encArr[:0])
			enc := hex.NewEncoder(&buf)
			r := bytes.NewReader(input)
			n, err := io.CopyBuffer(&enc, &r, copyBuf[:])
			if n != int64(len(input)) || err != nil {
				t.Errorf("mul %d: CopyBuffer() = %d, %s, want %d, nil",
					mul, n, errName(errCode(err)), len(input))
				continue
			}
			if buf.String() != output {
				t.Errorf("mul %d: the encoder wrote %s, want %s", mul, buf.String(), output)
				continue
			}

			// The decoder gives the input back.
			decBuf := bytes.NewBuffer(mem.NoAlloc, decArr[:0])
			dec := hex.NewDecoder(&buf)
			if _, err := io.CopyBuffer(&decBuf, &dec, copyBuf[:]); err != nil {
				t.Errorf("mul %d: CopyBuffer() = %s, want nil", mul, errName(errCode(err)))
				continue
			}
			if !bytes.Equal(decBuf.Bytes(), input) {
				t.Errorf("mul %d: the decoder wrote %d bytes, want %d",
					mul, decBuf.Len(), len(input))
			}
		}
	}
}

func TestDecoderErr(t *testing.T) {
	alloc := t.Allocator()
	for _, test := range errTests {
		r := strings.NewReader(test.in)
		dec := hex.NewDecoder(&r)
		out, err := io.ReadAll(alloc, &dec)

		// The decoder reads from a stream, so it reports io.ErrUnexpectedEOF
		// instead of ErrLength.
		want := test.err
		if want == errLength {
			want = errUnexpectedEOF
		}
		if string(out) != test.out {
			t.Errorf("NewDecoder(%s) gave %s, want %s", test.in, string(out), test.out)
		}
		if code := errCode(err); code != want {
			t.Errorf("NewDecoder(%s) = %s, want %s", test.in, errName(code), errName(want))
		}
		mem.FreeSlice(alloc, out)
	}
}

func TestDecoderOneByte(t *testing.T) {
	// Check the buffer refill of the decoder. A reader that gives one byte
	// at a time leaves an odd byte in the buffer at every step.
	alloc := t.Allocator()
	for i, test := range encDecTests {
		r := oneByteReader{s: test.enc}
		dec := hex.NewDecoder(&r)
		out, err := io.ReadAll(alloc, &dec)
		if err != nil {
			t.Errorf("#%d: ReadAll() = %s, want nil", i, errName(errCode(err)))
			continue
		}
		if string(out) != test.dec {
			t.Errorf("#%d: ReadAll() = %s, want %s", i, string(out), test.dec)
		}
		mem.FreeSlice(alloc, out)
	}
}

func TestEncoderWriteErr(t *testing.T) {
	var w errWriter
	enc := hex.NewEncoder(&w)

	n, err := enc.Write([]byte("hello"))
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Write() = %d, %s, want 0, errWriteFailed", n, errName(errCode(err)))
	}
	// The encoder keeps the error, so the next write does nothing.
	n, err = enc.Write([]byte("world"))
	if n != 0 || errCode(err) != errWrite {
		t.Errorf("Write() = %d, %s, want 0, errWriteFailed", n, errName(errCode(err)))
	}
}

func TestEncoderEmpty(t *testing.T) {
	var out strings.Builder
	defer out.Free()
	enc := hex.NewEncoder(&out)

	n, err := enc.Write(nil)
	if n != 0 || err != nil {
		t.Errorf("Write() = %d, %s, want 0, nil", n, errName(errCode(err)))
	}
	if out.String() != "" {
		t.Errorf("the encoder wrote %s, want nothing", out.String())
	}
}

func TestDecoderShortBuf(t *testing.T) {
	r := strings.NewReader("48656c6c6f")
	dec := hex.NewDecoder(&r)

	var got [5]byte
	for i := range 5 {
		n, err := dec.Read(got[i : i+1])
		if n != 1 || err != nil {
			t.Errorf("Read() = %d, %s, want 1, nil", n, errName(errCode(err)))
			return
		}
	}
	if string(got[:]) != "Hello" {
		t.Errorf("the decoder gave %s, want Hello", string(got[:]))
	}
	n, err := dec.Read(got[:])
	if n != 0 || errCode(err) != errEOF {
		t.Errorf("Read() = %d, %s, want 0, EOF", n, errName(errCode(err)))
	}
}

func TestDecoderSweep(t *testing.T) {
	// Run the decoder over every short word of the sweep
	// alphabet and compare the result against Decode.
	var wbuf [maxDecWord]byte
	var want [maxDecWord / 2]byte
	var got [maxDecWord]byte

	words := wordTotal(decAlpha, maxDecWord)
	for i := range words {
		word := wordAt(wbuf[:], decAlpha, maxDecWord, i)
		wantN, wantErr := decodeBrute(want[:], word)
		// The decoder reads from a stream, so it reports io.ErrUnexpectedEOF
		// instead of ErrLength, and io.EOF at the end of a valid word.
		switch wantErr {
		case errLength:
			wantErr = errUnexpectedEOF
		case errNone:
			wantErr = errEOF
		}

		// A read of no bytes reports no error, so the loop reads until the
		// decoder reports one.
		r := strings.NewReader(word)
		dec := hex.NewDecoder(&r)
		gotN := 0
		var gotErr error
		for gotErr == nil {
			var n int
			n, gotErr = dec.Read(got[gotN:])
			gotN += n
		}
		if gotN != wantN || string(got[:gotN]) != string(want[:wantN]) {
			t.Errorf("word %d: the decoder gave %d bytes, want %d", i, gotN, wantN)
			return
		}
		if code := errCode(gotErr); code != wantErr {
			t.Errorf("word %d: Read() = %s, want %s", i, errName(code), errName(wantErr))
			return
		}
	}
}
