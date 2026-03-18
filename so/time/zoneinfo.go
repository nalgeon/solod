// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

// UTC represents Universal Coordinated Time (UTC).
var UTC = &Location{name: "UTC", offset: 0, isDST: false}

// String returns a descriptive name for the time zone information,
// corresponding to the name argument to [LoadLocation] or [FixedZone].
func (l *Location) String() string {
	return l.get().name
}

func (l *Location) get() *Location {
	if l == nil {
		return UTC
	}
	return l
}

// FixedZone returns a [Location] that always uses
// the given zone name and offset (seconds east of UTC).
func FixedZone(name string, offset int) Location {
	return Location{
		name:   name,
		offset: offset,
		isDST:  false,
	}
}
