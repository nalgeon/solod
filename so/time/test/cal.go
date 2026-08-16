// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// The bounds of the day by day calendar sweep.
const (
	sweepStartYear = -1000
	sweepDays      = 1_100_000 // about 3011 years, up to year 2011
)

// The bounds of the year by year sweep.
const (
	yearSweepFirst = -9999
	yearSweepLast  = 9999
)

// maxErrors is the number of errors a sweep reports before it gives up.
const maxErrors = 20

func TestCalendarSweep(t *testing.T) {
	// Walk the calendar one day at a time and compare the date, the day
	// of the year and the weekday against counters the test keeps itself.
	tm := time.Date(sweepStartYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	wantYear := sweepStartYear
	wantMonth := time.January
	wantDay := 1
	wantYday := 1
	wantWeekday := tm.Weekday()
	bad := 0

	for range sweepDays {
		date := tm.Date(time.UTC)
		if date.Year != wantYear || date.Month != wantMonth || date.Day != wantDay {
			t.Errorf("Date() = %d-%02d-%02d, want %d-%02d-%02d",
				date.Year, int(date.Month), date.Day,
				wantYear, int(wantMonth), wantDay)
			bad++
		}
		if tm.Year() != wantYear || tm.Month() != wantMonth || tm.Day() != wantDay {
			t.Errorf("Year(), Month(), Day() = %d-%02d-%02d, want %d-%02d-%02d",
				tm.Year(), int(tm.Month()), tm.Day(),
				wantYear, int(wantMonth), wantDay)
			bad++
		}
		if yday := tm.YearDay(); yday != wantYday {
			t.Errorf("Date(%d-%02d-%02d).YearDay() = %d, want %d",
				wantYear, int(wantMonth), wantDay, yday, wantYday)
			bad++
		}
		if wday := tm.Weekday(); wday != wantWeekday {
			t.Errorf("Date(%d-%02d-%02d).Weekday() = %d, want %d",
				wantYear, int(wantMonth), wantDay, int(wday), int(wantWeekday))
			bad++
		}
		if bad >= maxErrors {
			t.Fatal("too many errors")
			return
		}

		tm = tm.Add(24 * time.Hour)
		wantWeekday = (wantWeekday + 1) % 7
		wantYday++
		wantDay++
		if wantDay > daysInMonth(wantYear, wantMonth) {
			wantDay = 1
			wantMonth++
		}
		if wantMonth > time.December {
			wantMonth = time.January
			wantYday = 1
			wantYear++
		}
	}
}

func TestYearLengthSweep(t *testing.T) {
	// Check the distance between two consecutive January 1st
	// against the length of the year between them.
	prev := time.Date(yearSweepFirst, time.January, 1, 0, 0, 0, 0, time.UTC)
	bad := 0
	for year := yearSweepFirst; year < yearSweepLast; year++ {
		next := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
		want := int64(daysInYear(year)) * secondsPerDay
		if got := next.Unix() - prev.Unix(); got != want {
			t.Errorf("Date(%d-01-01) - Date(%d-01-01) = %d seconds, want %d",
				year+1, year, got, want)
			if bad++; bad >= maxErrors {
				t.Fatal("too many errors")
				return
			}
		}
		if yday := next.YearDay(); yday != 1 {
			t.Errorf("Date(%d-01-01).YearDay() = %d, want 1", year+1, yday)
			if bad++; bad >= maxErrors {
				t.Fatal("too many errors")
				return
			}
		}
		prev = next
	}
}

func TestMonthLengthSweep(t *testing.T) {
	// Check the distance between the first day of two consecutive months
	// against the length of the month between them.
	bad := 0
	for year := 1800; year <= 2200; year++ {
		for month := time.January; month <= time.December; month++ {
			first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
			// Day 1 of the next month, written as the day after the last day
			// of this month, to make Date normalize it.
			next := time.Date(year, month, daysInMonth(year, month)+1, 0, 0, 0, 0, time.UTC)
			want := int64(daysInMonth(year, month)) * secondsPerDay
			if got := next.Unix() - first.Unix(); got != want {
				t.Errorf("Date(%d-%02d) is %d seconds long, want %d",
					year, int(month), got, want)
				if bad++; bad >= maxErrors {
					t.Fatal("too many errors")
					return
				}
			}
			date := next.Date(time.UTC)
			wantMonth := month + 1
			wantYear := year
			if wantMonth > time.December {
				wantMonth = time.January
				wantYear++
			}
			if date.Year != wantYear || date.Month != wantMonth || date.Day != 1 {
				t.Errorf("the day after %d-%02d-%02d is %d-%02d-%02d, want %d-%02d-01",
					year, int(month), daysInMonth(year, month),
					date.Year, int(date.Month), date.Day,
					wantYear, int(wantMonth))
				if bad++; bad >= maxErrors {
					t.Fatal("too many errors")
					return
				}
			}
		}
	}
}
