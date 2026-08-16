// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bufio

import (
	stdbufio "bufio"
	stdstrings "strings"
	"testing"

	"solod.dev/so/mem"
	"solod.dev/so/strings"
	"solod.dev/so/unicode/utf8"
)

// loopAtEOFSplit gives an empty token at the end of the input forever.
func loopAtEOFSplit(data []byte, atEOF bool) SplitResult {
	if len(data) > 0 {
		return SplitResult{Advance: 1, Token: data[:1], HasToken: true}
	}
	return SplitResult{Token: data, HasToken: true}
}

func TestDontLoopForever(t *testing.T) {
	sr := strings.NewReader("abc")
	s := NewScanner(nil, &sr)
	defer s.Free()
	s.Split(loopAtEOFSplit)

	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("should have panicked")
		}
		if msg, ok := err.(string); !ok || !strings.Contains(msg, "empty tokens") {
			panic(err)
		}
	}()

	for count := 0; s.Scan(); count++ {
		if count > 1000 {
			t.Fatal("looping")
		}
	}
	if s.Err() != nil {
		t.Fatal("after scan:", s.Err())
	}
}

// The split functions the scanner fuzzer selects from.
const (
	splitLines = iota
	splitWords
	splitRunes
	splitBytes
	splitCount
)

func FuzzScanner(f *testing.F) {
	// The size of the buffer decides when the scanner grows the buffer and
	// when a token is too long, so the fuzzer varies it.
	f.Add([]byte(""), uint8(splitLines), uint8(16))
	f.Add([]byte("first line\r\nsecond line\n\n"), uint8(splitLines), uint8(16))
	f.Add([]byte(" abc\tdef\nghi "), uint8(splitWords), uint8(16))
	f.Add([]byte("abc\xc2\xbc\x81日本語"), uint8(splitRunes), uint8(16))
	f.Add([]byte("a line that does not fit the buffer\n"), uint8(splitLines), uint8(4))

	f.Fuzz(func(t *testing.T, data []byte, split, bufSize uint8) {
		size := int(bufSize)%128 + utf8.UTFMax
		kind := int(split) % splitCount

		sr := strings.NewReader(string(data))
		s := NewScanner(mem.System, &sr)
		defer s.Free()
		refS := stdbufio.NewScanner(stdstrings.NewReader(string(data)))

		switch kind {
		case splitLines:
			s.Split(ScanLines)
			refS.Split(stdbufio.ScanLines)
		case splitWords:
			s.Split(ScanWords)
			refS.Split(stdbufio.ScanWords)
		case splitRunes:
			s.Split(ScanRunes)
			refS.Split(stdbufio.ScanRunes)
		case splitBytes:
			s.Split(ScanBytes)
			refS.Split(stdbufio.ScanBytes)
		}
		s.Buffer(make([]byte, 0, size), size)
		refS.Buffer(make([]byte, 0, size), size)

		for i := 0; ; i++ {
			ok := s.Scan()
			refOk := refS.Scan()
			if ok != refOk {
				t.Fatalf("#%d: Scan() = %v, want %v", i, ok, refOk)
			}
			if !ok {
				break
			}
			if s.Text() != refS.Text() {
				t.Fatalf("#%d: Text() = %q, want %q", i, s.Text(), refS.Text())
			}
		}
		if errKind(s.Err()) != errKind(refS.Err()) {
			t.Fatalf("Err() = %v, want %v", s.Err(), refS.Err())
		}
	})
}
