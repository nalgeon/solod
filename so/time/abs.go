// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import "github.com/nalgeon/solod/so/math/bits"

// A Month specifies a month of the year (January = 1, ...).
type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

// A Weekday specifies a day of the week (Sunday = 0, ...).
type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// AbsDate is a date specified by year, month, and day.
type AbsDate struct {
	Year  int
	Month Month
	Day   int
}

// AbsClock is a time of day specified by hour, minute, and second.
type AbsClock struct {
	Hour   int
	Minute int
	Second int
}

// An absSeconds counts the number of seconds since the absolute zero instant.
type absSeconds uint64

// An absDays counts the number of days since the absolute zero instant.
type absDays uint64

// An absCentury counts the number of centuries since the absolute zero instant.
type absCentury uint64

// An absCyear counts the number of years since the start of a century.
type absCyear int

// An absYday counts the number of days since the start of a year.
// Note that absolute years start on March 1.
type absYday int

// An absMonth counts the number of months since the start of a year.
// absMonth=0 denotes March.
type absMonth int

// An absLeap is a single bit (0 or 1) denoting whether a given year is a leap year.
type absLeap int

// An absJanFeb is a single bit (0 or 1) denoting whether a given day falls in
// January or February. That is a special case because the absolute years start
// in March (unlike normal calendar years).
type absJanFeb int

// absSplit is the result of splitting absolute days
// into century, year within century, and day within year.
type absSplit struct {
	century absCentury
	cyear   absCyear
	ayday   absYday
}

// absSeconds_days converts absolute seconds to absolute days.
func absSeconds_days(abs absSeconds) absDays {
	return absDays(abs / secondsPerDay)
}

// absSeconds_clock returns the hour, minute, and second within the day specified by abs.
func absSeconds_clock(abs absSeconds) AbsClock {
	sec := int(abs % secondsPerDay)
	hour := sec / secondsPerHour
	sec -= hour * secondsPerHour
	min := sec / secondsPerMinute
	sec -= min * secondsPerMinute
	return AbsClock{hour, min, sec}
}

// dateToAbsDays takes a standard year/month/day and returns the
// number of days from the absolute epoch to that day.
// The days argument can be out of range and in particular can be negative.
func dateToAbsDays(year int64, month Month, day int) absDays {
	// See "Computations on Times" comment above.
	amonth := uint32(month)
	janFeb := uint32(0)
	if amonth < 3 {
		janFeb = 1
	}
	amonth += 12 * janFeb
	y := uint64(year) - uint64(janFeb) + absoluteYears

	// For amonth is in the range [3,14], we want:
	//
	//	ayday := (153*amonth - 457) / 5
	//
	// (See the "Computations on Times" comment above
	// as well as Neri and Schneider, section 7.)
	//
	// That is equivalent to:
	//
	//	ayday := (979*amonth - 2919) >> 5
	//
	// and the latter form uses a couple fewer instructions,
	// so use it, saving a few cycles.
	// See Neri and Schneider, section 8.3
	// for more about this optimization.
	//
	// (Note that there is no saved division, because the compiler
	// implements / 5 without division in all cases.)
	ayday := (979*amonth - 2919) >> 5

	century := y / 100
	cyear := uint32(y % 100)
	cday := 1461 * cyear / 4
	centurydays := 146097 * century / 4

	return absDays(centurydays + uint64(int64(cday+ayday)+int64(day)-1))
}

// absDays_split splits days into century, cyear, ayday.
func absDays_split(days absDays) absSplit {
	// See "Computations on Times" comment above.
	d := 4*uint64(days) + 3
	century := absCentury(d / 146097)

	// This should be
	//	cday := uint32(d % 146097) / 4
	//	cd := 4*cday + 3
	// which is to say
	//	cday := uint32(d % 146097) >> 2
	//	cd := cday<<2 + 3
	// but of course (x>>2<<2)+3 == x|3,
	// so do that instead.
	cd := uint32(d%146097) | 3

	// For cdays in the range [0,146097] (100 years), we want:
	//
	//	cyear := (4 cdays + 3) / 1461
	//	yday := (4 cdays + 3) % 1461 / 4
	//
	// (See the "Computations on Times" comment above
	// as well as Neri and Schneider, section 7.)
	//
	// That is equivalent to:
	//
	//	cyear := (2939745 cdays) >> 32
	//	yday := (2939745 cdays) & 0xFFFFFFFF / 2939745 / 4
	//
	// so do that instead, saving a few cycles.
	// See Neri and Schneider, section 8.3
	// for more about this optimization.
	hi, lo := bits.Mul32(2939745, cd)
	cyear := absCyear(hi)
	ayday := absYday(lo / 2939745 / 4)
	return absSplit{century, cyear, ayday}
}

