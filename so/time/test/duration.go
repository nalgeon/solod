// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// A durationCase is a duration and its string form. neg holds the string form
// of the negated duration, and is empty when the duration is not positive.
type durationCase struct {
	str string
	neg string
	d   time.Duration
}

var durationCases = []durationCase{
	{"0s", "", 0},
	{"1ns", "-1ns", 1 * time.Nanosecond},
	{"1.1µs", "-1.1µs", 1100 * time.Nanosecond},
	{"2.2ms", "-2.2ms", 2200 * time.Microsecond},
	{"3.3s", "-3.3s", 3300 * time.Millisecond},
	{"4m5s", "-4m5s", 4*time.Minute + 5*time.Second},
	{"4m5.001s", "-4m5.001s", 4*time.Minute + 5001*time.Millisecond},
	{"5h6m7.001s", "-5h6m7.001s", 5*time.Hour + 6*time.Minute + 7001*time.Millisecond},
	{"8m0.000000001s", "-8m0.000000001s", 8*time.Minute + 1*time.Nanosecond},
	{"2562047h47m16.854775807s", "-2562047h47m16.854775807s", maxDuration},
	{"-2562047h47m16.854775808s", "", minDuration},
}

func TestDurationString(t *testing.T) {
	buf := make([]byte, time.MaxDurationLen)
	for _, c := range durationCases {
		if str := c.d.String(buf); str != c.str {
			t.Errorf("Duration(%d).String() = %s, want %s", int64(c.d), str, c.str)
		}
		if c.d <= 0 {
			continue
		}
		if str := (-c.d).String(buf); str != c.neg {
			t.Errorf("Duration(%d).String() = %s, want %s", int64(-c.d), str, c.neg)
		}
	}
}

// An intUnitCase is a duration and its count in a unit that divides it.
type intUnitCase struct {
	d    time.Duration
	want int64
}

var nsCases = []intUnitCase{
	{time.Duration(-1000), -1000},
	{time.Duration(-1), -1},
	{time.Duration(1), 1},
	{time.Duration(1000), 1000},
}

func TestDurationNanoseconds(t *testing.T) {
	for _, c := range nsCases {
		if got := c.d.Nanoseconds(); got != c.want {
			t.Errorf("Duration(%d).Nanoseconds() = %d, want %d", int64(c.d), got, c.want)
		}
	}
}

var usCases = []intUnitCase{
	{time.Duration(-1000), -1},
	{time.Duration(1000), 1},
}

func TestDurationMicroseconds(t *testing.T) {
	for _, c := range usCases {
		if got := c.d.Microseconds(); got != c.want {
			t.Errorf("Duration(%d).Microseconds() = %d, want %d", int64(c.d), got, c.want)
		}
	}
}

var msCases = []intUnitCase{
	{time.Duration(-1000000), -1},
	{time.Duration(1000000), 1},
}

func TestDurationMilliseconds(t *testing.T) {
	for _, c := range msCases {
		if got := c.d.Milliseconds(); got != c.want {
			t.Errorf("Duration(%d).Milliseconds() = %d, want %d", int64(c.d), got, c.want)
		}
	}
}

// A floatUnitCase is a duration and its count in a unit larger than it.
type floatUnitCase struct {
	d    time.Duration
	want float64
}

var secCases = []floatUnitCase{
	{time.Duration(300000000), 0.3},
}

func TestDurationSeconds(t *testing.T) {
	for _, c := range secCases {
		if got := c.d.Seconds(); got != c.want {
			t.Errorf("Duration(%d).Seconds() = %g, want %g", int64(c.d), got, c.want)
		}
	}
}

var minCases = []floatUnitCase{
	{time.Duration(-60000000000), -1},
	{time.Duration(-1), -1 / 60e9},
	{time.Duration(1), 1 / 60e9},
	{time.Duration(60000000000), 1},
	{time.Duration(3000), 5e-8},
}

func TestDurationMinutes(t *testing.T) {
	for _, c := range minCases {
		if got := c.d.Minutes(); got != c.want {
			t.Errorf("Duration(%d).Minutes() = %g, want %g", int64(c.d), got, c.want)
		}
	}
}

