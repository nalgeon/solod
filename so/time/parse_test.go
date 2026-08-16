// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
)

// parseLayouts are the layouts Parse handles without strptime. The C library
// provides strptime, so a Go test cannot reach the general path.
var parseLayouts = []string{RFC3339, RFC3339Nano, DateTime, DateOnly, TimeOnly}

func FuzzParse(f *testing.F) {
	// Parse reads the input at fixed positions, so a short input must not make
	// it read past the end.
	f.Add("2006-01-02T15:04:05Z")
	f.Add("2006-01-02T15:04:05+07:00")
	f.Add("2006-01-02T15:04:05-07:00")
	f.Add("2006-01-02T15:04:05.123456789Z")
	f.Add("2006-01-02T15:04:05.123456789+07:00")
	f.Add("2006-01-02 15:04:05")
	f.Add("2006-01-02")
	f.Add("15:04:05")
	f.Add("")
	f.Add("Z")
	f.Add("2006-01-02T15:04:05+")
	f.Add("2006-01-02T15:04:05.123456789+")
	f.Add("9999-99-99T99:99:99Z")
	f.Add("0000-00-00T00:00:00Z")

	f.Fuzz(func(t *testing.T, value string) {
		for _, layout := range parseLayouts {
			tm, err := Parse(layout, value, UTC)
			if err == nil && tm.IsZero() && value == "" {
				t.Errorf("Parse(%q, %q) accepted an empty value", layout, value)
			}
		}
	})
}
