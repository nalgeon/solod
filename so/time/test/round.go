// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"solod.dev/so/math/bits"
	"solod.dev/so/math/rand"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// unixToZero is the number of seconds from the zero time (January 1, year 1)
// to the Unix epoch. The package gives no way to read the absolute time,
// but the rounding tests need it to compute the expected result of a bizarre
// rounding like "to the nearest 3 ns".
//
// Compute it as t - year1 = (t - 1970) + (1970 - 2001) + (2001 - 1).
// t - 1970 is what Unix returns. 1970 - 2001 is -(31*365+8)*86400 seconds.
// 2001 - 1 is 2000*365.2425*86400 seconds.
const unixToZero = -978307200 + 63113904000

// roundCount is the number of random cases each rounding generator draws.
const roundCount = 20000

// remAbsNanos returns the remainder of the absolute nanosecond count of tm
// divided by d. The remainder is in the range [0, d), as Euclidean division
// defines it. d must be positive.
func remAbsNanos(tm time.Time, d time.Duration) int64 {
	sec := tm.Unix() + unixToZero
	nsec := uint64(tm.Nanosecond())

	// The absolute nanosecond count does not fit in an int64, so hold its
	// magnitude in the 128 bit pair (hi, lo) and the sign in neg.
	var hi, lo, extra uint64
	neg := sec < 0
	if neg {
		hi, lo = bits.Mul64(uint64(-sec), 1e9)
		// nsec is below 1e9 and the magnitude is at least 1e9, so the
		// subtraction does not go below zero.
		lo, extra = bits.Sub64(lo, nsec, 0)
		hi, _ = bits.Sub64(hi, 0, extra)
	} else {
		hi, lo = bits.Mul64(uint64(sec), 1e9)
		lo, extra = bits.Add64(lo, nsec, 0)
		hi, _ = bits.Add64(hi, 0, extra)
	}

	rem := int64(bits.Rem64(hi, lo, uint64(d)))
	if neg && rem != 0 {
		rem = int64(d) - rem
	}
	return rem
}

// roundOne checks Truncate and Round of the time Unix(sec, nsec) against the
// multiple d. It reports whether both results are correct.
func roundOne(t *testing.T, sec, nsec, di int64) bool {
	t0 := time.Unix(sec, nsec)
	d := time.Duration(di)
	if d < 0 {
		d = -d
	}
	if d <= 0 {
		d = 1
	}

	// To truncate, subtract the remainder.
	rem := remAbsNanos(t0, d)
	t1 := t0.Add(-time.Duration(rem))

	if trunc := t0.Truncate(d); !trunc.Equal(t1) {
		t.Errorf("Unix(%d, %d).Truncate(%d) = %d.%09d, want %d.%09d",
			sec, nsec, int64(d),
			trunc.Unix(), trunc.Nanosecond(), t1.Unix(), t1.Nanosecond())
		return false
	}

	// To round, add d back if the remainder is above half of d or is exactly
	// half of d. Rounding half to even instead would make the result depend
	// on the time zone, which is a bit strange.
	if rem > int64(d)/2 || rem+rem == int64(d) {
		t1 = t1.Add(d)
	}

	if rnd := t0.Round(d); !rnd.Equal(t1) {
		t.Errorf("Unix(%d, %d).Round(%d) = %d.%09d, want %d.%09d",
			sec, nsec, int64(d),
			rnd.Unix(), rnd.Nanosecond(), t1.Unix(), t1.Nanosecond())
		return false
	}
	return true
}

// A roundTimeCase is a calendar time and a multiple to round it to.
type roundTimeCase struct {
	year, month, day     int
	hour, minute, second int
	nsec                 int
	d                    int64
}

var roundTimeCases = []roundTimeCase{
	{-1, 1, 1, 12, 15, 30, 5e8, 3},
	{-1, 1, 1, 12, 15, 31, 5e8, 3},
	{2012, 1, 1, 12, 15, 30, 5e8, int64(time.Second)},
	{2012, 1, 1, 12, 15, 31, 5e8, int64(time.Second)},
}

func TestTruncateRound(t *testing.T) {
	for _, c := range roundTimeCases {
		tm := time.Date(c.year, time.Month(c.month), c.day, c.hour, c.minute, c.second, c.nsec, time.UTC)
		roundOne(t, tm.Unix(), int64(tm.Nanosecond()), c.d)
	}
	// 5.8*d rounds to 6*d, but .8*d + .8*d < 0 < d.
	roundOne(t, -19012425939, 649146258, 7435029458905025217)
	if t.Failed() {
		return
	}

	// Exhaustive near the zero time.
	for i := range 100 {
		for j := 1; j < 100; j++ {
			roundOne(t, unixToZero, int64(i), int64(j))
			roundOne(t, unixToZero, -int64(i), int64(j))
			if t.Failed() {
				return
			}
		}
	}
}

func TestTruncateRoundDivisors(t *testing.T) {
	// Check the multiples that divide a second.
	pcg := rand.NewPCG(11, 12)
	r := rand.New(&pcg)
	for range roundCount {
		v := r.Uint32()
		d := time.Duration(1)
		d <<= v % 9
		for range int(v>>16) % 9 {
			d *= 5
		}
		// Make room for the conversion between Unix and internal time. The
		// behavior too close to +-2^63 Unix seconds is full of wraparounds,
		// and a reasonable program never reaches it.
		sec := r.Int64() >> 1
		if !roundOne(t, sec, int64(r.Int32()), int64(d)) {
			return
		}
	}
}

func TestTruncateRoundMultiples(t *testing.T) {
	// Check the multiples of a second.
	pcg := rand.NewPCG(13, 14)
	r := rand.New(&pcg)
	for range roundCount {
		d := time.Duration(r.Int32()) * time.Second
		if d < 0 {
			d = -d
		}
		sec := r.Int64() >> 1
		if !roundOne(t, sec, int64(r.Int32()), int64(d)) {
			return
		}
	}
}

func TestTruncateRoundHalfway(t *testing.T) {
	// Check the times that lie exactly half of a multiple away from it.
	pcg := rand.NewPCG(15, 16)
	r := rand.New(&pcg)
	for range roundCount {
		di := r.Int64() & 0xfffffffe
		if di == 0 {
			di = 2
		}
		nsec := r.Int64()
		nsec -= nsec % di
		if nsec < 0 {
			nsec += di / 2
		} else {
			nsec -= di / 2
		}
		if !roundOne(t, 0, nsec, di) {
			return
		}
	}
}

func TestTruncateRoundAny(t *testing.T) {
	// Check arbitrary times and multiples.
	pcg := rand.NewPCG(17, 18)
	r := rand.New(&pcg)
	for range roundCount {
		sec := r.Int64() >> 1
		if !roundOne(t, sec, int64(r.Int32()), r.Int64()) {
			return
		}
	}
}
