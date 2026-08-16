// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/path"
	"solod.dev/so/testing"
)

// arenaSize is the size of the sweep allocator. It holds the allocations of
// one sweep step with room to spare. A sweep resets the arena after every
// step, so the allocations of the steps do not add up.
const arenaSize = 256

func TestCleanSweep(t *testing.T) {
	var backing [arenaSize]byte
	arena := mem.NewArena(backing[:])
	var pbuf, wbuf [maxPathWord]byte

	words := wordTotal(pathAlpha, maxPathWord)
	for i := range words {
		p := wordAt(pbuf[:], pathAlpha, maxPathWord, i)
		want := cleanBrute(wbuf[:], p)

		got := path.Clean(&arena, p)
		if got != want {
			t.Errorf("Clean(%s) = %s, want %s", p, got, want)
			return
		}
		// A clean path stays the same.
		again := path.Clean(&arena, got)
		if again != want {
			t.Errorf("Clean(%s) = %s, want %s", got, again, want)
			return
		}
		// Clean keeps a rooted path rooted and a relative path relative.
		if path.IsAbs(got) != path.IsAbs(p) {
			t.Errorf("IsAbs(Clean(%s)) = %t, want %t",
				p, path.IsAbs(got), path.IsAbs(p))
			return
		}
		// Join of one element cleans the element.
		if p != "" {
			joined := path.Join(&arena, p)
			if joined != want {
				t.Errorf("Join(%s) = %s, want %s", p, joined, want)
				return
			}
		}
		arena.Reset()
	}
}

func TestPathSweep(t *testing.T) {
	var backing [arenaSize]byte
	arena := mem.NewArena(backing[:])
	var pbuf, wbuf [maxPathWord]byte

	words := wordTotal(pathAlpha, maxPathWord)
	for i := range words {
		p := wordAt(pbuf[:], pathAlpha, maxPathWord, i)

		// The two parts of Split join back into the path.
		dir, file := path.Split(p)
		if len(dir)+len(file) != len(p) || p[:len(dir)] != dir || p[len(dir):] != file {
			t.Errorf("Split(%s) = %s, %s, want the two parts of the path",
				p, dir, file)
			return
		}
		// The directory ends in a slash, and the file holds none.
		if dir != "" && dir[len(dir)-1] != '/' {
			t.Errorf("Split(%s) = %s, %s: the directory does not end in a slash",
				p, dir, file)
			return
		}
		for j := 0; j < len(file); j++ {
			if file[j] == '/' {
				t.Errorf("Split(%s) = %s, %s: the file holds a slash",
					p, dir, file)
				return
			}
		}

		if got := path.Base(p); got != baseBrute(p) {
			t.Errorf("Base(%s) = %s, want %s", p, got, baseBrute(p))
			return
		}
		if got := path.Ext(p); got != extBrute(p) {
			t.Errorf("Ext(%s) = %s, want %s", p, got, extBrute(p))
			return
		}

		want := dirBrute(wbuf[:], p)
		got := path.Dir(&arena, p)
		if got != want {
			t.Errorf("Dir(%s) = %s, want %s", p, got, want)
			return
		}
		arena.Reset()
	}
}

func TestJoinSweep(t *testing.T) {
	var backing [arenaSize]byte
	arena := mem.NewArena(backing[:])
	var xbuf, ybuf [maxJoinWord]byte
	var jbuf, wbuf [2*maxJoinWord + 1]byte
	var elems [2]string

	words := wordTotal(pathAlpha, maxJoinWord)
	for i := range words {
		x := wordAt(xbuf[:], pathAlpha, maxJoinWord, i)
		for j := range words {
			y := wordAt(ybuf[:], pathAlpha, maxJoinWord, j)
			elems[0] = x
			elems[1] = y

			want := joinBrute(wbuf[:], jbuf[:], elems[:])
			got := path.Join(&arena, elems[:]...)
			if got != want {
				t.Errorf("Join(%s, %s) = %s, want %s", x, y, got, want)
				return
			}
			arena.Reset()
		}
	}
}

func TestMatchSweep(t *testing.T) {
	var pbuf [maxPattern]byte
	var nbuf [maxName]byte

	patterns := wordTotal(patternAlpha, maxPattern)
	names := wordTotal(nameAlpha, maxName)
	for i := range patterns {
		pattern := wordAt(pbuf[:], patternAlpha, maxPattern, i)
		for j := range names {
			name := wordAt(nbuf[:], nameAlpha, maxName, j)

			match, err := path.Match(pattern, name)
			if err != nil {
				t.Errorf("Match(%s, %s) = %s, want no error",
					pattern, name, err.Error())
				return
			}
			if want := matchBrute(pattern, name); match != want {
				t.Errorf("Match(%s, %s) = %t, want %t",
					pattern, name, match, want)
				return
			}
		}
	}
}