// absDays_date converts days into standard year, month, day.
func absDays_date(days absDays) AbsDate {
	split := absDays_split(days)
	amonth, day := absYday_split(split.ayday)
	janFeb := absYday_janFeb(split.ayday)
	year := absCentury_Year(split.century, split.cyear, janFeb)
	month := absMonth_month(amonth, janFeb)
	return AbsDate{year, month, day}
}

// absDays_yearYday converts days into the standard year and 1-based yday.
func absDays_yearYday(days absDays) (int, int) {
	split := absDays_split(days)
	janFeb := absYday_janFeb(split.ayday)
	year := absCentury_Year(split.century, split.cyear, janFeb)
	leap := absCentury_Leap(split.century, split.cyear)
	yday := absYday_yday(split.ayday, janFeb, leap)
	return year, yday
}

// absDays_weekday returns the day of the week specified by days.
func absDays_weekday(days absDays) Weekday {
	// March 1 of the absolute year, like March 1 of 2000, was a Wednesday.
	return Weekday((uint64(days) + uint64(Wednesday)) % 7)
}

// absCentury_Leap returns 1 if (century, cyear) is a leap year, 0 otherwise.
func absCentury_Leap(century absCentury, cyear absCyear) absLeap {
	// See "Computations on Times" comment above.
	y4ok := 0
	if cyear%4 == 0 {
		y4ok = 1
	}
	y100ok := 0
	if cyear != 0 {
		y100ok = 1
	}
	y400ok := 0
	if century%4 == 0 {
		y400ok = 1
	}
	return absLeap(y4ok & (y100ok | y400ok))
}

// absCentury_Year returns the standard year for (century, cyear, janFeb).
func absCentury_Year(century absCentury, cyear absCyear, janFeb absJanFeb) int {
	// See "Computations on Times" comment above.
	return int(uint64(century)*100-absoluteYears) + int(cyear) + int(janFeb)
}

// absYday_split splits ayday into absolute month and standard (1-based) day-in-month.
func absYday_split(ayday absYday) (absMonth, int) {
	// See "Computations on Times" comment above.
	//
	// For yday in the range [0,366],
	//
	//	amonth := (5 yday + 461) / 153
	//	mday := (5 yday + 461) % 153 / 5
	//
	// is equivalent to:
	//
	//	amonth = (2141 yday + 197913) >> 16
	//	mday = (2141 yday + 197913) & 0xFFFF / 2141
	//
	// so do that instead, saving a few cycles.
	// See Neri and Schneider, section 8.3.
	d := 2141*uint32(ayday) + 197913
	month := absMonth(d >> 16)
	mday := 1 + int((d&0xFFFF)/2141)
	return month, mday
}

// absYday_janFeb returns 1 if the March 1-based ayday
// is in January or February, 0 otherwise.
func absYday_janFeb(ayday absYday) absJanFeb {
	// See "Computations on Times" comment above.
	jf := absJanFeb(0)
	if ayday >= marchThruDecember {
		jf = 1
	}
	return jf
}

// absYday_yday returns the standard 1-based yday for (ayday, janFeb, leap).
func absYday_yday(ayday absYday, janFeb absJanFeb, leap absLeap) int {
	// See "Computations on Times" comment above.
	return int(ayday) + (1 + 31 + 28) + (int(leap) &^ int(janFeb)) - 365*int(janFeb)
}

// absMonth_month returns the standard Month for (m, janFeb)
func absMonth_month(m absMonth, janFeb absJanFeb) Month {
	// See "Computations on Times" comment above.
	return Month(m) - Month(janFeb)*12
}
