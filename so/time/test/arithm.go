// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// An offsetCase is a calendar time and the offset it is specified in.
type offsetCase struct {
	year, month, day     int
	hour, minute, second int
	offset               time.Offset
}

var offsetCases = []offsetCase{
	{2011, 11, 18, 15, 56, 35, time.Offset(1 * 3600)},    // UTC+1
	{2011, 11, 18, 10, 56, 35, time.Offset(-5 * 3600)},   // UTC-5
	{2011, 11, 19, 3, 56, 35, time.Offset(12 * 3600)},    // UTC+12
	{2011, 11, 18, 3, 56, 35, time.Offset(-12 * 3600)},   // UTC-12
	{2024, 6, 15, 23, 30, 0, time.Offset(5*3600 + 1800)}, // UTC+5:30
}

func TestDateOffset(t *testing.T) {
	for _, c := range offsetCases {
		tm := time.Date(c.year, time.Month(c.month), c.day, c.hour, c.minute, c.second, 0, c.offset)
		date := tm.Date(c.offset)
		clock := tm.Clock(c.offset)
		if date.Year != c.year || int(date.Month) != c.month || date.Day != c.day {
			t.Errorf("Date(%d-%02d-%02d, offset %d).Date() = %d-%02d-%02d",
				c.year, c.month, c.day, int(c.offset),
				date.Year, int(date.Month), date.Day)
		}
		if clock.Hour != c.hour || clock.Minute != c.minute || clock.Second != c.second {
			t.Errorf("Date(%02d:%02d:%02d, offset %d).Clock() = %02d:%02d:%02d",
				c.hour, c.minute, c.second, int(c.offset),
				clock.Hour, clock.Minute, clock.Second)
		}
	}
}

// An addDateCase is an argument list of AddDate. Every case moves
// Fri Nov 18 07:56:35 2011 UTC to Thu Mar 19 07:56:35 2016 UTC.
type addDateCase struct {
	years, months, days int
}

var addDateCases = []addDateCase{
	{4, 4, 1},
	{3, 16, 1},
	{3, 15, 30},
	{5, -6, -18 - 30 - 12},
}

func TestAddDate(t *testing.T) {
	t0 := time.Date(2011, 11, 18, 7, 56, 35, 0, time.UTC)
	t1 := time.Date(2016, 3, 19, 7, 56, 35, 0, time.UTC)
	for _, c := range addDateCases {
		got := t0.AddDate(c.years, c.months, c.days)
		if !got.Equal(t1) {
			t.Errorf("AddDate(%d, %d, %d) = %d, want %d",
				c.years, c.months, c.days, got.Unix(), t1.Unix())
		}
	}

	t2 := time.Date(1899, 12, 31, 0, 0, 0, 0, time.UTC)
	days := t2.Unix() / secondsPerDay
	t3 := time.Unix(0, 0).AddDate(0, 0, int(days))
	if !t2.Equal(t3) {
		t.Errorf("AddDate(0, 0, %d) = %d, want %d", days, t3.Unix(), t2.Unix())
	}
}

func TestAddToExactSecond(t *testing.T) {
	// Add an amount to the current time to round it up to the next exact second.
	// The nsec field must stay in the range [0, 999999999].
	t1 := time.Now()
	t2 := t1.Add(time.Second - time.Duration(t1.Nanosecond()))
	sec := (t1.Second() + 1) % 60
	if t2.Second() != sec || t2.Nanosecond() != 0 {
		t.Errorf("sec = %d, nsec = %d, want sec = %d, nsec = 0",
			t2.Second(), t2.Nanosecond(), sec)
	}
}

// A subCase is two times and the expected difference between them. A zero
// flag denotes the zero Time, which Date cannot build.
type subCase struct {
	tZero                         bool
	ty, tmo, td, th, tmi, ts, tns int
	uZero                         bool
	uy, umo, ud, uh, umi, us, uns int
	want                          time.Duration
}

