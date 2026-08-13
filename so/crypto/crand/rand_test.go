// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crand

import (
	"compress/flate"
	"testing"

	"solod.dev/so/bytes"
	"solod.dev/so/fmt"
)

func TestRead(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		testRead(t, Read)
	})
	t.Run("Reader.Read", func(t *testing.T) {
		testRead(t, Reader.Read)
	})
}

func testRead(t *testing.T, Read func([]byte) (int, error)) {
	var n int = 1e5
	b := make([]byte, n)
	n, err := Read(b)
	if n != len(b) || err != nil {
		t.Fatalf("Read(buf) = %d, %s", n, err)
	}

	var z bytes.Buffer
	f, _ := flate.NewWriter(&z, 5)
	f.Write(b)
	f.Close()
	if z.Len() < len(b)*99/100 {
		t.Fatalf("Compressed %d -> %d", len(b), z.Len())
	}
}

func TestReadByteValues(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		testReadByteValues(t, Read)
	})
	t.Run("Reader.Read", func(t *testing.T) {
		testReadByteValues(t, Reader.Read)
	})
}

func testReadByteValues(t *testing.T, Read func([]byte) (int, error)) {
	b := make([]byte, 1)
	v := make(map[byte]bool)
	for {
		n, err := Read(b)
		if n != 1 || err != nil {
			t.Fatalf("Read(b) = %d, %v", n, err)
		}
		v[b[0]] = true
		if len(v) == 256 {
			break
		}
	}
}

func TestReadEmpty(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		testReadEmpty(t, Read)
	})
	t.Run("Reader.Read", func(t *testing.T) {
		testReadEmpty(t, Reader.Read)
	})
}

func testReadEmpty(t *testing.T, Read func([]byte) (int, error)) {
	n, err := Read(make([]byte, 0))
	if n != 0 || err != nil {
		t.Fatalf("Read(make([]byte, 0)) = %d, %v", n, err)
	}
	n, err = Read(nil)
	if n != 0 || err != nil {
		t.Fatalf("Read(nil) = %d, %v", n, err)
	}
}

func TestText(t *testing.T) {
	set := make(map[string]struct{}) // hold every string produced
	var indexSet [26]map[rune]int    // hold every char produced at every position
	for i := range indexSet {
		indexSet[i] = make(map[rune]int)
	}

	// not getting a char in a position: (31/32)¹⁰⁰⁰ = 1.6e-14
	// test completion within 1000 rounds: (1-(31/32)¹⁰⁰⁰)²⁶ = 0.9999999999996
	// empirically, this should complete within 400 rounds = 0.999921
	rounds := 1000
	var done bool
	buf := make([]byte, 26)
	for range rounds {
		s := Text(buf)
		if len(s) != 26 {
			t.Errorf("len(Text()) = %d, want = 26", len(s))
		}
		for i, r := range s {
			if ('A' > r || r > 'Z') && ('2' > r || r > '7') {
				t.Errorf("Text()[%d] = %v, outside of base32 alphabet", i, r)
			}
		}
		if _, ok := set[s]; ok {
			t.Errorf("Text() = %s, duplicate of previously produced string", s)
		}
		set[s] = struct{}{}

		done = true
		for i, r := range s {
			indexSet[i][r]++
			if len(indexSet[i]) != 32 {
				done = false
			}
		}
		if done {
			break
		}
	}
	if !done {
		t.Errorf("failed to produce every char at every index after %d rounds", rounds)
		indexSetTable(t, indexSet)
	}
}

func indexSetTable(t *testing.T, indexSet [26]map[rune]int) {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	line := "   "
	buf := make([]byte, 16)
	for _, r := range alphabet {
		line += fmt.Sprintf(buf, " %3s", string(r))
	}
	t.Log(line)
	for i, set := range indexSet {
		line = fmt.Sprintf(buf, "%2d:", i)
		for _, r := range alphabet {
			line += fmt.Sprintf(buf, " %3d", set[r])
		}
		t.Log(line)
	}
}
