// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

func TestDate(t *testing.T) {
	tm := time.Date(2021, time.May, 10, 12, 33, 44, 777888999, time.UTC)
	if tm.Year() != 2021 {
		t.Error("unexpected Time.Year")
	}
	if tm.Month() != time.May {
		t.Error("unexpected Time.Month")
	}
	if tm.Day() != 10 {
		t.Error("unexpected Time.Day")
	}
	if tm.Hour() != 12 {
		t.Error("unexpected Time.Hour")
	}
	if tm.Minute() != 33 {
		t.Error("unexpected Time.Minute")
	}
	if tm.Second() != 44 {
		t.Error("unexpected Time.Second")
	}
	if tm.Nanosecond() != 777888999 {
		t.Error("unexpected Time.Nanosecond")
	}
}

// A dateCase is the argument list of Date and the Unix time it denotes.
type dateCase struct {
	year, month, day     int
	hour, minute, second int
	nsec                 int
	offset               time.Offset
	unixSec              int64
}

var dateCases = []dateCase{
	{2011, 11, 6, 8, 0, 0, 0, time.UTC, 1320566400},   // 8:00:00 UTC
	{2011, 11, 6, 8, 59, 59, 0, time.UTC, 1320569999}, // 8:59:59 UTC
	{2011, 11, 6, 10, 0, 0, 0, time.UTC, 1320573600},  // 10:00:00 UTC

	{2011, 3, 13, 9, 0, 0, 0, time.UTC, 1300006800},   // 9:00:00 UTC
	{2011, 3, 13, 9, 59, 59, 0, time.UTC, 1300010399}, // 9:59:59 UTC
	{2011, 3, 13, 10, 0, 0, 0, time.UTC, 1300010400},  // 10:00:00 UTC
	{2011, 3, 13, 9, 30, 0, 0, time.UTC, 1300008600},  // 9:30:00 UTC
	{2012, 12, 24, 8, 0, 0, 0, time.UTC, 1356336000},  // leap year

	// Many names for 2011-11-18 15:56:35.0 UTC.
	{2011, 11, 18, 15, 56, 35, 0, time.UTC, 1321631795},               // Nov 18 15:56:35
	{2011, 11, 19, -9, 56, 35, 0, time.UTC, 1321631795},               // Nov 19 -9:56:35
	{2011, 11, 17, 39, 56, 35, 0, time.UTC, 1321631795},               // Nov 17 39:56:35
	{2011, 11, 18, 14, 116, 35, 0, time.UTC, 1321631795},              // Nov 18 14:116:35
	{2011, 10, 49, 15, 56, 35, 0, time.UTC, 1321631795},               // Oct 49 15:56:35
	{2011, 11, 18, 15, 55, 95, 0, time.UTC, 1321631795},               // Nov 18 15:55:95
	{2011, 11, 18, 15, 56, 34, 1e9, time.UTC, 1321631795},             // Nov 18 15:56:34 + 10^9ns
	{2011, 12, -12, 15, 56, 35, 0, time.UTC, 1321631795},              // Dec -12 15:56:35
	{2012, 1, -43, 15, 56, 35, 0, time.UTC, 1321631795},               // 2012 Jan -43 15:56:35
	{2012, -1, 18, 15, 56, 35, 0, time.UTC, 1321631795},               // 2012 (Jan-2) 18 15:56:35
	{2010, 23, 18, 15, 56, 35, 0, time.UTC, 1321631795},               // 2010 (Dec+11) 18 15:56:35
	{1970, 1, 15297, 15, 56, 35, 0, time.UTC, 1321631795},             // large number of days
	{2011, 11, 18, 10, 56, 35, 0, time.Offset(-5 * 3600), 1321631795}, // UTC-5
	{2011, 11, 18, 3, 56, 35, 0, time.Offset(-12 * 3600), 1321631795}, // UTC-12
	{2011, 11, 18, 16, 56, 35, 0, time.Offset(1 * 3600), 1321631795},  // UTC+1
	{2011, 11, 19, 3, 56, 35, 0, time.Offset(12 * 3600), 1321631795},  // UTC+12

	{1970, 1, -25508, 8, 0, 0, 0, time.UTC, -2203948800}, // negative Unix time
}

func TestDateUnix(t *testing.T) {
	for _, c := range dateCases {
		tm := time.Date(c.year, time.Month(c.month), c.day, c.hour, c.minute, c.second, c.nsec, c.offset)
		want := time.Unix(c.unixSec, 0)
		if !tm.Equal(want) {
			t.Errorf("Date(%d, %d, %d, %d, %d, %d, %d, %d).Unix() = %d, want %d",
				c.year, c.month, c.day, c.hour, c.minute, c.second, c.nsec, int(c.offset),
				tm.Unix(), c.unixSec)
		}
	}
}

