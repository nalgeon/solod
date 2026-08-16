// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// randCount is the number of random values each randomized test draws.
const randCount = 10000

// nanosOf returns the nanosecond count of tm since the Unix epoch. It repeats
// the calculation of UnixNano instead of calling it, so that the round trip
// checks the seconds and the nanoseconds apart. The calculation runs in uint64:
// an instant outside the range overflows, and C does not define a signed
// overflow.
func nanosOf(tm time.Time) int64 {
	return int64(uint64(tm.Unix())*1e9 + uint64(tm.Nanosecond()))
}

func TestUnixUTC(t *testing.T) {
	for _, c := range utcCases {
		tm := time.Unix(c.seconds, 0)
		if got := tm.Unix(); got != c.seconds {
			t.Errorf("Unix(%d, 0).Unix() = %d, want %d", c.seconds, got, c.seconds)
		}
		if !sameTime(tm, c) {
			t.Errorf("Unix(%d, 0) is not %d-%02d-%02d %02d:%02d:%02d",
				c.seconds, c.year, int(c.month), c.day, c.hour, c.minute, c.second)
		}
	}
}

func TestUnixNanoUTC(t *testing.T) {
	for _, c := range nanoUTCCases {
		nsec := c.seconds*1e9 + int64(c.nsec)
		tm := time.Unix(0, nsec)
		if got := nanosOf(tm); got != nsec {
			t.Errorf("Unix(0, %d) nanoseconds = %d, want %d", nsec, got, nsec)
		}
		if !sameTime(tm, c) {
			t.Errorf("Unix(0, %d) is not %d-%02d-%02d %02d:%02d:%02d.%09d",
				nsec, c.year, int(c.month), c.day, c.hour, c.minute, c.second, c.nsec)
		}
	}
}

func TestUnixUTCAndBack(t *testing.T) {
	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	// Try reasonable dates first, then the huge ones.
	for range randCount {
		sec := int64(r.Int32())
		if got := time.Unix(sec, 0).Unix(); got != sec {
			t.Errorf("Unix(%d, 0).Unix() = %d", sec, got)
			return
		}
	}
	for range randCount {
		sec := r.Int64()
		if got := time.Unix(sec, 0).Unix(); got != sec {
			t.Errorf("Unix(%d, 0).Unix() = %d", sec, got)
			return
		}
	}
}

func TestUnixNanoUTCAndBack(t *testing.T) {
	pcg := rand.NewPCG(3, 4)
	r := rand.New(&pcg)
	// Try small dates first, then the large ones. The span is only a few
	// hundred years for nanoseconds in an int64.
	for range randCount {
		nsec := int64(r.Int32())
		tm := time.Unix(0, nsec)
		if got := nanosOf(tm); got != nsec {
			t.Errorf("Unix(0, %d) nanoseconds = %d", nsec, got)
			return
		}
	}
	for range randCount {
		nsec := r.Int64()
		tm := time.Unix(0, nsec)
		if got := nanosOf(tm); got != nsec {
			t.Errorf("Unix(0, %d) nanoseconds = %d", nsec, got)
			return
		}
	}
}

func TestUnixMilli(t *testing.T) {
	pcg := rand.NewPCG(5, 6)
	r := rand.New(&pcg)
	for range randCount {
		msec := r.Int64()
		if got := time.UnixMilli(msec).UnixMilli(); got != msec {
			t.Errorf("UnixMilli(%d).UnixMilli() = %d", msec, got)
			return
		}
	}
}

func TestUnixMicro(t *testing.T) {
	pcg := rand.NewPCG(7, 8)
	r := rand.New(&pcg)
	for range randCount {
		usec := r.Int64()
		if got := time.UnixMicro(usec).UnixMicro(); got != usec {
			t.Errorf("UnixMicro(%d).UnixMicro() = %d", usec, got)
			return
		}
	}
}

func TestUnixNano(t *testing.T) {
	pcg := rand.NewPCG(9, 10)
	r := rand.New(&pcg)
	for range randCount {
		// Keep the value inside the range UnixNano can hold.
		nsec := r.Int64()
		tm := time.Unix(0, nsec)
		if got := tm.UnixNano(); got != nsec {
			t.Errorf("Unix(0, %d).UnixNano() = %d", nsec, got)
			return
		}
	}
}