var subCases = []subCase{
	{true, 0, 0, 0, 0, 0, 0, 0, true, 0, 0, 0, 0, 0, 0, 0, 0},
	{false, 2009, 11, 23, 0, 0, 0, 1, false, 2009, 11, 23, 0, 0, 0, 0, 1},
	{false, 2009, 11, 23, 0, 0, 0, 0, false, 2009, 11, 24, 0, 0, 0, 0, -24 * time.Hour},
	{false, 2009, 11, 24, 0, 0, 0, 0, false, 2009, 11, 23, 0, 0, 0, 0, 24 * time.Hour},
	{false, -2009, 11, 24, 0, 0, 0, 0, false, -2009, 11, 23, 0, 0, 0, 0, 24 * time.Hour},
	{true, 0, 0, 0, 0, 0, 0, 0, false, 2109, 11, 23, 0, 0, 0, 0, minDuration},
	{false, 2109, 11, 23, 0, 0, 0, 0, true, 0, 0, 0, 0, 0, 0, 0, maxDuration},
	{true, 0, 0, 0, 0, 0, 0, 0, false, -2109, 11, 23, 0, 0, 0, 0, maxDuration},
	{false, -2109, 11, 23, 0, 0, 0, 0, true, 0, 0, 0, 0, 0, 0, 0, minDuration},
	{false, 2290, 1, 1, 0, 0, 0, 0, false, 2000, 1, 1, 0, 0, 0, 0, 290*365*24*time.Hour + 71*24*time.Hour},
	{false, 2300, 1, 1, 0, 0, 0, 0, false, 2000, 1, 1, 0, 0, 0, 0, maxDuration},
	{false, 2000, 1, 1, 0, 0, 0, 0, false, 2290, 1, 1, 0, 0, 0, 0, -290*365*24*time.Hour - 71*24*time.Hour},
	{false, 2000, 1, 1, 0, 0, 0, 0, false, 2300, 1, 1, 0, 0, 0, 0, minDuration},
	{false, 2311, 11, 26, 2, 16, 47, 63535996, false, 2019, 8, 16, 2, 29, 30, 268436582, 9223372036795099414},
}

func TestSub(t *testing.T) {
	for i, c := range subCases {
		var tm, um time.Time
		if !c.tZero {
			tm = time.Date(c.ty, time.Month(c.tmo), c.td, c.th, c.tmi, c.ts, c.tns, time.UTC)
		}
		if !c.uZero {
			um = time.Date(c.uy, time.Month(c.umo), c.ud, c.uh, c.umi, c.us, c.uns, time.UTC)
		}
		if got := tm.Sub(um); got != c.want {
			t.Errorf("#%d: Sub() = %d, want %d", i, int64(got), int64(c.want))
		}
	}
}

// An isoWeekCase is a date and the ISO 8601 year and week it belongs to.
type isoWeekCase struct {
	year, month, day int
	wantYear         int
	wantWeek         int
}

var isoWeekCases = []isoWeekCase{
	{1981, 1, 1, 1981, 1}, {1982, 1, 1, 1981, 53}, {1983, 1, 1, 1982, 52},
	{1984, 1, 1, 1983, 52}, {1985, 1, 1, 1985, 1}, {1986, 1, 1, 1986, 1},
	{1987, 1, 1, 1987, 1}, {1988, 1, 1, 1987, 53}, {1989, 1, 1, 1988, 52},
	{1990, 1, 1, 1990, 1}, {1991, 1, 1, 1991, 1}, {1992, 1, 1, 1992, 1},
	{1993, 1, 1, 1992, 53}, {1994, 1, 1, 1993, 52}, {1995, 1, 2, 1995, 1},
	{1996, 1, 1, 1996, 1}, {1996, 1, 7, 1996, 1}, {1996, 1, 8, 1996, 2},
	{1997, 1, 1, 1997, 1}, {1998, 1, 1, 1998, 1}, {1999, 1, 1, 1998, 53},
	{2000, 1, 1, 1999, 52}, {2001, 1, 1, 2001, 1}, {2002, 1, 1, 2002, 1},
	{2003, 1, 1, 2003, 1}, {2004, 1, 1, 2004, 1}, {2005, 1, 1, 2004, 53},
	{2006, 1, 1, 2005, 52}, {2007, 1, 1, 2007, 1}, {2008, 1, 1, 2008, 1},
	{2009, 1, 1, 2009, 1}, {2010, 1, 1, 2009, 53}, {2010, 1, 1, 2009, 53},
	{2011, 1, 1, 2010, 52}, {2011, 1, 2, 2010, 52}, {2011, 1, 3, 2011, 1},
	{2011, 1, 4, 2011, 1}, {2011, 1, 5, 2011, 1}, {2011, 1, 6, 2011, 1},
	{2011, 1, 7, 2011, 1}, {2011, 1, 8, 2011, 1}, {2011, 1, 9, 2011, 1},
	{2011, 1, 10, 2011, 2}, {2011, 1, 11, 2011, 2}, {2011, 6, 12, 2011, 23},
	{2011, 6, 13, 2011, 24}, {2011, 12, 25, 2011, 51}, {2011, 12, 26, 2011, 52},
	{2011, 12, 27, 2011, 52}, {2011, 12, 28, 2011, 52}, {2011, 12, 29, 2011, 52},
	{2011, 12, 30, 2011, 52}, {2011, 12, 31, 2011, 52}, {1995, 1, 1, 1994, 52},
	{2012, 1, 1, 2011, 52}, {2012, 1, 2, 2012, 1}, {2012, 1, 8, 2012, 1},
	{2012, 1, 9, 2012, 2}, {2012, 12, 23, 2012, 51}, {2012, 12, 24, 2012, 52},
	{2012, 12, 30, 2012, 52}, {2012, 12, 31, 2013, 1}, {2013, 1, 1, 2013, 1},
	{2013, 1, 6, 2013, 1}, {2013, 1, 7, 2013, 2}, {2013, 12, 22, 2013, 51},
	{2013, 12, 23, 2013, 52}, {2013, 12, 29, 2013, 52}, {2013, 12, 30, 2014, 1},
	{2014, 1, 1, 2014, 1}, {2014, 1, 5, 2014, 1}, {2014, 1, 6, 2014, 2},
	{2015, 1, 1, 2015, 1}, {2016, 1, 1, 2015, 53}, {2017, 1, 1, 2016, 52},
	{2018, 1, 1, 2018, 1}, {2019, 1, 1, 2019, 1}, {2020, 1, 1, 2020, 1},
	{2021, 1, 1, 2020, 53}, {2022, 1, 1, 2021, 52}, {2023, 1, 1, 2022, 52},
	{2024, 1, 1, 2024, 1}, {2025, 1, 1, 2025, 1}, {2026, 1, 1, 2026, 1},
	{2027, 1, 1, 2026, 53}, {2028, 1, 1, 2027, 52}, {2029, 1, 1, 2029, 1},
	{2030, 1, 1, 2030, 1}, {2031, 1, 1, 2031, 1}, {2032, 1, 1, 2032, 1},
	{2033, 1, 1, 2032, 53}, {2034, 1, 1, 2033, 52}, {2035, 1, 1, 2035, 1},
	{2036, 1, 1, 2036, 1}, {2037, 1, 1, 2037, 1}, {2038, 1, 1, 2037, 53},
	{2039, 1, 1, 2038, 52}, {2040, 1, 1, 2039, 52},
}

