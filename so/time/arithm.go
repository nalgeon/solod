// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import "github.com/nalgeon/solod/so"

// Computations on Times
//
// The zero value for a Time is defined to be
//	January 1, year 1, 00:00:00.000000000 UTC
// which (1) looks like a zero, or as close as you can get in a date
// (1-1-1 00:00:00 UTC), (2) is unlikely enough to arise in practice to
// be a suitable "not set" sentinel, unlike Jan 1 1970, and (3) has a
// non-negative year even in time zones west of UTC, unlike 1-1-0
// 00:00:00 UTC, which would be 12-31-(-1) 19:00:00 in New York.
//
// The zero Time value does not force a specific epoch for the time
// representation. For example, to use the Unix epoch internally, we
// could define that to distinguish a zero value from Jan 1 1970, that
// time would be represented by sec=-1, nsec=1e9. However, it does
// suggest a representation, namely using 1-1-1 00:00:00 UTC as the
// epoch, and that's what we do.
//
// The Add and Sub computations are oblivious to the choice of epoch.
//
// The presentation computations - year, month, minute, and so on - all
// rely heavily on division and modulus by positive constants. For
// calendrical calculations we want these divisions to round down, even
// for negative values, so that the remainder is always positive, but
// Go's division (like most hardware division instructions) rounds to
// zero. We can still do those computations and then adjust the result
// for a negative numerator, but it's annoying to write the adjustment
// over and over. Instead, we can change to a different epoch so long
// ago that all the times we care about will be positive, and then round
// to zero and round down coincide. These presentation routines already
// have to add the zone offset, so adding the translation to the
// alternate epoch is cheap. For example, having a non-negative time t
// means that we can write
//
//	sec = t % 60
//
// instead of
//
//	sec = t % 60
//	if sec < 0 {
//		sec += 60
//	}
//
// everywhere.
//
// The calendar runs on an exact 400 year cycle: a 400-year calendar
// printed for 1970-2369 will apply as well to 2370-2769. Even the days
// of the week match up. It simplifies date computations to choose the
// cycle boundaries so that the exceptional years are always delayed as
// long as possible: March 1, year 0 is such a day:
// the first leap day (Feb 29) is four years minus one day away,
// the first multiple-of-4 year without a Feb 29 is 100 years minus one day away,
// and the first multiple-of-100 year with a Feb 29 is 400 years minus one day away.
// March 1 year Y for any Y = 0 mod 400 is also such a day.
//
// Finally, it's convenient if the delta between the Unix epoch and
// long-ago epoch is representable by an int64 constant.
//
// These three considerations—choose an epoch as early as possible, that
// starts on March 1 of a year equal to 0 mod 400, and that is no more than
// 2⁶³ seconds earlier than 1970—bring us to the year -292277022400.
// We refer to this moment as the absolute zero instant, and to times
// measured as a uint64 seconds since this year as absolute times.
//
// Times measured as an int64 seconds since the year 1—the representation
// used for Time's sec field—are called internal times.
//
// Times measured as an int64 seconds since the year 1970 are called Unix
// times.
//
// It is tempting to just use the year 1 as the absolute epoch, defining
// that the routines are only valid for years >= 1. However, the
// routines would then be invalid when displaying the epoch in time zones
// west of UTC, since it is year 0. It doesn't seem tenable to say that
// printing the zero time correctly isn't supported in half the time
// zones. By comparison, it's reasonable to mishandle some times in
// the year -292277022400.
//
// All this is opaque to clients of the API and can be changed if a
// better implementation presents itself.
//
// The date calculations are implemented using the following clever math from
// Cassio Neri and Lorenz Schneider, "Euclidean affine functions and their
// application to calendar algorithms," SP&E 2023. https://doi.org/10.1002/spe.3172
//
// Define a "calendrical division" (f, f°, f*) to be a triple of functions converting
// one time unit into a whole number of larger units and the remainder and back.
// For example, in a calendar with no leap years, (d/365, d%365, y*365) is the
// calendrical division for days into years:
//
//	(f)  year := days/365
//	(f°) yday := days%365
//	(f*) days := year*365 (+ yday)
//
// Note that f* is usually the "easy" function to write: it's the
// calendrical multiplication that inverts the more complex division.
//
// Neri and Schneider prove that when f* takes the form
//
//	f*(n) = (a n + b) / c
//
// using integer division rounding down with a ≥ c > 0,
// which they call a Euclidean affine function or EAF, then:
//
//	f(n) = (c n + c - b - 1) / a
//	f°(n) = (c n + c - b - 1) % a / c
//
// This gives a fairly direct calculation for any calendrical division for which
// we can write the calendrical multiplication in EAF form.
// Because the epoch has been shifted to March 1, all the calendrical
// multiplications turn out to be possible to write in EAF form.
// When a date is broken into [century, cyear, amonth, mday],
// with century, cyear, and mday 0-based,
// and amonth 3-based (March = 3, ..., January = 13, February = 14),
// the calendrical multiplications written in EAF form are:
//
//	yday = (153 (amonth-3) + 2) / 5 = (153 amonth - 457) / 5
//	cday = 365 cyear + cyear/4 = 1461 cyear / 4
//	centurydays = 36524 century + century/4 = 146097 century / 4
//	days = centurydays + cday + yday + mday.
//
// We can only handle one periodic cycle per equation, so the year
// calculation must be split into [century, cyear], handling both the
// 100-year cycle and the 400-year cycle.
//
// The yday calculation is not obvious but derives from the fact
// that the March through January calendar repeats the 5-month
// 153-day cycle 31, 30, 31, 30, 31 (we don't care about February
// because yday only ever count the days _before_ February 1,
// since February is the last month).
//
// Using the rule for deriving f and f° from f*, these multiplications
// convert to these divisions:
//
//	century := (4 days + 3) / 146097
//	cdays := (4 days + 3) % 146097 / 4
//	cyear := (4 cdays + 3) / 1461
//	ayday := (4 cdays + 3) % 1461 / 4
//	amonth := (5 ayday + 461) / 153
//	mday := (5 ayday + 461) % 153 / 5
//
// The a in ayday and amonth stands for absolute (March 1-based)
// to distinguish from the standard yday (January 1-based).
//
// After computing these, we can translate from the March 1 calendar
// to the standard January 1 calendar with branch-free math assuming a
// branch-free conversion from bool to int 0 or 1, denoted int(b) here:
//
//	isJanFeb := int(yday >= marchThruDecember)
//	month := amonth - isJanFeb*12
//	year := century*100 + cyear + isJanFeb
//	isLeap := int(cyear%4 == 0) & (int(cyear != 0) | int(century%4 == 0))
//	day := 1 + mday
//	yday := 1 + ayday + 31 + 28 + isLeap&^isJanFeb - 365*isJanFeb
//
// isLeap is the standard leap-year rule, but the split year form
// makes the divisions all reduce to binary masking.
// Note that day and yday are 1-based, in contrast to mday and ayday.

