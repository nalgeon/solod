// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	stdtime "time"
)

func FuzzUnix(f *testing.F) {
	// Compare the Unix constructors and the Unix methods with the time package.
	f.Add(int64(0), int64(0))
	f.Add(int64(1), int64(0))
	f.Add(int64(-1), int64(0))
	f.Add(int64(1221681866), int64(0))
	f.Add(int64(-1221681866), int64(0))
	f.Add(int64(1221681866), int64(-1000000000))
	f.Add(int64(1221681866), int64(1000000000))
	f.Add(int64(1221681866), int64(999999999))
	f.Add(int64(1221681866), int64(-999999999))
	f.Add(int64(-62135596800), int64(0)) // the zero Time
	f.Add(int64(1e18), int64(1e18))
	f.Add(int64(-1e18), int64(-1e18))

	f.Fuzz(func(t *testing.T, sec, nsec int64) {
		checkTime(t, "Unix", Unix(sec, nsec), stdtime.Unix(sec, nsec).UTC())
		checkTime(t, "UnixMilli", UnixMilli(sec), stdtime.UnixMilli(sec).UTC())
		checkTime(t, "UnixMicro", UnixMicro(sec), stdtime.UnixMicro(sec).UTC())
	})
}
