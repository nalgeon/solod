// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package time provides functionality for measuring and displaying time.
//
// The calendrical calculations always assume a Gregorian calendar, with
// no leap seconds.
//
// Based on the [time] package, with fewer features:
//
//   - Time is always stored as UTC internally
//   - Fixed time zones (UTC offsets) instead of Locations.
//   - Formatting and parsing use C verbs instead of Go verbs.
//
// # Monotonic Clocks
//
// Operating systems provide both a "wall clock," which is subject to
// changes for clock synchronization, and a "monotonic clock," which is
// not. The general rule is that the wall clock is for telling time and
// the monotonic clock is for measuring time. Rather than split the API,
// in this package the Time returned by [Now] contains both a wall
// clock reading and a monotonic clock reading; later time-telling
// operations use the wall clock reading, but later time-measuring
// operations, specifically comparisons and subtractions, use the
// monotonic clock reading.
//
// For example, this code always computes a positive elapsed time of
// approximately 20 milliseconds, even if the wall clock is changed during
// the operation being timed:
//
//	start := time.Now()
//	... operation that takes 20 milliseconds ...
//	t := time.Now()
//	elapsed := t.Sub(start)
//
// Other idioms, such as [Since](start), [Until](deadline), and
// Now().Before(deadline), are similarly robust against wall clock
// resets.
//
// The rest of this section gives the precise details of how operations
// use monotonic clocks, but understanding those details is not required
// to use this package.
//
// The Time returned by [Now] contains a monotonic clock reading.
// If Time t has a monotonic clock reading, t.Add adds the same duration to
// both the wall clock and monotonic clock readings to compute the result.
// Because t.AddDate(y, m, d), t.Round(d), and t.Truncate(d) are wall time
// computations, they always strip any monotonic clock reading from their results.
// The canonical way to strip a monotonic clock reading is to use t = t.Round(0).
//
// If Times t and u both contain monotonic clock readings, the operations
// t.After(u), t.Before(u), t.Equal(u), t.Compare(u), and t.Sub(u) are carried out
// using the monotonic clock readings alone, ignoring the wall clock
// readings. If either t or u contains no monotonic clock reading, these
// operations fall back to using the wall clock readings.
//
// On some systems the monotonic clock will stop if the computer goes to sleep.
// On such a system, t.Sub(u) may not accurately reflect the actual
// time that passed between t and u. The same applies to other functions and
// methods that subtract times, such as [Since], [Until], [Time.Before], [Time.After],
// [Time.Add], [Time.Equal] and [Time.Compare]. In some cases, you may need to strip
// the monotonic clock to get accurate results.
//
// Because the monotonic clock reading has no meaning outside
// the current process, the constructors [Date], [Parse],
// and [Unix], always create times with no monotonic clock reading.
//
// The monotonic clock reading exists only in [Time] values. It is not
// a part of [Duration] values or the Unix times returned by t.Unix and
// friends.
//
// Note that the == operator compares not just the time instant but
// also the monotonic clock reading. See the documentation for the
// Time type for a discussion of equality testing for Time values.
//
// For debugging, the result of t.String does include the monotonic
// clock reading if present. If t != u because of different monotonic clock readings,
// that difference will be visible when printing t.String() and u.String().
//
// [time]: https://github.com/golang/go/blob/go1.26.1/src/time/time.go
package time

// Date returns the Time in UTC corresponding to
//
//	yyyy-mm-dd hh:mm:ss + nsec nanoseconds
//
// with respect to the given offset (seconds east of UTC).
//
// The month, day, hour, min, sec, and nsec values may be outside
// their usual ranges and will be normalized during the conversion.
// For example, October 32 converts to November 1.
//
// A daylight savings time transition skips or repeats times.
// For example, in the United States, March 13, 2011 2:15am never occurred,
// while November 6, 2011 1:15am occurred twice. In such cases, the
// choice of time zone, and therefore the time, is not well-defined.
// Date returns a time that is correct in one of the two zones involved
// in the transition, but it does not guarantee which.
func Date(year int, month Month, day, hour, min, sec, nsec int, offset Offset) Time {
	// Normalize month, overflowing into year.
	m := int(month) - 1
	year, m = norm(year, m, 12)
	month = Month(m) + 1

	// Normalize nsec, sec, min, hour, overflowing into day.
	sec, nsec = norm(sec, nsec, 1000000000)
	min, sec = norm(min, sec, 60)
	hour, min = norm(hour, min, 60)
	day, hour = norm(day, hour, 24)

	// Convert to absolute time and then Unix time.
	//
	// The absolute epoch is about 292 billion years back, so the day count
	// times the length of a day is above int64 for every date. The subtraction
	// brings the value back into range. C does not define a signed overflow, so
	// the calculation runs in uint64.
	abs := uint64(dateToAbsDays(int64(year), month, day))*secondsPerDay +
		uint64(int64(hour*secondsPerHour+min*secondsPerMinute+sec))
	unixSec := int64(abs - uint64(unixToInternal+internalToAbsolute))

	// Adjust to UTC by subtracting the offset.
	if offset != 0 {
		unixSec -= int64(offset)
	}

	return unixTime(unixSec, int32(nsec))
}