// To keep the various units separate, we define integer types
// for each. These are never stored in interfaces nor allocated,
// so their type information does not appear in Go binaries.
const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour

	// Days from March 1 through end of year
	marchThruDecember = 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31

	// absoluteYears is the number of years we subtract from internal time to get absolute time.
	// This value must be 0 mod 400, and it defines the "absolute zero instant"
	// mentioned in the "Computations on Times" comment above: March 1, -absoluteYears.
	// Dates before the absolute epoch will not compute correctly,
	// but otherwise the value can be changed as needed.
	absoluteYears = 292277022400

	// Offsets to convert between internal and absolute or Unix times.
	absoluteToInternal int64 = -(absoluteYears*365.2425 + marchThruDecember) * secondsPerDay
	internalToAbsolute       = -absoluteToInternal

	unixToInternal int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsPerDay
	internalToUnix int64 = -unixToInternal

	absoluteToUnix = absoluteToInternal + internalToUnix

	wallToInternal int64 = (1884*365 + 1884/4 - 1884/100 + 1884/400) * secondsPerDay
)

// Date returns the year, month, and day in which t occurs.
func (t Time) Date() AbsDate {
	sec := t.absSec()
	days := absSeconds_days(sec)
	return absDays_date(days)
}

// Year returns the year in which t occurs.
func (t Time) Year() int {
	sec := t.absSec()
	days := absSeconds_days(sec)
	split := absDays_split(days)
	janFeb := absYday_janFeb(split.ayday)
	return absCentury_Year(split.century, split.cyear, janFeb)
}

