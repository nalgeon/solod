// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"testing"

	. "github.com/nalgeon/solod/so/time"
)

var tzUTCMinus5 = FixedZone("UTC-5", -5*3600)
var tzUTCMinus12 = FixedZone("UTC-12", -12*3600)
var tzUTCPlus1 = FixedZone("UTC+1", 1*3600)
var tzUTCPlus12 = FixedZone("UTC+12", 12*3600)

func TestZeroTime(t *testing.T) {
	var zero Time
	date := zero.Date()
	clock := zero.Clock()
	nsec := zero.Nanosecond()
	yday := zero.YearDay()
	wday := zero.Weekday()
	if date.Year != 1 || date.Month != January || date.Day != 1 || clock.Hour != 0 || clock.Minute != 0 || clock.Second != 0 || nsec != 0 || yday != 1 || wday != Monday {
		t.Errorf("zero time = %v %v %v year %v %02d:%02d:%02d.%09d yday %d want Monday Jan 1 year 1 00:00:00.000000000 yday 1",
			wday, date.Month, date.Day, date.Year, clock.Hour, clock.Minute, clock.Second, nsec, yday)
	}
}

// parsedTime is the struct representing a parsed time value.
type parsedTime struct {
	Year                 int
	Month                Month
	Day                  int
	Hour, Minute, Second int // 15:04:05 is 15, 4, 5.
	Nanosecond           int // Fractional second.
	Weekday              Weekday
	ZoneOffset           int    // seconds east of UTC, e.g. -7*60*60 for -0700
	Zone                 string // e.g., "MST"
}

type TimeTest struct {
	seconds int64
	golden  parsedTime
}

var utctests = []TimeTest{
	{0, parsedTime{1970, January, 1, 0, 0, 0, 0, Thursday, 0, "UTC"}},
	{1221681866, parsedTime{2008, September, 17, 20, 4, 26, 0, Wednesday, 0, "UTC"}},
	{-1221681866, parsedTime{1931, April, 16, 3, 55, 34, 0, Thursday, 0, "UTC"}},
	{-11644473600, parsedTime{1601, January, 1, 0, 0, 0, 0, Monday, 0, "UTC"}},
	{599529660, parsedTime{1988, December, 31, 0, 1, 0, 0, Saturday, 0, "UTC"}},
	{978220860, parsedTime{2000, December, 31, 0, 1, 0, 0, Sunday, 0, "UTC"}},
}

var nanoutctests = []TimeTest{
	{0, parsedTime{1970, January, 1, 0, 0, 0, 1e8, Thursday, 0, "UTC"}},
	{1221681866, parsedTime{2008, September, 17, 20, 4, 26, 2e8, Wednesday, 0, "UTC"}},
}

var dateTests = []struct {
	year, month, day, hour, min, sec, nsec int
	z                                      *Location
	unix                                   int64
}{
	{2011, 11, 6, 8, 0, 0, 0, UTC, 1320566400},   // 8:00:00 UTC
	{2011, 11, 6, 8, 59, 59, 0, UTC, 1320569999}, // 8:59:59 UTC
	{2011, 11, 6, 10, 0, 0, 0, UTC, 1320573600},  // 10:00:00 UTC

	{2011, 3, 13, 9, 0, 0, 0, UTC, 1300006800},   // 9:00:00 UTC
	{2011, 3, 13, 9, 59, 59, 0, UTC, 1300010399}, // 9:59:59 UTC
	{2011, 3, 13, 10, 0, 0, 0, UTC, 1300010400},  // 10:00:00 UTC
	{2011, 3, 13, 9, 30, 0, 0, UTC, 1300008600},  // 9:30:00 UTC
	{2012, 12, 24, 8, 0, 0, 0, UTC, 1356336000},  // Leap year

	// Many names for 2011-11-18 15:56:35.0 UTC
	{2011, 11, 18, 15, 56, 35, 0, UTC, 1321631795},                 // Nov 18 15:56:35
	{2011, 11, 19, -9, 56, 35, 0, UTC, 1321631795},                 // Nov 19 -9:56:35
	{2011, 11, 17, 39, 56, 35, 0, UTC, 1321631795},                 // Nov 17 39:56:35
	{2011, 11, 18, 14, 116, 35, 0, UTC, 1321631795},                // Nov 18 14:116:35
	{2011, 10, 49, 15, 56, 35, 0, UTC, 1321631795},                 // Oct 49 15:56:35
	{2011, 11, 18, 15, 55, 95, 0, UTC, 1321631795},                 // Nov 18 15:55:95
	{2011, 11, 18, 15, 56, 34, 1e9, UTC, 1321631795},               // Nov 18 15:56:34 + 10⁹ns
	{2011, 12, -12, 15, 56, 35, 0, UTC, 1321631795},                // Dec -12 15:56:35
	{2012, 1, -43, 15, 56, 35, 0, UTC, 1321631795},                 // 2012 Jan -43 15:56:35
	{2012, int(January - 2), 18, 15, 56, 35, 0, UTC, 1321631795},   // 2012 (Jan-2) 18 15:56:35
	{2010, int(December + 11), 18, 15, 56, 35, 0, UTC, 1321631795}, // 2010 (Dec+11) 18 15:56:35
	{1970, 1, 15297, 15, 56, 35, 0, UTC, 1321631795},               // large number of days
	{2011, 11, 18, 10, 56, 35, 0, &tzUTCMinus5, 1321631795},        // UTC-5
	{2011, 11, 18, 3, 56, 35, 0, &tzUTCMinus12, 1321631795},        // UTC-12
	{2011, 11, 18, 16, 56, 35, 0, &tzUTCPlus1, 1321631795},         // UTC+1
	{2011, 11, 19, 3, 56, 35, 0, &tzUTCPlus12, 1321631795},         // UTC+12

	{1970, 1, -25508, 8, 0, 0, 0, UTC, -2203948800}, // negative Unix time
}

