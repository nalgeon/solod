// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"testing"
	"testing/quick"

	. "github.com/nalgeon/solod/so/time"
)

func TestUnixUTC(t *testing.T) {
	for _, test := range utctests {
		sec := test.seconds
		golden := &test.golden
		tm := Unix(sec, 0).UTC()
		newsec := tm.Unix()
		if newsec != sec {
			t.Errorf("Unix(%d, 0).Unix() = %d", sec, newsec)
		}
		if !same(tm, golden) {
			t.Errorf("Unix(%d, 0):  // %#v", sec, tm)
			t.Errorf("  want=%+v", *golden)
			t.Errorf("  have=%+v", tm)
		}
	}
}

func TestUnixNanoUTC(t *testing.T) {
	for _, test := range nanoutctests {
		golden := &test.golden
		nsec := test.seconds*1e9 + int64(golden.Nanosecond)
		tm := Unix(0, nsec).UTC()
		newnsec := tm.Unix()*1e9 + int64(tm.Nanosecond())
		if newnsec != nsec {
			t.Errorf("Unix(0, %d).Nanoseconds() = %d", nsec, newnsec)
		}
		if !same(tm, golden) {
			t.Errorf("Unix(0, %d):", nsec)
			t.Errorf("  want=%+v", *golden)
			t.Errorf("  have=%+v", tm)
		}
	}
}

func TestUnixUTCAndBack(t *testing.T) {
	f := func(sec int64) bool { return Unix(sec, 0).UTC().Unix() == sec }
	f32 := func(sec int32) bool { return f(int64(sec)) }
	cfg := &quick.Config{MaxCount: 10000}

	// Try a reasonable date first, then the huge ones.
	if err := quick.Check(f32, cfg); err != nil {
		t.Fatal(err)
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestUnixNanoUTCAndBack(t *testing.T) {
	f := func(nsec int64) bool {
		t := Unix(0, nsec).UTC()
		ns := t.Unix()*1e9 + int64(t.Nanosecond())
		return ns == nsec
	}
	f32 := func(nsec int32) bool { return f(int64(nsec)) }
	cfg := &quick.Config{MaxCount: 10000}

	// Try a small date first, then the large ones. (The span is only a few hundred years
	// for nanoseconds in an int64.)
	if err := quick.Check(f32, cfg); err != nil {
		t.Fatal(err)
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestUnixMilli(t *testing.T) {
	f := func(msec int64) bool {
		t := UnixMilli(msec)
		return t.UnixMilli() == msec
	}
	cfg := &quick.Config{MaxCount: 10000}
	if err := quick.Check(f, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestUnixMicro(t *testing.T) {
	f := func(usec int64) bool {
		t := UnixMicro(usec)
		return t.UnixMicro() == usec
	}
	cfg := &quick.Config{MaxCount: 10000}
	if err := quick.Check(f, cfg); err != nil {
		t.Fatal(err)
	}
}

func same(t Time, u *parsedTime) bool {
	// Check aggregates.
	date := t.Date()
	clock := t.Clock()
	name, offset := t.Zone()
	if date.Year != u.Year || date.Month != u.Month || date.Day != u.Day ||
		clock.Hour != u.Hour || clock.Minute != u.Minute || clock.Second != u.Second ||
		name != u.Zone || offset != u.ZoneOffset {
		return false
	}
	// Check individual entries.
	return t.Year() == u.Year &&
		t.Month() == u.Month &&
		t.Day() == u.Day &&
		t.Hour() == u.Hour &&
		t.Minute() == u.Minute &&
		t.Second() == u.Second &&
		t.Nanosecond() == u.Nanosecond &&
		t.Weekday() == u.Weekday
}