// Month returns the month of the year specified by t.
func (t Time) Month() Month {
	sec := t.absSec()
	days := absSeconds_days(sec)
	split := absDays_split(days)
	month, _ := absYday_split(split.ayday)
	janFeb := absYday_janFeb(split.ayday)
	return absMonth_month(month, janFeb)
}

// Day returns the day of the month specified by t.
func (t Time) Day() int {
	sec := t.absSec()
	days := absSeconds_days(sec)
	split := absDays_split(days)
	_, day := absYday_split(split.ayday)
	return day
}

// Weekday returns the day of the week specified by t.
func (t Time) Weekday() Weekday {
	sec := t.absSec()
	days := absSeconds_days(sec)
	return absDays_weekday(days)
}

// ISOWeek returns the ISO 8601 year and week number in which t occurs.
// Week ranges from 1 to 53. Jan 01 to Jan 03 of year n might belong to
// week 52 or 53 of year n-1, and Dec 29 to Dec 31 might belong to week 1
// of year n+1.
func (t Time) ISOWeek() (int, int) {
	// According to the rule that the first calendar week of a calendar year is
	// the week including the first Thursday of that year, and that the last one is
	// the week immediately preceding the first calendar week of the next calendar year.
	// See https://www.iso.org/obp/ui#iso:std:iso:8601:-1:ed-1:v1:en:term:3.1.1.23 for details.

	// weeks start with Monday
	// Monday Tuesday Wednesday Thursday Friday Saturday Sunday
	// 1      2       3         4        5      6        7
	// +3     +2      +1        0        -1     -2       -3
	// the offset to Thursday
	sec := t.absSec()
	days := absSeconds_days(sec)
	wday := absDays_weekday(days-1) + 1
	thu := days + absDays(Thursday-wday)
	year, yday := absDays_yearYday(thu)
	week := (yday-1)/7 + 1
	return year, week
}

// Clock returns the hour, minute, and second within the day specified by t.
func (t Time) Clock() AbsClock {
	sec := t.absSec()
	return absSeconds_clock(sec)
}

// Hour returns the hour within the day specified by t, in the range [0, 23].
func (t Time) Hour() int {
	return int(t.absSec()%secondsPerDay) / secondsPerHour
}

// Minute returns the minute offset within the hour specified by t, in the range [0, 59].
func (t Time) Minute() int {
	return int(t.absSec()%secondsPerHour) / secondsPerMinute
}

// Second returns the second offset within the minute specified by t, in the range [0, 59].
func (t Time) Second() int {
	return int(t.absSec() % secondsPerMinute)
}

// Nanosecond returns the nanosecond offset within the second specified by t,
// in the range [0, 999999999].
func (t Time) Nanosecond() int {
	return int(t.nsec())
}

// YearDay returns the day of the year specified by t, in the range [1,365] for non-leap years,
// and [1,366] in leap years.
func (t Time) YearDay() int {
	sec := t.absSec()
	days := absSeconds_days(sec)
	_, yday := absDays_yearYday(days)
	return yday
}

// UTC returns t with the location set to UTC.
func (t Time) UTC() Time {
	t.setLoc(UTC)
	return t
}

