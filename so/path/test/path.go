// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import (
	"solod.dev/so/mem"
	"solod.dev/so/path"
	"solod.dev/so/testing"
)

// A pathCase is a test case of a function that takes one path.
type pathCase struct {
	path, want string
}

var cleanCases = []pathCase{
	// Already clean
	{"", "."},
	{"abc", "abc"},
	{"abc/def", "abc/def"},
	{"a/b/c", "a/b/c"},
	{".", "."},
	{"..", ".."},
	{"../..", "../.."},
	{"../../abc", "../../abc"},
	{"/abc", "/abc"},
	{"/", "/"},

	// Remove trailing slash
	{"abc/", "abc"},
	{"abc/def/", "abc/def"},
	{"a/b/c/", "a/b/c"},
	{"./", "."},
	{"../", ".."},
	{"../../", "../.."},
	{"/abc/", "/abc"},

	// Remove doubled slash
	{"abc//def//ghi", "abc/def/ghi"},
	{"//abc", "/abc"},
	{"///abc", "/abc"},
	{"//abc//", "/abc"},
	{"abc//", "abc"},

	// Remove . elements
	{"abc/./def", "abc/def"},
	{"/./abc/def", "/abc/def"},
	{"abc/.", "abc"},

	// Remove .. elements
	{"abc/def/ghi/../jkl", "abc/def/jkl"},
	{"abc/def/../ghi/../jkl", "abc/jkl"},
	{"abc/def/..", "abc"},
	{"abc/def/../..", "."},
	{"/abc/def/../..", "/"},
	{"abc/def/../../..", ".."},
	{"/abc/def/../../..", "/"},
	{"abc/def/../../../ghi/jkl/../../../mno", "../../mno"},

	// Combinations
	{"abc/./../def", "def"},
	{"abc//./../def", "def"},
	{"abc/../../././../def", "../../def"},

	// A name that only looks like a dot element
	{"...", "..."},
	{"a/.../b", "a/.../b"},
	{".x/..y", ".x/..y"},
	{"a/..b/../c", "a/c"},
}

func TestClean(t *testing.T) {
	alloc := t.Allocator()
	for _, tc := range cleanCases {
		got := path.Clean(alloc, tc.path)
		if got != tc.want {
			t.Errorf("Clean(%s) = %s, want %s", tc.path, got, tc.want)
		}
		mem.FreeString(alloc, got)

		// A clean path stays the same.
		got = path.Clean(alloc, tc.want)
		if got != tc.want {
			t.Errorf("Clean(%s) = %s, want %s", tc.want, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

// A splitCase is a test case of Split.
type splitCase struct {
	path, dir, file string
}

var splitCases = []splitCase{
	{"a/b", "a/", "b"},
	{"a/b/", "a/b/", ""},
	{"a/", "a/", ""},
	{"a", "", "a"},
	{"/", "/", ""},
	{"", "", ""},
	{"//", "//", ""},
	{"/a", "/", "a"},
	{"a//b", "a//", "b"},
}

func TestSplit(t *testing.T) {
	for _, tc := range splitCases {
		dir, file := path.Split(tc.path)
		if dir != tc.dir || file != tc.file {
			t.Errorf("Split(%s) = %s, %s, want %s, %s",
				tc.path, dir, file, tc.dir, tc.file)
		}
	}
}

// A joinCase is a test case of Join. The elements are joined with elemSep,
// because Solod has no nested composite literal in a table.
type joinCase struct {
	elems string // the elements, separated by elemSep
	n     int    // the number of the elements
	want  string
}

var joinCases = []joinCase{
	// zero elements
	{"", 0, ""},

	// one element
	{"", 1, ""},
	{"a", 1, "a"},
	{"a/b/../c", 1, "a/c"},

	// two elements
	{"a\x01b", 2, "a/b"},
	{"a\x01", 2, "a"},
	{"\x01b", 2, "b"},
	{"/\x01a", 2, "/a"},
	{"/\x01", 2, "/"},
	{"a/\x01b", 2, "a/b"},
	{"a/\x01", 2, "a"},
	{"\x01", 2, ""},
	{"a\x01../b", 2, "b"},
	{"a\x01/b", 2, "a/b"},

	// three elements
	{"a\x01\x01b", 3, "a/b"},
	{"a\x01b\x01c", 3, "a/b/c"},
	{"\x01\x01", 3, ""},
	{"a\x01..\x01..", 3, ".."},
}

func TestJoin(t *testing.T) {
	alloc := t.Allocator()
	var elems [maxElem]string
	for _, tc := range joinCases {
		for i := range tc.n {
			elems[i] = elemAt(tc.elems, i)
		}
		got := path.Join(alloc, elems[:tc.n]...)
		if got != tc.want {
			t.Errorf("Join(%s) = %s, want %s", tc.elems, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

var extCases = []pathCase{
	{"path.go", ".go"},
	{"path.pb.go", ".go"},
	{"a.dir/b", ""},
	{"a.dir/b.go", ".go"},
	{"a.dir/", ""},
	{"", ""},
	{".", "."},
	{"..", "."},
	{"a/.b", ".b"},
	{"a.b/c", ""},
	{"a/b.", "."},
}

func TestExt(t *testing.T) {
	for _, tc := range extCases {
		if got := path.Ext(tc.path); got != tc.want {
			t.Errorf("Ext(%s) = %s, want %s", tc.path, got, tc.want)
		}
	}
}

var baseCases = []pathCase{
	// Already clean
	{"", "."},
	{".", "."},
	{"/.", "."},
	{"/", "/"},
	{"////", "/"},
	{"x/", "x"},
	{"abc", "abc"},
	{"abc/def", "def"},
	{"a/b/.x", ".x"},
	{"a/b/c.", "c."},
	{"a/b/c.x", "c.x"},
	{"..", ".."},
	{"a/b/../", ".."},
	{"a//b//", "b"},
}

func TestBase(t *testing.T) {
	for _, tc := range baseCases {
		if got := path.Base(tc.path); got != tc.want {
			t.Errorf("Base(%s) = %s, want %s", tc.path, got, tc.want)
		}
	}
}

var dirCases = []pathCase{
	{"", "."},
	{".", "."},
	{"/.", "/"},
	{"/", "/"},
	{"////", "/"},
	{"/foo", "/"},
	{"x/", "x"},
	{"abc", "."},
	{"abc/def", "abc"},
	{"abc////def", "abc"},
	{"a/b/.x", "a/b"},
	{"a/b/c.", "a/b"},
	{"a/b/c.x", "a/b"},
	{"..", "."},
	{"../..", ".."},
	{"a/../b", "."},
	{"../a/b", "../a"},
}

func TestDir(t *testing.T) {
	alloc := t.Allocator()
	for _, tc := range dirCases {
		got := path.Dir(alloc, tc.path)
		if got != tc.want {
			t.Errorf("Dir(%s) = %s, want %s", tc.path, got, tc.want)
		}
		mem.FreeString(alloc, got)
	}
}

// An isAbsCase is a test case of IsAbs.
type isAbsCase struct {
	path string
	want bool
}

var isAbsCases = []isAbsCase{
	{"", false},
	{"/", true},
	{"/usr/bin/gcc", true},
	{"..", false},
	{"/a/../bb", true},
	{".", false},
	{"./", false},
	{"lala", false},
	{"//a", true},
}

func TestIsAbs(t *testing.T) {
	for _, tc := range isAbsCases {
		if got := path.IsAbs(tc.path); got != tc.want {
			t.Errorf("IsAbs(%s) = %t, want %t", tc.path, got, tc.want)
		}
	}
}
