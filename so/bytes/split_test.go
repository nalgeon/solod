// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	stdbytes "bytes"
	"testing"
)

func FuzzSplit(f *testing.F) {
	// Compare Split, SplitN and Join with the bytes package.
	f.Add([]byte(""), []byte(""), -1)
	f.Add([]byte("abcd"), []byte(""), 2)
	f.Add([]byte("abcd"), []byte("a"), 0)
	f.Add([]byte("1,2,3,4"), []byte(","), -1)
	f.Add([]byte("1 2 3 4"), []byte(" "), 3)
	f.Add([]byte("1....2....3....4"), []byte("..."), -1)
	f.Add([]byte("☺☻☹"), []byte("☹"), -1)
	f.Add([]byte("\xff-\xff"), []byte(""), -1)

	f.Fuzz(func(t *testing.T, s, sep []byte, n int) {
		got := SplitN(nil, s, sep, n)
		want := stdbytes.SplitN(s, sep, n)
		if len(got) != len(want) {
			t.Fatalf("SplitN(%q, %q, %d) = %q, want %q", s, sep, n, got, want)
		}
		for i, p := range got {
			if !Equal(p, want[i]) {
				t.Fatalf("SplitN(%q, %q, %d)[%d] = %q, want %q", s, sep, n, i, p, want[i])
			}
		}

		parts := Split(nil, s, sep)
		wantAll := stdbytes.Split(s, sep)
		if len(parts) != len(wantAll) {
			t.Fatalf("Split(%q, %q) = %q, want %q", s, sep, parts, wantAll)
		}
		for i, p := range parts {
			if !Equal(p, wantAll[i]) {
				t.Fatalf("Split(%q, %q)[%d] = %q, want %q", s, sep, i, p, wantAll[i])
			}
		}

		joined := Join(nil, parts, sep)
		wantJoined := stdbytes.Join(wantAll, sep)
		if !Equal(joined, wantJoined) {
			t.Fatalf("Join(Split(%q, %q), %q) = %q, want %q", s, sep, sep, joined, wantJoined)
		}
	})
}