// Add returns the time t+d.
func (t Time) Add(d Duration) Time {
	dsec := int64(d / 1000000000)
	nsec := t.nsec() + int32(d%1000000000)
	if nsec >= 1000000000 {
		dsec++
		nsec -= 1000000000
	} else if nsec < 0 {
		dsec--
		nsec += 1000000000
	}
	t.wall = (t.wall &^ nsecMask) | uint64(nsec) // update nsec
	t.addSec(dsec)
	if (t.wall & hasMonotonic) != 0 {
		te := t.ext + int64(d)
		if (d < 0 && te > t.ext) || (d > 0 && te < t.ext) {
			// Monotonic clock reading now out of range; degrade to wall-only.
			t.stripMono()
		} else {
			t.ext = te
		}
	}
	return t
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a [Duration], the maximum (or minimum) duration
// will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (t Time) Sub(u Time) Duration {
	if (t.wall & u.wall & hasMonotonic) != 0 {
		return subMono(t.ext, u.ext)
	}
	d := Duration(t.sec()-u.sec())*Second + Duration(t.nsec()-u.nsec())
	// Check for overflow or underflow.
	if u.Add(d).Equal(t) {
		return d // d is correct
	} else if t.Before(u) {
		return minDuration // t - u is negative out of range
	} else {
		return maxDuration // t - u is positive out of range
	}
}

func subMono(t, u int64) Duration {
	d := Duration(t - u)
	if d < 0 && t > u {
		return maxDuration // t - u is positive out of range
	}
	if d > 0 && t < u {
		return minDuration // t - u is negative out of range
	}
	return d
}

// Since returns the time elapsed since t.
// It is shorthand for time.Now().Sub(t).
func Since(t Time) Duration {
	if (t.wall & hasMonotonic) != 0 {
		// Common case optimization: if t has monotonic time, then Sub will use only it.
		return subMono(time_mono()-time_monoStart, t.ext)
	}
	return Now().Sub(t)
}

// Until returns the duration until t.
// It is shorthand for t.Sub(time.Now()).
func Until(t Time) Duration {
	if (t.wall & hasMonotonic) != 0 {
		// Common case optimization: if t has monotonic time, then Sub will use only it.
		return subMono(t.ext, time_mono()-time_monoStart)
	}
	return t.Sub(Now())
}

// AddDate returns the time corresponding to adding the
// given number of years, months, and days to t.
// For example, AddDate(-1, 2, 3) applied to January 1, 2011
// returns March 4, 2010.
//
// Note that dates are fundamentally coupled to timezones, and calendrical
// periods like days don't have fixed durations. AddDate uses the Location of
// the Time value to determine these durations. That means that the same
// AddDate arguments can produce a different shift in absolute time depending on
// the base Time value and its Location. For example, AddDate(0, 0, 1) applied
// to 12:00 on March 27 always returns 12:00 on March 28. At some locations and
// in some years this is a 24 hour shift. In others it's a 23 hour shift due to
// daylight savings time transitions.
//
// AddDate normalizes its result in the same way that Date does,
// so, for example, adding one month to October 31 yields
// December 1, the normalized form for November 31.
func (t Time) AddDate(years int, months int, days int) Time {
	date := t.Date()
	clock := t.Clock()
	return Date(
		date.Year+years, date.Month+Month(months), date.Day+days,
		clock.Hour, clock.Minute, clock.Second, int(t.nsec()),
		t.Location(),
	)
}

// Truncate returns the result of rounding t down to a multiple of d (since the zero time).
// If d <= 0, Truncate returns t stripped of any monotonic clock reading but otherwise unchanged.
//
// Truncate operates on the time as an absolute duration since the
// zero time; it does not operate on the presentation form of the
// time. Thus, Truncate(Hour) may return a time with a non-zero
// minute, depending on the time's Location.
func (t Time) Truncate(d Duration) Time {
	t.stripMono()
	if d <= 0 {
		return t
	}
	r := time_div(t, d)
	return t.Add(-r)
}

