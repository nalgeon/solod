// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	stdtime "time"
)

// bound maps an arbitrary value into the range [-lim, lim]. The fuzzer draws
// the whole int range, and a year of 10^18 tells nothing the calendar tests do
// not already tell.
func bound(v, lim int) int {
	return v % (lim + 1)
}

// checkTime compares every value the two packages report for one instant. want
// must be a Time in UTC, because the ported package reports UTC.
func checkTime(t *testing.T, ctx string, got Time, want stdtime.Time) {
	t.Helper()
	if got.Unix() != want.Unix() {
		t.Errorf("%s: Unix() = %d, want %d", ctx, got.Unix(), want.Unix())
	}
	if got.UnixMilli() != want.UnixMilli() {
		t.Errorf("%s: UnixMilli() = %d, want %d", ctx, got.UnixMilli(), want.UnixMilli())
	}
	if got.UnixMicro() != want.UnixMicro() {
		t.Errorf("%s: UnixMicro() = %d, want %d", ctx, got.UnixMicro(), want.UnixMicro())
	}
	if got.UnixNano() != want.UnixNano() {
		t.Errorf("%s: UnixNano() = %d, want %d", ctx, got.UnixNano(), want.UnixNano())
	}
	if got.IsZero() != want.IsZero() {
		t.Errorf("%s: IsZero() = %v, want %v", ctx, got.IsZero(), want.IsZero())
	}

	wantYear, wantMonth, wantDay := want.Date()
	if got.Year() != wantYear || int(got.Month()) != int(wantMonth) || got.Day() != wantDay {
		t.Errorf("%s: Year(), Month(), Day() = %d-%02d-%02d, want %d-%02d-%02d",
			ctx, got.Year(), int(got.Month()), got.Day(), wantYear, int(wantMonth), wantDay)
	}
	if date := got.Date(UTC); date.Year != wantYear || int(date.Month) != int(wantMonth) || date.Day != wantDay {
		t.Errorf("%s: Date() = %d-%02d-%02d, want %d-%02d-%02d",
			ctx, date.Year, int(date.Month), date.Day, wantYear, int(wantMonth), wantDay)
	}

	wantHour, wantMin, wantSec := want.Clock()
	if got.Hour() != wantHour || got.Minute() != wantMin || got.Second() != wantSec {
		t.Errorf("%s: Hour(), Minute(), Second() = %02d:%02d:%02d, want %02d:%02d:%02d",
			ctx, got.Hour(), got.Minute(), got.Second(), wantHour, wantMin, wantSec)
	}
	if clock := got.Clock(UTC); clock.Hour != wantHour || clock.Minute != wantMin || clock.Second != wantSec {
		t.Errorf("%s: Clock() = %02d:%02d:%02d, want %02d:%02d:%02d",
			ctx, clock.Hour, clock.Minute, clock.Second, wantHour, wantMin, wantSec)
	}
	if got.Nanosecond() != want.Nanosecond() {
		t.Errorf("%s: Nanosecond() = %d, want %d", ctx, got.Nanosecond(), want.Nanosecond())
	}

	if int(got.Weekday()) != int(want.Weekday()) {
		t.Errorf("%s: Weekday() = %d, want %d", ctx, int(got.Weekday()), int(want.Weekday()))
	}
	if got.YearDay() != want.YearDay() {
		t.Errorf("%s: YearDay() = %d, want %d", ctx, got.YearDay(), want.YearDay())
	}
	gotISOYear, gotISOWeek := got.ISOWeek()
	wantISOYear, wantISOWeek := want.ISOWeek()
	if gotISOYear != wantISOYear || gotISOWeek != wantISOWeek {
		t.Errorf("%s: ISOWeek() = %d/%d, want %d/%d",
			ctx, gotISOYear, gotISOWeek, wantISOYear, wantISOWeek)
	}
}

func FuzzDate(f *testing.F) {
	// Compare Date and the calendar methods with the time package. Every
	// component may be outside its usual range, so this also compares the
	// normalization of the two packages.
	f.Add(1, 1, 1, 0, 0, 0, 0, 0)
	f.Add(1970, 1, 1, 0, 0, 0, 0, 0)
	f.Add(2011, 11, 18, 15, 56, 35, 0, 0)
	f.Add(2011, 11, 18, 15, 56, 34, 1_000_000_000, 0)
	f.Add(2011, 11, 19, -9, 56, 35, 0, 0)
	f.Add(2011, 10, 49, 15, 56, 35, 0, 0)
	f.Add(2012, -1, 18, 15, 56, 35, 0, 0)
	f.Add(2010, 23, 18, 15, 56, 35, 0, 0)
	f.Add(1970, 1, 15297, 15, 56, 35, 0, 0)
	f.Add(1970, 1, -25508, 8, 0, 0, 0, 0)
	f.Add(2011, 11, 18, 10, 56, 35, 0, -5*3600)
	f.Add(2011, 11, 19, 3, 56, 35, 0, 12*3600)
	f.Add(2024, 6, 15, 23, 30, 0, 0, 5*3600+1800)
	f.Add(-1000, 2, 29, 0, 0, 0, 0, 0)

	f.Fuzz(func(t *testing.T, year, month, day, hour, min, sec, nsec, offset int) {
		year = bound(year, 9999)
		month = bound(month, 30)
		day = bound(day, 400)
		hour = bound(hour, 100)
		min = bound(min, 200)
		sec = bound(sec, 200)
		nsec = bound(nsec, 2_000_000_000)
		offset = bound(offset, 18*3600)

		zone := stdtime.FixedZone("", offset)
		got := Date(year, Month(month), day, hour, min, sec, nsec, Offset(offset))
		want := stdtime.Date(year, stdtime.Month(month), day, hour, min, sec, nsec, zone)
		checkTime(t, "Date", got, want.UTC())

		// The calendar methods that take an offset must report the components
		// the call passed in, once normalized.
		wantYear, wantMonth, wantDay := want.Date()
		if date := got.Date(Offset(offset)); date.Year != wantYear ||
			int(date.Month) != int(wantMonth) || date.Day != wantDay {
			t.Errorf("Date(%d) = %d-%02d-%02d, want %d-%02d-%02d", offset,
				date.Year, int(date.Month), date.Day, wantYear, int(wantMonth), wantDay)
		}
		wantHour, wantMin, wantSec := want.Clock()
		if clock := got.Clock(Offset(offset)); clock.Hour != wantHour ||
			clock.Minute != wantMin || clock.Second != wantSec {
			t.Errorf("Clock(%d) = %02d:%02d:%02d, want %02d:%02d:%02d", offset,
				clock.Hour, clock.Minute, clock.Second, wantHour, wantMin, wantSec)
		}
	})
}
