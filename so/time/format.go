package time

import (
	"unsafe"

	"solod.dev/so/c"
)

// Commonly used layouts for Format and Parse.
const (
	RFC3339     = "%Y-%m-%dT%H:%M:%S%z"
	RFC3339Nano = "%Y-%m-%dT%H:%M:%S.%f%z"
	DateTime    = "%Y-%m-%d %H:%M:%S"
	DateOnly    = "%Y-%m-%d"
	TimeOnly    = "%H:%M:%S"
)

// Lengths of the common date/time layouts.
const (
	RFC3339Len     = 25 // 2006-01-02T15:04:05+07:00
	RFC3339NanoLen = 35 // 2006-01-02T15:04:05.999999999+07:00
	DateTimeLen    = 19 // 2006-01-02 15:04:05
	DateOnlyLen    = 10 // 2006-01-02
	TimeOnlyLen    = 8  // 15:04:05
)

// Format formats the time per layout (strftime verbs like "%Y-%m-%d"),
// writing into buf. Returns the formatted string (a view into buf).
// buf length must be large enough for the formatted output
// (see [RFC3339Len], etc. for common layouts).
// Returns "" for an empty buf.
func (t Time) Format(buf []byte, layout string, offset Offset) string {
	if len(buf) == 0 {
		return ""
	}
	sec := t.absSec() + absSeconds(offset)
	days := absSeconds_days(sec)
	clock := absSeconds_clock(sec)

	// Fast paths for known layouts - avoid strftime overhead.
	if layout == RFC3339 {
		date := absDays_date(days)
		n := fmtDate(buf, 0, date)
		buf[n] = 'T'
		n = fmtClock(buf, n+1, clock)
		n = fmtOffset(buf, n, offset)
		return string(buf[:n])
	}
	if layout == RFC3339Nano {
		date := absDays_date(days)
		n := fmtDate(buf, 0, date)
		buf[n] = 'T'
		n = fmtClock(buf, n+1, clock)
		buf[n] = '.'
		n = fmtNano(buf, n+1, int(t.nsec()))
		n = fmtOffset(buf, n, offset)
		return string(buf[:n])
	}
	if layout == DateTime {
		date := absDays_date(days)
		n := fmtDate(buf, 0, date)
		buf[n] = ' '
		n = fmtClock(buf, n+1, clock)
		return string(buf[:n])
	}
	if layout == DateOnly {
		date := absDays_date(days)
		n := fmtDate(buf, 0, date)
		return string(buf[:n])
	}
	if layout == TimeOnly {
		n := fmtClock(buf, 0, clock)
		return string(buf[:n])
	}

	// General case: strftime.
	date := absDays_date(days)
	split := absDays_split(days)
	janFeb := absYday_janFeb(split.ayday)
	wday := absDays_weekday(days)
	leap := absCentury_Leap(split.century, split.cyear)
	yday := absYday_yday(split.ayday, janFeb, leap)

	var tm time_tm
	tm.tm_year = c.Int(date.Year - 1900)
	tm.tm_mon = c.Int(int(date.Month) - 1)
	tm.tm_mday = c.Int(date.Day)
	tm.tm_hour = c.Int(clock.Hour)
	tm.tm_min = c.Int(clock.Minute)
	tm.tm_sec = c.Int(clock.Second)
	tm.tm_wday = c.Int(wday)
	tm.tm_yday = c.Int(yday - 1)
	tm.tm_isdst = 0

	bufp := (*c.Char)(unsafe.SliceData(buf))
	n := strftime(bufp, uintptr(len(buf)), layout, &tm)
	return string(buf[:n])
}

// String formats the time as ISO 8601 "2006-01-02T15:04:05Z",
// writing into buf. Returns the formatted string (a view into buf).
// buf length must be at least [RFC3339Len] bytes.
func (t Time) String(buf []byte) string {
	return t.Format(buf, RFC3339, UTC)
}

// fmtDate writes "YYYY-MM-DD" into buf at position i.
// Returns the position after the last byte written.
func fmtDate(buf []byte, i int, date CalDate) int {
	i = fmt4(buf, i, date.Year)
	buf[i] = '-'
	i = fmt2(buf, i+1, int(date.Month))
	buf[i] = '-'
	return fmt2(buf, i+1, date.Day)
}

// fmtClock writes "HH:MM:SS" into buf at position i.
// Returns the position after the last byte written.
func fmtClock(buf []byte, i int, clock CalClock) int {
	i = fmt2(buf, i, clock.Hour)
	buf[i] = ':'
	i = fmt2(buf, i+1, clock.Minute)
	buf[i] = ':'
	return fmt2(buf, i+1, clock.Second)
}

// fmtOffset writes "Z" (for UTC) or "+HH:MM"/"-HH:MM" into buf at position i.
func fmtOffset(buf []byte, i int, offset Offset) int {
	if offset == UTC {
		buf[i] = 'Z'
		return i + 1
	}
	off := int(offset)
	if off < 0 {
		buf[i] = '-'
		off = -off
	} else {
		buf[i] = '+'
	}
	i = fmt2(buf, i+1, off/3600)
	buf[i] = ':'
	return fmt2(buf, i+1, (off%3600)/60)
}

// fmtNano writes a 9-digit zero-padded nanosecond value into buf at position i.
func fmtNano(buf []byte, i int, ns int) int {
	for j := 8; j >= 0; j-- {
		buf[i+j] = byte('0' + ns%10)
		ns /= 10
	}
	return i + 9
}

// fmt2 writes a 2-digit zero-padded number into buf at position i.
func fmt2(buf []byte, i int, v int) int {
	buf[i] = byte('0' + v/10)
	buf[i+1] = byte('0' + v%10)
	return i + 2
}

// fmt4 writes a 4-digit zero-padded number into buf at position i.
func fmt4(buf []byte, i int, v int) int {
	buf[i] = byte('0' + v/1000)
	buf[i+1] = byte('0' + (v/100)%10)
	buf[i+2] = byte('0' + (v/10)%10)
	buf[i+3] = byte('0' + v%10)
	return i + 4
}