// Round returns the result of rounding t to the nearest multiple of d (since the zero time).
// The rounding behavior for halfway values is to round up.
// If d <= 0, Round returns t stripped of any monotonic clock reading but otherwise unchanged.
//
// Round operates on the time as an absolute duration since the
// zero time; it does not operate on the presentation form of the
// time. Thus, Round(Hour) may return a time with a non-zero
// minute, depending on the time's Location.
func (t Time) Round(d Duration) Time {
	t.stripMono()
	if d <= 0 {
		return t
	}
	r := time_div(t, d)
	if lessThanHalf(r, d) {
		return t.Add(-r)
	}
	return t.Add(d - r)
}

// absSec returns the time t as an absolute seconds, adjusted by the zone offset.
// It is called when computing a presentation property like Month or Hour.
// We'd rather call it abs, but there are linknames to abs that make that problematic.
// See timeAbs below.
func (t Time) absSec() absSeconds {
	sec := t.unixSec()
	if t.loc != nil {
		sec += int64(t.loc.offset)
	}
	return absSeconds(sec + (unixToInternal + internalToAbsolute))
}

// addSec adds d seconds to the time.
func (t *Time) addSec(d int64) {
	if (t.wall & hasMonotonic) != 0 {
		sec := int64((t.wall << 1) >> (nsecShift + 1))
		dsec := sec + d
		if (0 <= dsec) && (dsec <= 8589934591) { // 1<<33 - 1
			t.wall = (t.wall & nsecMask) | (uint64(dsec) << nsecShift) | hasMonotonic
			return
		}
		// Wall second now out of range for packed field.
		// Move to ext.
		t.stripMono()
	}

	// Check if the sum of t.ext and d overflows and handle it properly.
	sum := t.ext + d
	if (sum > t.ext) == (d > 0) {
		t.ext = sum
	} else if d > 0 {
		t.ext = int64(so.MaxInt64)
	} else {
		t.ext = -int64(so.MaxInt64)
	}
}

// div divides t by d and returns the remainder.
func time_div(t Time, d Duration) Duration {
	var r Duration

	neg := false
	nsec := t.nsec()
	sec := t.sec()
	if sec < 0 {
		// Operate on absolute value.
		neg = true
		sec = -sec
		nsec = -nsec
		if nsec < 0 {
			nsec += 1000000000
			sec-- // sec >= 1 before the -- so safe
		}
	}

	// Special case: 2d divides 1 second.
	if d < Second && Second%(d+d) == 0 {
		r = Duration(nsec % int32(d))

		// Special case: d is a multiple of 1 second.
	} else if d%Second == 0 {
		d1 := int64(d / Second)
		r = Duration(sec%d1)*Second + Duration(nsec)

		// General case.
		// This could be faster if more cleverness were applied,
		// but it's really only here to avoid special case restrictions in the API.
		// No one will care about these cases.
	} else {
		// Compute nanoseconds as 128-bit number.
		usec := uint64(sec)
		tmp := (usec >> 32) * 1000000000
		u1 := tmp >> 32
		u0 := tmp << 32
		tmp = (usec & 0xFFFFFFFF) * 1000000000
		u0x, u0 := u0, u0+tmp
		if u0 < u0x {
			u1++
		}
		u0x, u0 = u0, u0+uint64(nsec)
		if u0 < u0x {
			u1++
		}

		// Compute remainder by subtracting r<<k for decreasing k.
		// Quotient parity is whether we subtract on last round.
		d1 := uint64(d)
		for d1>>63 != 1 {
			d1 <<= 1
		}
		d0 := uint64(0)
		for {
			if u1 > d1 || (u1 == d1 && u0 >= d0) {
				// subtract
				u0x, u0 = u0, u0-d0
				if u0 > u0x {
					u1--
				}
				u1 -= d1
			}
			if d1 == 0 && d0 == uint64(d) {
				break
			}
			d0 >>= 1
			d0 |= (d1 & 1) << 63
			d1 >>= 1
		}
		r = Duration(u0)
	}

	if neg && r != 0 {
		r = d - r
	}
	return r
}
