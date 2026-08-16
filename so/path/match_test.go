// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	gopath "path"
	"testing"
)

// The patterns of the fuzz corpus. They hold every pattern shape: a star, a
// question mark, a character class, a negated class, a range, and an escape.
var fuzzPatterns = []string{
	"",
	"*",
	"**",
	"a*",
	"*c",
	"a*b?c*x",
	"a*b*c*d*e*/f",
	"ab[c]",
	"ab[b-d]",
	"ab[^e-g]",
	"[a-ζ]*",
	"[\\-x]",
	"[]a]",
	"[",
	"[^",
	"a[",
	"\\",
	"a\\*b",
	"/*/*",
}

// The names of the fuzz corpus.
var fuzzNames = []string{
	"",
	"a",
	"abc",
	"a/b",
	"ab/c",
	"axbxcxdxe/f",
	"a☺b",
	"α",
	"-",
	"]",
	"\x00\xff",
}

func FuzzMatch(f *testing.F) {
	for i, pattern := range fuzzPatterns {
		f.Add(pattern, fuzzNames[i%len(fuzzNames)])
	}
	f.Fuzz(func(t *testing.T, pattern, name string) {
		match, err := Match(pattern, name)
		wantMatch, wantErr := gopath.Match(pattern, name)
		if match != wantMatch {
			t.Errorf("Match(%q, %q) = %v, want %v", pattern, name, match, wantMatch)
		}
		if (err != nil) != (wantErr != nil) {
			t.Errorf("Match(%q, %q) = %v, want %v", pattern, name, err, wantErr)
		}
		if err != nil && err != ErrBadPattern {
			t.Errorf("Match(%q, %q) = %v, want ErrBadPattern", pattern, name, err)
		}
	})
}
