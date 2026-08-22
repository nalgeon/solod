// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/runtime"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// All tests use variants of 2024-03-15T14:30:45Z as the input time.

func TestParse_RFC3339(t *testing.T) {
	tm, err := time.Parse(time.RFC3339, "2024-03-15T14:30:45Z", 0)
	if err != nil {
		t.Fatal("unexpected Parse RFC3339 error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse RFC3339 date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 14 || clock.Minute != 30 || clock.Second != 45 {
		t.Error("unexpected Parse RFC3339 clock")
	}
}

func TestParse_RFC3339Nano(t *testing.T) {
	tm, err := time.Parse(time.RFC3339Nano, "2024-03-15T14:30:45.123456789Z", 0)
	if err != nil {
		t.Fatal("unexpected Parse RFC3339Nano error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse RFC3339Nano date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 14 || clock.Minute != 30 || clock.Second != 45 {
		t.Error("unexpected Parse RFC3339Nano clock")
	}
	if tm.Nanosecond() != 123456789 {
		t.Error("unexpected Parse RFC3339Nano nanosecond")
	}
}

func TestParse_RFC3339PosOffset(t *testing.T) {
	// 14:30:45+05:00 is 09:30:45 UTC.
	tm, err := time.Parse(time.RFC3339, "2024-03-15T14:30:45+05:00", 0)
	if err != nil {
		t.Fatal("unexpected Parse RFC3339+offset error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse RFC3339+offset date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 9 || clock.Minute != 30 || clock.Second != 45 {
		t.Error("unexpected Parse RFC3339+offset clock")
	}
}

func TestParse_RFC3339NegOffset(t *testing.T) {
	// 14:30:45-03:00 is 17:30:45 UTC.
	tm, err := time.Parse(time.RFC3339, "2024-03-15T14:30:45-03:00", 0)
	if err != nil {
		t.Fatal("unexpected Parse RFC3339-offset error")
		return
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 17 || clock.Minute != 30 || clock.Second != 45 {
		t.Error("unexpected Parse RFC3339-offset clock")
	}
}

func TestParse_RFC3339NanoOffset(t *testing.T) {
	// 14:30:45+05:30 is 09:00:45 UTC.
	tm, err := time.Parse(time.RFC3339Nano, "2024-03-15T14:30:45.123456789+05:30", 0)
	if err != nil {
		t.Fatal("unexpected Parse RFC3339Nano+offset error")
		return
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 9 || clock.Minute != 0 || clock.Second != 45 {
		t.Error("unexpected Parse RFC3339Nano+offset clock")
	}
	if tm.Nanosecond() != 123456789 {
		t.Error("unexpected Parse RFC3339Nano+offset nanosecond")
	}
}

func TestParse_DateTime(t *testing.T) {
	tm, err := time.Parse(time.DateTime, "2024-03-15 14:30:45", time.UTC)
	if err != nil {
		t.Fatal("unexpected Parse DateTime error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse DateTime date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 14 || clock.Minute != 30 || clock.Second != 45 {
		t.Error("unexpected Parse DateTime clock")
	}
}

func TestParse_DateTimeOffset(t *testing.T) {
	// 14:30:45+05:30 is 09:00:45 UTC.
	offset := time.Offset(5*3600 + 30*60) // UTC+5:30
	tm, err := time.Parse(time.DateTime, "2024-03-15 14:30:45", offset)
	if err != nil {
		t.Fatal("unexpected Parse DateTime+offset error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse DateTime+offset date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 9 || clock.Minute != 0 || clock.Second != 45 {
		t.Error("unexpected Parse DateTime+offset clock")
	}
}

func TestParse_DateOnly(t *testing.T) {
	tm, err := time.Parse(time.DateOnly, "2024-03-15", time.UTC)
	if err != nil {
		t.Fatal("unexpected Parse DateOnly error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse DateOnly date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 0 || clock.Minute != 0 || clock.Second != 0 {
		t.Error("unexpected Parse DateOnly clock")
	}
}

func TestParse_TimeOnly(t *testing.T) {
	tm, err := time.Parse(time.TimeOnly, "14:30:45", time.UTC)
	if err != nil {
		t.Fatal("unexpected Parse TimeOnly error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 0 || date.Month != time.January || date.Day != 1 {
		t.Error("unexpected Parse TimeOnly date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 14 || clock.Minute != 30 || clock.Second != 45 {
		t.Error("unexpected Parse TimeOnly clock")
	}
}

func TestParse_Custom(t *testing.T) {
	if !runtime.Hosted || runtime.GOOS == "windows" {
		t.Skip("a custom layout needs a POSIX environment")
		return
	}
	tm, err := time.Parse("%d.%m.%Y", "15.03.2024", time.UTC)
	if err != nil {
		t.Fatal("unexpected Parse custom error")
		return
	}
	date := tm.Date(time.UTC)
	if date.Year != 2024 || date.Month != time.March || date.Day != 15 {
		t.Error("unexpected Parse custom date")
	}
	clock := tm.Clock(time.UTC)
	if clock.Hour != 0 || clock.Minute != 0 || clock.Second != 0 {
		t.Error("unexpected Parse custom clock")
	}
}

// A badCase is a layout and a value Parse must reject.
type badCase struct {
	layout string
	value  string
	why    string
}

// badOffsetCases hold the lengths parseRFC3339 and parseRFC3339Nano accept,
// with an offset field that does not fill the rest of the value.
var badOffsetCases = []badCase{
	{time.RFC3339, "2024-03-15T14:30:45+", "a sign with no offset digits"},
	{time.RFC3339, "2024-03-15T14:30:45-", "a sign with no offset digits"},
	{time.RFC3339, "2024-03-15T14:30:45*", "an offset that is neither Z nor a sign"},
	{time.RFC3339, "2024-03-15T14:30:45Zxxxxx", "text after Z"},
	{time.RFC3339, "2024-03-15T14:30:45+0500x", "an offset with no colon"},
	{time.RFC3339Nano, "2024-03-15T14:30:45.123456789+", "a sign with no offset digits"},
	{time.RFC3339Nano, "2024-03-15T14:30:45.123456789-", "a sign with no offset digits"},
	{time.RFC3339Nano, "2024-03-15T14:30:45.123456789Zxxxxx", "text after Z"},
	{time.RFC3339Nano, "2024-03-15T14:30:45.123456789+0500x", "an offset with no colon"},
}

func TestParse_BadOffset(t *testing.T) {
	// Check that Parse rejects a truncated or overlong offset field.
	// Parse reads the offset at a fixed position, so it must not read past
	// the end of the value.
	for _, c := range badOffsetCases {
		_, err := time.Parse(c.layout, c.value, time.UTC)
		if err == nil {
			t.Errorf("Parse(%s) accepted %s", c.value, c.why)
		}
	}
}

func TestParse_Error(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("a custom layout needs a hosted environment")
		return
	}
	_, err := time.Parse("%Y-%m-%d", "not-a-date", time.UTC)
	if err == nil {
		t.Error("expected Parse error")
	}
}
