// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/math"
	"solod.dev/so/time"
)

// The limits of a Duration.
const (
	minDuration time.Duration = math.MinInt64
	maxDuration time.Duration = math.MaxInt64
)

// secondsPerDay is the number of seconds in a day.
const secondsPerDay = 24 * 60 * 60

// daysPerMonth holds the number of days in each month of a common year.
var daysPerMonth = []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// isLeap reports whether the year is a leap year of the proleptic Gregorian
// calendar. The year is an astronomical year number, so year 0 is 1 BC.
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// daysInYear returns the number of days in the year.
func daysInYear(year int) int {
	if isLeap(year) {
		return 366
	}
	return 365
}

// daysInMonth returns the number of days in the month of the year.
func daysInMonth(year int, month time.Month) int {
	if month == time.February && isLeap(year) {
		return 29
	}
	return daysPerMonth[month-1]
}

// A utcCase is a Unix second count and the UTC calendar time it denotes.
type utcCase struct {
	seconds int64
	year    int
	month   time.Month
	day     int
	hour    int
	minute  int
	second  int
	nsec    int
	weekday time.Weekday
}

var utcCases = []utcCase{
	{0, 1970, time.January, 1, 0, 0, 0, 0, time.Thursday},
	{1221681866, 2008, time.September, 17, 20, 4, 26, 0, time.Wednesday},
	{-1221681866, 1931, time.April, 16, 3, 55, 34, 0, time.Thursday},
	{-11644473600, 1601, time.January, 1, 0, 0, 0, 0, time.Monday},
	{599529660, 1988, time.December, 31, 0, 1, 0, 0, time.Saturday},
	{978220860, 2000, time.December, 31, 0, 1, 0, 0, time.Sunday},
}

var nanoUTCCases = []utcCase{
	{0, 1970, time.January, 1, 0, 0, 0, 1e8, time.Thursday},
	{1221681866, 2008, time.September, 17, 20, 4, 26, 2e8, time.Wednesday},
}

// sameTime reports whether tm denotes the UTC calendar time of the case.
func sameTime(tm time.Time, c utcCase) bool {
	// Check the aggregates.
	date := tm.Date(time.UTC)
	clock := tm.Clock(time.UTC)
	if date.Year != c.year || date.Month != c.month || date.Day != c.day {
		return false
	}
	if clock.Hour != c.hour || clock.Minute != c.minute || clock.Second != c.second {
		return false
	}
	// Check the individual entries.
	return tm.Year() == c.year &&
		tm.Month() == c.month &&
		tm.Day() == c.day &&
		tm.Hour() == c.hour &&
		tm.Minute() == c.minute &&
		tm.Second() == c.second &&
		tm.Nanosecond() == c.nsec &&
		tm.Weekday() == c.weekday
}
