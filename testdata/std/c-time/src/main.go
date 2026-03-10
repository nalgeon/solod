package main

import (
	"github.com/nalgeon/solod/so/c/stdio"
	"github.com/nalgeon/solod/so/c/time"
)

func main() {
	{
		// Current time.
		var now time.TimeT
		now = time.Time(&now)
		if now <= 0 {
			panic("want now > 0")
		}
	}
	{
		// Clock.
		ticks := time.Clock()
		_ = ticks
	}
	{
		// ClocksPerSec.
		cps := time.ClocksPerSec
		if cps <= 0 {
			panic("want ClocksPerSec > 0")
		}
	}
	{
		// Difftime.
		var t1 time.TimeT
		t1 = time.Time(&t1)
		var t2 time.TimeT
		t2 = time.Time(&t2)
		diff := time.Difftime(t2, t1)
		if diff < 0.0 {
			panic("want diff >= 0")
		}
	}
	{
		// Gmtime.
		var ts time.TimeT
		ts = 0
		tm := time.Gmtime(&ts)
		// Unix epoch: 1970-01-01 00:00:00 UTC.
		if tm.Year != 70 {
			panic("want Year == 70")
		}
		if tm.Mon != 0 {
			panic("want Mon == 0")
		}
		if tm.Mday != 1 {
			panic("want Mday == 1")
		}
	}
	{
		// Mktime.
		tm := time.Tm{
			Sec:   0,
			Min:   0,
			Hour:  0,
			Mday:  1,
			Mon:   0,
			Year:  70,
			Isdst: -1,
		}
		ts := time.Mktime(&tm)
		// Should normalize and return a valid timestamp.
		_ = ts
	}
	{
		// Strftime.
		var buf [64]byte
		var ts time.TimeT
		ts = 0
		tm := time.Gmtime(&ts)
		n := time.Strftime(&buf[0], 64, "%Y-%m-%d", &tm)
		if n == 0 {
			panic("strftime failed")
		}
		stdio.Printf("%s\n", &buf[0])
	}
}