func TestDate(t *testing.T) {
	for _, tt := range dateTests {
		time := Date(tt.year, Month(tt.month), tt.day, tt.hour, tt.min, tt.sec, tt.nsec, tt.z)
		want := Unix(tt.unix, 0)
		if !time.Equal(want) {
			t.Errorf("Date(%d, %d, %d, %d, %d, %d, %d, %s) = %v, want %v",
				tt.year, tt.month, tt.day, tt.hour, tt.min, tt.sec, tt.nsec, tt.z,
				time, want)
		}
	}
}

var defaultLocTests = []struct {
	name string
	f    func(t1, t2 Time) bool
}{
	{"After", func(t1, t2 Time) bool { return t1.After(t2) == t2.After(t1) }},
	{"Before", func(t1, t2 Time) bool { return t1.Before(t2) == t2.Before(t1) }},
	{"Equal", func(t1, t2 Time) bool { return t1.Equal(t2) == t2.Equal(t1) }},
	{"Compare", func(t1, t2 Time) bool { return t1.Compare(t2) == t2.Compare(t1) }},

	{"IsZero", func(t1, t2 Time) bool { return t1.IsZero() == t2.IsZero() }},
	{"Date", func(t1, t2 Time) bool {
		d1 := t1.Date()
		d2 := t2.Date()
		return d1 == d2
	}},
	{"Year", func(t1, t2 Time) bool { return t1.Year() == t2.Year() }},
	{"Month", func(t1, t2 Time) bool { return t1.Month() == t2.Month() }},
	{"Day", func(t1, t2 Time) bool { return t1.Day() == t2.Day() }},
	{"Weekday", func(t1, t2 Time) bool { return t1.Weekday() == t2.Weekday() }},
	{"ISOWeek", func(t1, t2 Time) bool {
		a1, b1 := t1.ISOWeek()
		a2, b2 := t2.ISOWeek()
		return a1 == a2 && b1 == b2
	}},
	{"Clock", func(t1, t2 Time) bool {
		c1 := t1.Clock()
		c2 := t2.Clock()
		return c1 == c2
	}},
	{"Hour", func(t1, t2 Time) bool { return t1.Hour() == t2.Hour() }},
	{"Minute", func(t1, t2 Time) bool { return t1.Minute() == t2.Minute() }},
	{"Second", func(t1, t2 Time) bool { return t1.Second() == t2.Second() }},
	{"Nanosecond", func(t1, t2 Time) bool { return t1.Hour() == t2.Hour() }},
	{"YearDay", func(t1, t2 Time) bool { return t1.YearDay() == t2.YearDay() }},

	// Using Equal since Add don't modify loc using "==" will cause a fail
	{"Add", func(t1, t2 Time) bool { return t1.Add(Hour).Equal(t2.Add(Hour)) }},
	{"Sub", func(t1, t2 Time) bool { return t1.Sub(t2) == t2.Sub(t1) }},

	// Original cause for this test case bug 15852
	{"AddDate", func(t1, t2 Time) bool { return t1.AddDate(1991, 9, 3) == t2.AddDate(1991, 9, 3) }},

	{"UTC", func(t1, t2 Time) bool { return t1.UTC() == t2.UTC() }},
	{"In", func(t1, t2 Time) bool { return t1.In(UTC) == t2.In(UTC) }},

	{"Zone", func(t1, t2 Time) bool {
		a1, b1 := t1.Zone()
		a2, b2 := t2.Zone()
		return a1 == a2 && b1 == b2
	}},

	{"Unix", func(t1, t2 Time) bool { return t1.Unix() == t2.Unix() }},
	{"UnixNano", func(t1, t2 Time) bool { return t1.UnixNano() == t2.UnixNano() }},
	{"UnixMilli", func(t1, t2 Time) bool { return t1.UnixMilli() == t2.UnixMilli() }},
	{"UnixMicro", func(t1, t2 Time) bool { return t1.UnixMicro() == t2.UnixMicro() }},

	{"Truncate", func(t1, t2 Time) bool { return t1.Truncate(Hour).Equal(t2.Truncate(Hour)) }},
	{"Round", func(t1, t2 Time) bool { return t1.Round(Hour).Equal(t2.Round(Hour)) }},

	{"== Time{}", func(t1, t2 Time) bool { return (t1 == Time{}) == (t2 == Time{}) }},
}

func TestDefaultLoc(t *testing.T) {
	// Verify that all of Time's methods behave identically if loc is set to
	// nil or UTC.
	for _, tt := range defaultLocTests {
		t1 := Time{}
		t2 := Time{}.UTC()
		if !tt.f(t1, t2) {
			t.Errorf("Time{} and Time{}.UTC() behave differently for %s", tt.name)
		}
	}
}
