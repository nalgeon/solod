// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	gopath "path"
	"testing"

	"solod.dev/so/mem"
)

// The paths of the fuzz corpus. They hold every element shape: a name, a dot
// element, a parent element, an empty element, and a leading slash.
var fuzzPaths = []string{
	"",
	".",
	"..",
	"/",
	"//",
	"abc",
	"abc/def",
	"/abc/",
	"abc//def//ghi",
	"abc/./def",
	"abc/def/../ghi",
	"abc/def/../../../ghi/jkl/../../../mno",
	"a/../..",
	"...",
	".x/..y",
	"a.dir/b.go",
	"/usr/bin/gcc",
	"\x00/\xff",
	"日本語/ファイル.txt",
}

func FuzzClean(f *testing.F) {
	for _, p := range fuzzPaths {
		f.Add(p)
	}
	f.Fuzz(func(t *testing.T, p string) {
		clean := Clean(nil, p)
		if want := gopath.Clean(p); clean != want {
			t.Errorf("Clean(%q) = %q, want %q", p, clean, want)
		}
		mem.FreeString(nil, clean)

		dir := Dir(nil, p)
		if want := gopath.Dir(p); dir != want {
			t.Errorf("Dir(%q) = %q, want %q", p, dir, want)
		}
		mem.FreeString(nil, dir)

		if got, want := Base(p), gopath.Base(p); got != want {
			t.Errorf("Base(%q) = %q, want %q", p, got, want)
		}
		if got, want := Ext(p), gopath.Ext(p); got != want {
			t.Errorf("Ext(%q) = %q, want %q", p, got, want)
		}
		if got, want := IsAbs(p), gopath.IsAbs(p); got != want {
			t.Errorf("IsAbs(%q) = %v, want %v", p, got, want)
		}

		gotDir, gotFile := Split(p)
		wantDir, wantFile := gopath.Split(p)
		if gotDir != wantDir || gotFile != wantFile {
			t.Errorf("Split(%q) = %q, %q, want %q, %q",
				p, gotDir, gotFile, wantDir, wantFile)
		}
	})
}

func FuzzJoin(f *testing.F) {
	for i, x := range fuzzPaths {
		f.Add(x, fuzzPaths[len(fuzzPaths)-1-i])
	}
	f.Fuzz(func(t *testing.T, x, y string) {
		joined := Join(nil, x, y)
		if want := gopath.Join(x, y); joined != want {
			t.Errorf("Join(%q, %q) = %q, want %q", x, y, joined, want)
		}
		mem.FreeString(nil, joined)
	})
}