// Now returns the current time in UTC.
//
// In a freestanding environment, Now reads the clock through the so_time_wall
// and so_time_mono hooks of the target.
func Now() Time {
	sec, nsec := time_wall()
	mono := time_mono()
	if mono == 0 {
		return Time{uint64(nsec), sec + unixToInternal}
	}
	mono -= monoStart
	sec += unixToInternal - minWall
	if (uint64(sec) >> 33) != 0 {
		// Seconds field overflowed the 33 bits available when
		// storing a monotonic time. This will be true after
		// March 16, 2157.
		return Time{uint64(nsec), sec + minWall}
	}
	wall := hasMonotonic | (uint64(sec) << nsecShift) | uint64(nsec)
	return Time{wall, mono}
}

// A Time represents an instant in time with nanosecond precision.
// Time always represents UTC internally.
//
// Programs using times should typically store and pass them as values,
// not pointers. That is, time variables and struct fields should be of
// type time.Time, not *time.Time.
//
// The zero value of type Time is January 1, year 1, 00:00:00.000000000 UTC.
// As this time is unlikely to come up in practice, the [Time.IsZero] method gives
// a simple way of detecting a time that has not been initialized explicitly.
//
// In addition to the required "wall clock" reading, a Time may contain an optional
// reading of the current process's monotonic clock, to provide additional precision
// for comparison or subtraction. See the "Monotonic Clocks" section in the package
// documentation for details.
type Time struct {
	// wall and ext encode the wall time seconds, wall time nanoseconds,
	// and optional monotonic clock reading in nanoseconds.
	//
	// From high to low bit position, wall encodes a 1-bit flag (hasMonotonic),
	// a 33-bit seconds field, and a 30-bit wall time nanoseconds field.
	// The nanoseconds field is in the range [0, 999999999].
	// If the hasMonotonic bit is 0, then the 33-bit field must be zero
	// and the full signed 64-bit wall seconds since Jan 1 year 1 is stored in ext.
	// If the hasMonotonic bit is 1, then the 33-bit field holds a 33-bit
	// unsigned wall seconds since Jan 1 year 1885, and ext holds a
	// signed 64-bit monotonic clock reading, nanoseconds since process start.
	wall uint64
	ext  int64
}

const (
	hasMonotonic uint64 = 1 << 63
	// maxWall      = wallToInternal + ((1 << 33) - 1) // year 2157
	minWall   = wallToInternal // year 1885
	nsecMask  = (1 << 30) - 1
	nsecShift = 30
)

// These helpers for manipulating the wall and monotonic clock readings
// take pointer receivers, even when they don't modify the time,
// to make them cheaper to call.

// IsZero reports whether t represents the zero time instant,
// January 1, year 1, 00:00:00 UTC.
func (t Time) IsZero() bool {
	// If hasMonotonic is set in t.wall, then the time can't be before 1885,
	// so it can't be the year 1.
	// If hasMonotonic is zero, then all the bits in wall other than the
	// nanoseconds field should be 0.
	// So if there are no nanoseconds then t.wall == 0, and if there are
	// no seconds then t.ext == 0.
	// This is equivalent to t.sec() == 0 && t.nsec() == 0, but is more efficient.
	return t.wall == 0 && t.ext == 0
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	if (t.wall & u.wall & hasMonotonic) != 0 {
		return t.ext > u.ext
	}
	ts := t.sec()
	us := u.sec()
	return ts > us || (ts == us && t.nsec() > u.nsec())
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	if (t.wall & u.wall & hasMonotonic) != 0 {
		return t.ext < u.ext
	}
	ts := t.sec()
	us := u.sec()
	return ts < us || (ts == us && t.nsec() < u.nsec())
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t Time) Compare(u Time) int {
	var tc, uc int64
	if (t.wall & u.wall & hasMonotonic) != 0 {
		tc, uc = t.ext, u.ext
	} else {
		tc, uc = t.sec(), u.sec()
		if tc == uc {
			tc, uc = int64(t.nsec()), int64(u.nsec())
		}
	}
	if tc < uc {
		return -1
	} else if tc > uc {
		return +1
	}
	return 0
}

// Equal reports whether t and u represent the same time instant.
// Unlike the == operator, Equal ignores monotonic clock readings.
func (t Time) Equal(u Time) bool {
	if (t.wall & u.wall & hasMonotonic) != 0 {
		return t.ext == u.ext
	}
	return t.sec() == u.sec() && t.nsec() == u.nsec()
}

// nsec returns the time's nanoseconds.
func (t *Time) nsec() int32 {
	return int32(t.wall & nsecMask)
}

// sec returns the time's seconds since Jan 1 year 1.
func (t *Time) sec() int64 {
	if (t.wall & hasMonotonic) != 0 {
		return wallToInternal + int64((t.wall<<1)>>(nsecShift+1))
	}
	return t.ext
}

// stripMono strips the monotonic clock reading in t.
func (t *Time) stripMono() {
	if (t.wall & hasMonotonic) != 0 {
		t.ext = t.sec()
		t.wall &= nsecMask
	}
}

// Sleep pauses the current thread for at least the duration d.
// A negative or zero duration causes Sleep to return immediately.
//
// In a freestanding environment, Sleep waits through the so_time_sleep
// hook of the target.
func Sleep(d Duration) {
	if d <= 0 {
		return
	}
	time_sleep(int64(d))
}

// norm returns nhi, nlo such that
//
//	hi * base + lo == nhi * base + nlo
//	0 <= nlo < base
func norm(hi, lo, base int) (int, int) {
	if lo < 0 {
		n := (-lo-1)/base + 1
		hi -= n
		lo += n * base
	}
	if lo >= base {
		n := lo / base
		hi += n
		lo -= n * base
	}
	return hi, lo
}