func TestISOWeek(t *testing.T) {
	// Selected dates and corner cases.
	for _, c := range isoWeekCases {
		tm := time.Date(c.year, time.Month(c.month), c.day, 0, 0, 0, 0, time.UTC)
		y, w := tm.ISOWeek()
		if y != c.wantYear || w != c.wantWeek {
			t.Errorf("Date(%d-%02d-%02d).ISOWeek() = %d/%d, want %d/%d",
				c.year, c.month, c.day, y, w, c.wantYear, c.wantWeek)
		}
	}

	// The only real invariant: Jan 04 is in week 1.
	for year := 1950; year < 2100; year++ {
		y, w := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC).ISOWeek()
		if y != year || w != 1 {
			t.Errorf("Date(%d-01-04).ISOWeek() = %d/%d, want %d/1", year, y, w, year)
		}
	}
}

// A yearDayCase is a date and the day of the year it falls on.
type yearDayCase struct {
	year, month, day int
	yday             int
}

var yearDayCases = []yearDayCase{
	// Common year.
	{2007, 1, 1, 1},
	{2007, 1, 15, 15},
	{2007, 2, 1, 32},
	{2007, 2, 15, 46},
	{2007, 3, 1, 60},
	{2007, 3, 15, 74},
	{2007, 4, 1, 91},
	{2007, 12, 31, 365},

	// Leap year.
	{2008, 1, 1, 1},
	{2008, 1, 15, 15},
	{2008, 2, 1, 32},
	{2008, 2, 15, 46},
	{2008, 3, 1, 61},
	{2008, 3, 15, 75},
	{2008, 4, 1, 92},
	{2008, 12, 31, 366},

	// Looks like a leap year, but is not.
	{1900, 1, 1, 1},
	{1900, 1, 15, 15},
	{1900, 2, 1, 32},
	{1900, 2, 15, 46},
	{1900, 3, 1, 60},
	{1900, 3, 15, 74},
	{1900, 4, 1, 91},
	{1900, 12, 31, 365},

	// Year one, a common year.
	{1, 1, 1, 1},
	{1, 1, 15, 15},
	{1, 2, 1, 32},
	{1, 2, 15, 46},
	{1, 3, 1, 60},
	{1, 3, 15, 74},
	{1, 4, 1, 91},
	{1, 12, 31, 365},

	// Year minus one, a common year.
	{-1, 1, 1, 1},
	{-1, 1, 15, 15},
	{-1, 2, 1, 32},
	{-1, 2, 15, 46},
	{-1, 3, 1, 60},
	{-1, 3, 15, 74},
	{-1, 4, 1, 91},
	{-1, 12, 31, 365},

	// 400 BC, a leap year.
	{-400, 1, 1, 1},
	{-400, 1, 15, 15},
	{-400, 2, 1, 32},
	{-400, 2, 15, 46},
	{-400, 3, 1, 61},
	{-400, 3, 15, 75},
	{-400, 4, 1, 92},
	{-400, 12, 31, 366},

	// The change to the Gregorian calendar has no effect.
	{1582, 10, 4, 277},
	{1582, 10, 15, 288},
}

func TestYearDay(t *testing.T) {
	for _, c := range yearDayCases {
		tm := time.Date(c.year, time.Month(c.month), c.day, 0, 0, 0, 0, time.UTC)
		if got := tm.YearDay(); got != c.yday {
			t.Errorf("Date(%d-%02d-%02d).YearDay() = %d, want %d",
				c.year, c.month, c.day, got, c.yday)
		}
	}
}
