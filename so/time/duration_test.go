// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	stdtime "time"
)

func FuzzDuration(f *testing.F) {
	// Compare the Duration methods with the time package. m is the argument of
	// Truncate and Round.
	f.Add(int64(0), int64(0))
	f.Add(int64(1), int64(1))
	f.Add(int64(-1), int64(1))
	f.Add(int64(Second), int64(Millisecond))
	f.Add(int64(Hour+7*Minute), int64(Minute))
	f.Add(int64(-(Hour + 7*Minute)), int64(Minute))
	f.Add(int64(minDuration), int64(Second))
	f.Add(int64(maxDuration), int64(Second))
	f.Add(int64(minDuration), int64(minDuration))
	f.Add(int64(maxDuration), int64(maxDuration))
	f.Add(int64(1_200_000_000), int64(1_000_000_000))
	f.Add(int64(1e9+1), int64(-1))

	f.Fuzz(func(t *testing.T, d, m int64) {
		got := Duration(d)
		want := stdtime.Duration(d)
		gotUnit := Duration(m)
		wantUnit := stdtime.Duration(m)

		buf := make([]byte, MaxDurationLen)
		if s := got.String(buf); s != want.String() {
			t.Errorf("Duration(%d).String() = %q, want %q", d, s, want.String())
		}
		if got.Nanoseconds() != want.Nanoseconds() {
			t.Errorf("Duration(%d).Nanoseconds() = %d, want %d", d, got.Nanoseconds(), want.Nanoseconds())
		}
		if got.Microseconds() != want.Microseconds() {
			t.Errorf("Duration(%d).Microseconds() = %d, want %d", d, got.Microseconds(), want.Microseconds())
		}
		if got.Milliseconds() != want.Milliseconds() {
			t.Errorf("Duration(%d).Milliseconds() = %d, want %d", d, got.Milliseconds(), want.Milliseconds())
		}
		if got.Seconds() != want.Seconds() {
			t.Errorf("Duration(%d).Seconds() = %v, want %v", d, got.Seconds(), want.Seconds())
		}
		if got.Minutes() != want.Minutes() {
			t.Errorf("Duration(%d).Minutes() = %v, want %v", d, got.Minutes(), want.Minutes())
		}
		if got.Hours() != want.Hours() {
			t.Errorf("Duration(%d).Hours() = %v, want %v", d, got.Hours(), want.Hours())
		}
		if int64(got.Abs()) != int64(want.Abs()) {
			t.Errorf("Duration(%d).Abs() = %d, want %d", d, int64(got.Abs()), int64(want.Abs()))
		}
		if int64(got.Truncate(gotUnit)) != int64(want.Truncate(wantUnit)) {
			t.Errorf("Duration(%d).Truncate(%d) = %d, want %d", d, m,
				int64(got.Truncate(gotUnit)), int64(want.Truncate(wantUnit)))
		}
		if int64(got.Round(gotUnit)) != int64(want.Round(wantUnit)) {
			t.Errorf("Duration(%d).Round(%d) = %d, want %d", d, m,
				int64(got.Round(gotUnit)), int64(want.Round(wantUnit)))
		}
	})
}