var hourCases = []floatUnitCase{
	{time.Duration(-3600000000000), -1},
	{time.Duration(-1), -1 / 3600e9},
	{time.Duration(1), 1 / 3600e9},
	{time.Duration(3600000000000), 1},
	{time.Duration(36), 1e-11},
}

func TestDurationHours(t *testing.T) {
	for _, c := range hourCases {
		if got := c.d.Hours(); got != c.want {
			t.Errorf("Duration(%d).Hours() = %g, want %g", int64(c.d), got, c.want)
		}
	}
}

// A roundCase is a duration, a multiple to round it to, and the result.
type roundCase struct {
	d    time.Duration
	m    time.Duration
	want time.Duration
}

var truncateCases = []roundCase{
	{0, time.Second, 0},
	{time.Minute, -7 * time.Second, time.Minute},
	{time.Minute, 0, time.Minute},
	{time.Minute, 1, time.Minute},
	{time.Minute + 10*time.Second, 10 * time.Second, time.Minute + 10*time.Second},
	{2*time.Minute + 10*time.Second, time.Minute, 2 * time.Minute},
	{10*time.Minute + 10*time.Second, 3 * time.Minute, 9 * time.Minute},
	{time.Minute + 10*time.Second, time.Minute + 10*time.Second + 1, 0},
	{time.Minute + 10*time.Second, time.Hour, 0},
	{-time.Minute, time.Second, -time.Minute},
	{-10 * time.Minute, 3 * time.Minute, -9 * time.Minute},
	{-10 * time.Minute, time.Hour, 0},
}

func TestDurationTruncate(t *testing.T) {
	for _, c := range truncateCases {
		if got := c.d.Truncate(c.m); got != c.want {
			t.Errorf("Duration(%d).Truncate(%d) = %d, want %d",
				int64(c.d), int64(c.m), int64(got), int64(c.want))
		}
	}
}

var roundCases = []roundCase{
	{0, time.Second, 0},
	{time.Minute, -11 * time.Second, time.Minute},
	{time.Minute, 0, time.Minute},
	{time.Minute, 1, time.Minute},
	{2 * time.Minute, time.Minute, 2 * time.Minute},
	{2*time.Minute + 10*time.Second, time.Minute, 2 * time.Minute},
	{2*time.Minute + 30*time.Second, time.Minute, 3 * time.Minute},
	{2*time.Minute + 50*time.Second, time.Minute, 3 * time.Minute},
	{-time.Minute, 1, -time.Minute},
	{-2 * time.Minute, time.Minute, -2 * time.Minute},
	{-2*time.Minute - 10*time.Second, time.Minute, -2 * time.Minute},
	{-2*time.Minute - 30*time.Second, time.Minute, -3 * time.Minute},
	{-2*time.Minute - 50*time.Second, time.Minute, -3 * time.Minute},
	{8e18, 3e18, 9e18},
	{9e18, 5e18, maxDuration},
	{-8e18, 3e18, -9e18},
	{-9e18, 5e18, minDuration},
	{3<<61 - 1, 3 << 61, 3 << 61},
}

func TestDurationRound(t *testing.T) {
	for _, c := range roundCases {
		if got := c.d.Round(c.m); got != c.want {
			t.Errorf("Duration(%d).Round(%d) = %d, want %d",
				int64(c.d), int64(c.m), int64(got), int64(c.want))
		}
	}
}

// An absCase is a duration and its absolute value.
type absCase struct {
	d    time.Duration
	want time.Duration
}

var absCases = []absCase{
	{0, 0},
	{1, 1},
	{-1, 1},
	{1 * time.Minute, 1 * time.Minute},
	{-1 * time.Minute, 1 * time.Minute},
	{minDuration, maxDuration},
	{minDuration + 1, maxDuration},
	{minDuration + 2, maxDuration - 1},
	{maxDuration, maxDuration},
	{maxDuration - 1, maxDuration - 1},
}

func TestDurationAbs(t *testing.T) {
	for _, c := range absCases {
		if got := c.d.Abs(); got != c.want {
			t.Errorf("Duration(%d).Abs() = %d, want %d", int64(c.d), int64(got), int64(c.want))
		}
	}
}
