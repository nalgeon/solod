// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"testing"

	. "github.com/nalgeon/solod/so/time"
)

var durationTests = []struct {
	str string
	d   Duration
}{
	{"0s", 0},
	{"1ns", 1 * Nanosecond},
	{"1.1µs", 1100 * Nanosecond},
	{"2.2ms", 2200 * Microsecond},
	{"3.3s", 3300 * Millisecond},
	{"4m5s", 4*Minute + 5*Second},
	{"4m5.001s", 4*Minute + 5001*Millisecond},
	{"5h6m7.001s", 5*Hour + 6*Minute + 7001*Millisecond},
	{"8m0.000000001s", 8*Minute + 1*Nanosecond},
	{"2562047h47m16.854775807s", 1<<63 - 1},
	{"-2562047h47m16.854775808s", -1 << 63},
}

func TestDurationString(t *testing.T) {
	var buf [32]byte
	for _, tt := range durationTests {
		if str := tt.d.String(buf); str != tt.str {
			t.Errorf("Duration(%d).String() = %s, want %s", int64(tt.d), str, tt.str)
		}
		if tt.d > 0 {
			if str := (-tt.d).String(buf); str != "-"+tt.str {
				t.Errorf("Duration(%d).String() = %s, want %s", int64(-tt.d), str, "-"+tt.str)
			}
		}
	}
}

var nsDurationTests = []struct {
	d    Duration
	want int64
}{
	{Duration(-1000), -1000},
	{Duration(-1), -1},
	{Duration(1), 1},
	{Duration(1000), 1000},
}

func TestDurationNanoseconds(t *testing.T) {
	for _, tt := range nsDurationTests {
		if got := tt.d.Nanoseconds(); got != tt.want {
			t.Errorf("Duration(%v).Nanoseconds() = %d; want: %d", tt.d, got, tt.want)
		}
	}
}

var usDurationTests = []struct {
	d    Duration
	want int64
}{
	{Duration(-1000), -1},
	{Duration(1000), 1},
}

func TestDurationMicroseconds(t *testing.T) {
	for _, tt := range usDurationTests {
		if got := tt.d.Microseconds(); got != tt.want {
			t.Errorf("Duration(%v).Microseconds() = %d; want: %d", tt.d, got, tt.want)
		}
	}
}

var msDurationTests = []struct {
	d    Duration
	want int64
}{
	{Duration(-1000000), -1},
	{Duration(1000000), 1},
}

func TestDurationMilliseconds(t *testing.T) {
	for _, tt := range msDurationTests {
		if got := tt.d.Milliseconds(); got != tt.want {
			t.Errorf("Duration(%v).Milliseconds() = %d; want: %d", tt.d, got, tt.want)
		}
	}
}

var secDurationTests = []struct {
	d    Duration
	want float64
}{
	{Duration(300000000), 0.3},
}

func TestDurationSeconds(t *testing.T) {
	for _, tt := range secDurationTests {
		if got := tt.d.Seconds(); got != tt.want {
			t.Errorf("Duration(%v).Seconds() = %g; want: %g", tt.d, got, tt.want)
		}
	}
}

var minDurationTests = []struct {
	d    Duration
	want float64
}{
	{Duration(-60000000000), -1},
	{Duration(-1), -1 / 60e9},
	{Duration(1), 1 / 60e9},
	{Duration(60000000000), 1},
	{Duration(3000), 5e-8},
}

func TestDurationMinutes(t *testing.T) {
	for _, tt := range minDurationTests {
		if got := tt.d.Minutes(); got != tt.want {
			t.Errorf("Duration(%v).Minutes() = %g; want: %g", tt.d, got, tt.want)
		}
	}
}

var hourDurationTests = []struct {
	d    Duration
	want float64
}{
	{Duration(-3600000000000), -1},
	{Duration(-1), -1 / 3600e9},
	{Duration(1), 1 / 3600e9},
	{Duration(3600000000000), 1},
	{Duration(36), 1e-11},
}

func TestDurationHours(t *testing.T) {
	for _, tt := range hourDurationTests {
		if got := tt.d.Hours(); got != tt.want {
			t.Errorf("Duration(%v).Hours() = %g; want: %g", tt.d, got, tt.want)
		}
	}
}

var durationTruncateTests = []struct {
	d    Duration
	m    Duration
	want Duration
}{
	{0, Second, 0},
	{Minute, -7 * Second, Minute},
	{Minute, 0, Minute},
	{Minute, 1, Minute},
	{Minute + 10*Second, 10 * Second, Minute + 10*Second},
	{2*Minute + 10*Second, Minute, 2 * Minute},
	{10*Minute + 10*Second, 3 * Minute, 9 * Minute},
	{Minute + 10*Second, Minute + 10*Second + 1, 0},
	{Minute + 10*Second, Hour, 0},
	{-Minute, Second, -Minute},
	{-10 * Minute, 3 * Minute, -9 * Minute},
	{-10 * Minute, Hour, 0},
}

func TestDurationTruncate(t *testing.T) {
	for _, tt := range durationTruncateTests {
		if got := tt.d.Truncate(tt.m); got != tt.want {
			t.Errorf("Duration(%v).Truncate(%v) = %v; want: %v", tt.d, tt.m, got, tt.want)
		}
	}
}

var durationRoundTests = []struct {
	d    Duration
	m    Duration
	want Duration
}{
	{0, Second, 0},
	{Minute, -11 * Second, Minute},
	{Minute, 0, Minute},
	{Minute, 1, Minute},
	{2 * Minute, Minute, 2 * Minute},
	{2*Minute + 10*Second, Minute, 2 * Minute},
	{2*Minute + 30*Second, Minute, 3 * Minute},
	{2*Minute + 50*Second, Minute, 3 * Minute},
	{-Minute, 1, -Minute},
	{-2 * Minute, Minute, -2 * Minute},
	{-2*Minute - 10*Second, Minute, -2 * Minute},
	{-2*Minute - 30*Second, Minute, -3 * Minute},
	{-2*Minute - 50*Second, Minute, -3 * Minute},
	{8e18, 3e18, 9e18},
	{9e18, 5e18, 1<<63 - 1},
	{-8e18, 3e18, -9e18},
	{-9e18, 5e18, -1 << 63},
	{3<<61 - 1, 3 << 61, 3 << 61},
}

func TestDurationRound(t *testing.T) {
	for _, tt := range durationRoundTests {
		if got := tt.d.Round(tt.m); got != tt.want {
			t.Errorf("Duration(%v).Round(%v) = %v; want: %v", tt.d, tt.m, got, tt.want)
		}
	}
}

var durationAbsTests = []struct {
	d    Duration
	want Duration
}{
	{0, 0},
	{1, 1},
	{-1, 1},
	{1 * Minute, 1 * Minute},
	{-1 * Minute, 1 * Minute},
	{minDuration, maxDuration},
	{minDuration + 1, maxDuration},
	{minDuration + 2, maxDuration - 1},
	{maxDuration, maxDuration},
	{maxDuration - 1, maxDuration - 1},
}

func TestDurationAbs(t *testing.T) {
	for _, tt := range durationAbsTests {
		if got := tt.d.Abs(); got != tt.want {
			t.Errorf("Duration(%v).Abs() = %v; want: %v", tt.d, got, tt.want)
		}
	}
}
