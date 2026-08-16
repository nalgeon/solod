// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

import (
	stdconv "strconv"
	"testing"
)

// intBufLen is the length of the scratch buffer of the integer fuzzers. It
// holds the longest text that FormatInt and FormatUint write.
const intBufLen = MaxIntBase2Len

// fuzzBase maps a fuzzed value to a base that the format functions accept.
func fuzzBase(base int) int {
	return 2 + (base&0x7f)%35
}

func FuzzFormatInt(f *testing.F) {
	// Compare FormatInt with the strconv package.
	f.Add(int64(0), 10)
	f.Add(int64(1), 10)
	f.Add(int64(-1), 10)
	f.Add(int64(1<<31-1), 10)
	f.Add(int64(-1<<31), 10)
	f.Add(int64(1<<63-1), 10)
	f.Add(int64(-1<<63), 10)
	f.Add(int64(-1<<63), 2)
	f.Add(int64(-0x123456789abcdef), 16)
	f.Add(int64(32544027072), 35)
	f.Add(int64(38493362624), 36)

	f.Fuzz(func(t *testing.T, i int64, base int) {
		base = fuzzBase(base)
		buf := make([]byte, intBufLen)
		got, want := FormatInt(buf, i, base), stdconv.FormatInt(i, base)
		if got != want {
			t.Fatalf("FormatInt(%d, %d) = %q, want %q", i, base, got, want)
		}
		gotApp := AppendInt([]byte("abc"), i, base)
		wantApp := stdconv.AppendInt([]byte("abc"), i, base)
		if string(gotApp) != string(wantApp) {
			t.Fatalf("AppendInt(%q, %d, %d) = %q, want %q", "abc", i, base, gotApp, wantApp)
		}
		if int64(int(i)) == i && base == 10 {
			if got, want := Itoa(buf, int(i)), stdconv.Itoa(int(i)); got != want {
				t.Fatalf("Itoa(%d) = %q, want %q", i, got, want)
			}
		}
	})
}

func FuzzFormatUint(f *testing.F) {
	// Compare FormatUint with the strconv package.
	f.Add(uint64(0), 10)
	f.Add(uint64(1), 10)
	f.Add(uint64(1<<63-1), 10)
	f.Add(uint64(1<<63), 10)
	f.Add(uint64(1<<64-1), 10)
	f.Add(uint64(1<<64-1), 2)
	f.Add(uint64(1<<64-1), 36)
	f.Add(uint64(0xdeadbeef), 16)

	f.Fuzz(func(t *testing.T, u uint64, base int) {
		base = fuzzBase(base)
		buf := make([]byte, intBufLen)
		got, want := FormatUint(buf, u, base), stdconv.FormatUint(u, base)
		if got != want {
			t.Fatalf("FormatUint(%d, %d) = %q, want %q", u, base, got, want)
		}
		gotApp := AppendUint([]byte("abc"), u, base)
		wantApp := stdconv.AppendUint([]byte("abc"), u, base)
		if string(gotApp) != string(wantApp) {
			t.Fatalf("AppendUint(%q, %d, %d) = %q, want %q", "abc", u, base, gotApp, wantApp)
		}
	})
}

func TestFormatUintIllegalBase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to illegal base")
		}
	}()
	buf := make([]byte, intBufLen)
	FormatUint(buf, 12345678, 1)
}
