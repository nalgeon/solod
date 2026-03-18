// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import "testing"

func TestAbsYdaySplit(t *testing.T) {
	ends := []int{31, 30, 31, 30, 31, 31, 30, 31, 30, 31, 31, 29}
	bad := 0
	wantMonth := absMonth(3)
	wantDay := 1
	for yday := range absYday(366) {
		month, day := absYday_split(yday)
		if month != wantMonth || day != wantDay {
			t.Errorf("absYday(%d).split() = %d, %d, want %d, %d", yday, month, day, wantMonth, wantDay)
			if bad++; bad >= 20 {
				t.Fatalf("too many errors")
			}
		}
		if wantDay++; wantDay > ends[wantMonth-3] {
			wantMonth++
			wantDay = 1
		}
	}
}

func TestAbsDate(t *testing.T) {
	ends := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	isLeap := func(year int) bool {
		y := uint64(year) + absoluteYears
		return y%4 == 0 && (y%100 != 0 || y%400 == 0)
	}
	wantYear := 0
	wantMonth := March
	wantMday := 1
	wantYday := 31 + 29 + 1
	bad := 0
	absoluteYears := int64(absoluteYears)
	for days := range absDays(1e6) {
		date := absDays_date(days)
		year, month, mday := date.Year, date.Month, date.Day
		year += int(absoluteYears)
		if year != wantYear || month != wantMonth || mday != wantMday {
			t.Errorf("days(%d).date() = %v, %v, %v, want %v, %v, %v", days,
				year, month, mday,
				wantYear, wantMonth, wantMday)
			if bad++; bad >= 20 {
				t.Fatalf("too many errors")
			}
		}

		year, yday := absDays_yearYday(days)
		year += int(absoluteYears)
		if year != wantYear || yday != wantYday {
			t.Errorf("days(%d).yearYday() = %v, %v, want %v, %v, ", days,
				year, yday,
				wantYear, wantYday)
			if bad++; bad >= 20 {
				t.Fatalf("too many errors")
			}
		}

		if wantMday++; wantMday == ends[wantMonth-1]+1 || wantMonth == February && wantMday == 29 && !isLeap(year) {
			wantMonth++
			wantMday = 1
		}
		wantYday++
		if wantMonth == December+1 {
			wantYear++
			wantMonth = January
			wantMday = 1
			wantYday = 1
		}
	}
}

func TestDateToAbsDays(t *testing.T) {
	isLeap := func(year int64) bool {
		return year%4 == 0 && (year%100 != 0 || year%400 == 0)
	}
	wantDays := absDays(marchThruDecember)
	bad := 0
	for year := int64(1); year < 10000; year++ {
		days := dateToAbsDays(year-absoluteYears, January, 1)
		if days != wantDays {
			t.Errorf("dateToAbsDays(abs %d, Jan, 1) = %d, want %d", year, days, wantDays)
			if bad++; bad >= 20 {
				t.Fatalf("too many errors")
			}
		}
		wantDays += 365
		if isLeap(year) {
			wantDays++
		}
	}
}
