// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import (
	"solod.dev/so/path"
	"solod.dev/so/testing"
)

// A matchCase is a test case of Match. ErrBadPattern is the only error the
// function returns, so a flag is enough to name the wanted error.
type matchCase struct {
	pattern, s string
	match      bool // the wanted result
	bad        bool // the pattern is malformed
}

var matchCases = []matchCase{
	{"abc", "abc", true, false},
	{"*", "abc", true, false},
	{"*c", "abc", true, false},
	{"a*", "a", true, false},
	{"a*", "abc", true, false},
	{"a*", "ab/c", false, false},
	{"a*/b", "abc/b", true, false},
	{"a*/b", "a/c/b", false, false},
	{"a*b*c*d*e*/f", "axbxcxdxe/f", true, false},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/f", true, false},
	{"a*b*c*d*e*/f", "axbxcxdxe/xxx/f", false, false},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/fff", false, false},
	{"a*b?c*x", "abxbbxdbxebxczzx", true, false},
	{"a*b?c*x", "abxbbxdbxebxczzy", false, false},
	{"ab[c]", "abc", true, false},
	{"ab[b-d]", "abc", true, false},
	{"ab[e-g]", "abc", false, false},
	{"ab[^c]", "abc", false, false},
	{"ab[^b-d]", "abc", false, false},
	{"ab[^e-g]", "abc", true, false},
	{"a\\*b", "a*b", true, false},
	{"a\\*b", "ab", false, false},
	{"a?b", "a☺b", true, false},
	{"a[^a]b", "a☺b", true, false},
	{"a???b", "a☺b", false, false},
	{"a[^a][^a][^a]b", "a☺b", false, false},
	{"[a-ζ]*", "α", true, false},
	{"*[a-ζ]", "A", false, false},
	{"a?b", "a/b", false, false},
	{"a*b", "a/b", false, false},
	{"[\\]a]", "]", true, false},
	{"[\\-]", "-", true, false},
	{"[x\\-]", "x", true, false},
	{"[x\\-]", "-", true, false},
	{"[x\\-]", "z", false, false},
	{"[\\-x]", "x", true, false},
	{"[\\-x]", "-", true, false},
	{"[\\-x]", "a", false, false},
	{"[]a]", "]", false, true},
	{"[-]", "-", false, true},
	{"[x-]", "x", false, true},
	{"[x-]", "-", false, true},
	{"[x-]", "z", false, true},
	{"[-x]", "x", false, true},
	{"[-x]", "-", false, true},
	{"[-x]", "a", false, true},
	{"\\", "a", false, true},
	{"[a-b-c]", "a", false, true},
	{"[", "a", false, true},
	{"[^", "a", false, true},
	{"[^bc", "a", false, true},
	{"a[", "a", false, true},
	{"a[", "ab", false, true},
	{"a[", "x", false, true},
	{"a/b[", "x", false, true},
	{"*x", "xxx", true, false},

	// An empty pattern matches an empty name only.
	{"", "", true, false},
	{"", "a", false, false},
	// A star matches an empty run of bytes.
	{"*", "", true, false},
	{"**", "abc", true, false},
	{"a**b", "ab", true, false},
	// A question mark matches one code point, not one byte.
	{"?", "☺", true, false},
	{"??", "☺", false, false},
	// A slash matches a slash only.
	{"*/*", "a/b", true, false},
	{"*/*", "a/b/c", false, false},
	{"/", "/", true, false},
	// A class matches one code point.
	{"[☺☻]", "☺", true, false},
	{"[a-c]", "b", true, false},
	{"[^a-c]", "d", true, false},
	{"[abc]d", "cd", true, false},
	// An escape holds the letter after it.
	{"\\[", "[", true, false},
	{"\\?", "?", true, false},
	{"a\\", "a", false, true},
}

func TestMatch(t *testing.T) {
	for _, tc := range matchCases {
		match, err := path.Match(tc.pattern, tc.s)
		if match != tc.match {
			t.Errorf("Match(%s, %s) = %t, want %t",
				tc.pattern, tc.s, match, tc.match)
		}
		if tc.bad && err != path.ErrBadPattern {
			t.Errorf("Match(%s, %s) = no error, want ErrBadPattern",
				tc.pattern, tc.s)
		}
		if !tc.bad && err != nil {
			t.Errorf("Match(%s, %s) = %s, want no error",
				tc.pattern, tc.s, err.Error())
		}
	}
}