func TestZeroTime(t *testing.T) {
	var zero time.Time
	date := zero.Date(time.UTC)
	clock := zero.Clock(time.UTC)
	if date.Year != 1 || date.Month != time.January || date.Day != 1 {
		t.Errorf("zero Time date = %d-%02d-%02d, want 0001-01-01",
			date.Year, int(date.Month), date.Day)
	}
	if clock.Hour != 0 || clock.Minute != 0 || clock.Second != 0 {
		t.Errorf("zero Time clock = %02d:%02d:%02d, want 00:00:00",
			clock.Hour, clock.Minute, clock.Second)
	}
	if zero.Nanosecond() != 0 || zero.YearDay() != 1 || zero.Weekday() != time.Monday {
		t.Errorf("zero Time nsec %d, yday %d, weekday %d, want 0, 1, %d",
			zero.Nanosecond(), zero.YearDay(), int(zero.Weekday()), int(time.Monday))
	}
	if !zero.IsZero() {
		t.Error("zero Time IsZero() = false, want true")
	}

	// Two zero times behave the same way, and the methods that need no other
	// time hold their documented values.
	var other time.Time
	if zero.After(other) || zero.Before(other) || !zero.Equal(other) || zero.Compare(other) != 0 {
		t.Error("two zero Time values do not compare equal")
	}
	if zero.Sub(other) != 0 {
		t.Errorf("zero Time Sub() = %d, want 0", int64(zero.Sub(other)))
	}
	if !zero.Add(0).Equal(zero) || !zero.AddDate(0, 0, 0).Equal(zero) {
		t.Error("zero Time changed after adding nothing")
	}
	if !zero.Truncate(time.Hour).Equal(zero) || !zero.Round(time.Hour).Equal(zero) {
		t.Error("zero Time changed after rounding to the hour")
	}
	if zero.Unix() != -62135596800 {
		t.Errorf("zero Time Unix() = %d, want -62135596800", zero.Unix())
	}
	if zero.UnixMilli() != zero.Unix()*1e3 || zero.UnixMicro() != zero.Unix()*1e6 {
		t.Error("zero Time UnixMilli or UnixMicro does not match Unix")
	}
	if y, w := zero.ISOWeek(); y != 1 || w != 1 {
		t.Errorf("zero Time ISOWeek() = %d/%d, want 1/1", y, w)
	}
}

func TestCompare(t *testing.T) {
	early := time.Date(2011, time.November, 18, 15, 56, 35, 0, time.UTC)
	late := time.Date(2011, time.November, 18, 15, 56, 35, 1, time.UTC)
	same := time.Date(2011, time.November, 18, 15, 56, 35, 0, time.UTC)

	if !early.Before(late) || early.After(late) || early.Equal(late) {
		t.Error("an earlier Time does not compare before a later one")
	}
	if !late.After(early) || late.Before(early) {
		t.Error("a later Time does not compare after an earlier one")
	}
	if !early.Equal(same) || early.Before(same) || early.After(same) {
		t.Error("two equal Time values do not compare equal")
	}
	if early.Compare(late) != -1 || late.Compare(early) != 1 || early.Compare(same) != 0 {
		t.Errorf("Compare() = %d, %d, %d, want -1, 1, 0",
			early.Compare(late), late.Compare(early), early.Compare(same))
	}
}

func TestSinceUntil(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	if since := time.Since(past); since < time.Hour {
		t.Errorf("Since() = %d, want at least an hour", int64(since))
	}
	future := time.Now().Add(time.Hour)
	if until := time.Until(future); until > time.Hour {
		t.Errorf("Until() = %d, want at most an hour", int64(until))
	}
}

func TestNow(t *testing.T) {
	tm := time.Now()
	if tm.IsZero() {
		t.Error("unexpected Time.IsZero")
	}
}

func TestSleep(t *testing.T) {
	start := time.Now()
	time.Sleep(20 * time.Millisecond)
	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond {
		t.Error("Sleep returned before the duration elapsed")
	}
	// No upper bound is asserted: Sleep only guarantees a lower bound, and a
	// loaded machine (e.g. a CI runner) may schedule the thread much later.
}

func TestSleepNonPositive(t *testing.T) {
	start := time.Now()
	// Returns immediately without blocking.
	time.Sleep(0)
	time.Sleep(-1 * time.Second)
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Error("Sleep should return immediately")
	}
}
