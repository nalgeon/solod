// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	stdtime "time"
)

func FuzzArithm(f *testing.F) {
	// Compare Add, Sub, AddDate, Truncate and Round with the time package. The
	// two times are built from the same pair of numbers, in both orders.
	f.Add(int64(0), int64(0), int64(0), 0, 0, 0)
	f.Add(int64(1221681866), int64(0), int64(Second), 1, 1, 1)
	f.Add(int64(1221681866), int64(999999999), int64(-Second), -1, 2, 3)
	f.Add(int64(-62135596800), int64(0), int64(Hour), 0, 0, 0)
	f.Add(int64(1136214245), int64(0), int64(24*Hour), 0, 1, -1)
	f.Add(int64(0), int64(1), int64(minDuration), 0, 0, 0)
	f.Add(int64(0), int64(-1), int64(maxDuration), 0, 0, 0)
	f.Add(int64(1e18), int64(1e18), int64(1e18), 4000, 100, 1000)
	f.Add(int64(-1e18), int64(-1e18), int64(-1e18), -4000, -100, -1000)

	f.Fuzz(func(t *testing.T, sec, nsec, d int64, years, months, days int) {
		years = bound(years, 9999)
		months = bound(months, 12*9999)
		days = bound(days, 365*9999)

		got := Unix(sec, nsec)
		want := stdtime.Unix(sec, nsec).UTC()
		gotDur := Duration(d)
		wantDur := stdtime.Duration(d)

		checkTime(t, "Add", got.Add(gotDur), want.Add(wantDur))
		checkTime(t, "Truncate", got.Truncate(gotDur), want.Truncate(wantDur))
		checkTime(t, "Round", got.Round(gotDur), want.Round(wantDur))
		checkTime(t, "AddDate", got.AddDate(years, months, days), want.AddDate(years, months, days))

		// The second time swaps the two numbers, so that Sub sees a distance
		// that is often far outside the range of a Duration.
		gotOther := Unix(nsec, sec)
		wantOther := stdtime.Unix(nsec, sec).UTC()
		if diff, wantDiff := got.Sub(gotOther), want.Sub(wantOther); int64(diff) != int64(wantDiff) {
			t.Errorf("Sub() = %d, want %d", int64(diff), int64(wantDiff))
		}
		if got.Before(gotOther) != want.Before(wantOther) {
			t.Errorf("Before() = %v, want %v", got.Before(gotOther), want.Before(wantOther))
		}
		if got.After(gotOther) != want.After(wantOther) {
			t.Errorf("After() = %v, want %v", got.After(gotOther), want.After(wantOther))
		}
		if got.Equal(gotOther) != want.Equal(wantOther) {
			t.Errorf("Equal() = %v, want %v", got.Equal(gotOther), want.Equal(wantOther))
		}
		if got.Compare(gotOther) != want.Compare(wantOther) {
			t.Errorf("Compare() = %d, want %d", got.Compare(gotOther), want.Compare(wantOther))
		}
	})
}
