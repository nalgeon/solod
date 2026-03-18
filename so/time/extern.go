package time

import _ "embed"

//so:embed time.h
var time_h string

//so:embed time.c
var time_c string

// wall returns the current wall clock time.
//
//so:extern
func time_wall() (int64, int32) { return 0, 0 }

// mono returns the current monotonic time in nanoseconds.
//
//so:extern
func time_mono() int64 { return 0 }

// Monotonic times are reported as offsets from monoStart.
// We initialize monoStart to time_mono() - 1 so that on systems where
// monotonic time resolution is fairly low (e.g. Windows 2008
// which appears to have a default resolution of 15ms),
// we avoid ever reporting a monotonic time of 0.
// (Callers may want to use 0 as "time not set".)
//
//so:extern
var time_monoStart int64 = time_mono() - 1
