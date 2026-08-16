// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"
)

func TestRepeatCatchesOverflow(t *testing.T) {
	// golang.org/issue/16237
	type testCase struct {
		s      string
		count  int
		errStr string
	}

	runTestCases := func(prefix string, tests []testCase) {
		for i, tt := range tests {
			errStr := repeat(tt.s, tt.count)
			if tt.errStr == "" {
				if errStr != "" {
					t.Errorf("#%d panicked %v", i, errStr)
				}
				continue
			}

			if errStr == "" || !Contains(errStr, tt.errStr) {
				t.Errorf("%s#%d got %q want %q", prefix, i, errStr, tt.errStr)
			}
		}
	}

	const maxInt = int(^uint(0) >> 1)

	runTestCases("", []testCase{
		0: {"--", -2147483647, "negative"},
		1: {"", maxInt, ""},
		2: {"-", 10, ""},
		3: {"gopher", 0, ""},
		4: {"-", -1, "negative"},
		5: {"--", -102, "negative"},
		6: {string(make([]byte, 255)), int((^uint(0))/255 + 1), "overflow"},
	})

	const is64Bit = 1<<(^uintptr(0)>>63)/2 != 0
	if !is64Bit {
		return
	}

	runTestCases("64-bit", []testCase{
		0: {"-", maxInt, "capacity overflow"},
	})
}

// repeat calls Repeat and returns the message of the panic,
// or an empty string.
func repeat(s string, count int) (err string) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v.Error()
			default:
				err = v.(string)
			}
		}
	}()

	Repeat(nil, s, count)

	return
}
