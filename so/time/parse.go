package time

import "solod.dev/so/errors"

// ErrParse is returned by Parse when the input cannot be parsed.
var ErrParse = errors.New("time: cannot parse")

// Parse parses value per layout (strptime verbs) and returns the Time.
// offset specifies what timezone the input value is in.
func Parse(layout string, value string, offset Offset) (Time, error) {
	// Fast paths for known layouts - avoid strptime
	// overhead and verb issues (%z, %f).
	if layout == RFC3339 {
		return parseRFC3339(value, offset)
	}
	if layout == RFC3339Nano {
		return parseRFC3339Nano(value, offset)
	}
	if layout == DateTime {
		return parseDateTime(value, offset)
	}
	if layout == DateOnly {
		return parseDateOnly(value, offset)
	}
	if layout == TimeOnly {
		return parseTimeOnly(value, offset)
	}

	// General case: strptime.
	var tm time_tm
	end := strptime(value, layout, &tm)
	if end == nil {
		return Time{}, ErrParse
	}
	return Date(int(tm.tm_year)+1900, Month(int(tm.tm_mon)+1), int(tm.tm_mday),
		int(tm.tm_hour), int(tm.tm_min), int(tm.tm_sec), 0, offset), nil
}

// parseRFC3339 parses "YYYY-MM-DDTHH:MM:SSZ" or "YYYY-MM-DDTHH:MM:SS+HH:MM".
func parseRFC3339(value string, offset Offset) (Time, error) {
	// YYYY-MM-DDTHH:MM:SSZ      (20)
	// YYYY-MM-DDTHH:MM:SS+HH:MM (25)
	// 0123456789012345678901234
	n := len(value)
	ok := (n == 20 || n == 25) &&
		value[4] == '-' && value[7] == '-' && value[10] == 'T' &&
		value[13] == ':' && value[16] == ':'
	if !ok {
		return Time{}, ErrParse
	}
	year := parse4(value, 0)
	month := parse2(value, 5)
	day := parse2(value, 8)
	hour := parse2(value, 11)
	min := parse2(value, 14)
	sec := parse2(value, 17)
	if (year | month | day | hour | min | sec) < 0 {
		return Time{}, ErrParse
	}
	off, ok := parseOffset(value, 19)
	if !ok {
		return Time{}, ErrParse
	}
	return Date(year, Month(month), day, hour, min, sec, 0, offset+off), nil
}

// parseRFC3339Nano parses "YYYY-MM-DDTHH:MM:SS.nnnnnnnnnZ" or with +HH:MM/-HH:MM.
func parseRFC3339Nano(value string, offset Offset) (Time, error) {
	// YYYY-MM-DDTHH:MM:SS.nnnnnnnnnZ      (30)
	// YYYY-MM-DDTHH:MM:SS.nnnnnnnnn+HH:MM (35)
	// 01234567890123456789012345678901234
	n := len(value)
	ok := (n == 30 || n == 35) &&
		value[4] == '-' && value[7] == '-' && value[10] == 'T' &&
		value[13] == ':' && value[16] == ':' && value[19] == '.'
	if !ok {
		return Time{}, ErrParse
	}
	year := parse4(value, 0)
	month := parse2(value, 5)
	day := parse2(value, 8)
	hour := parse2(value, 11)
	min := parse2(value, 14)
	sec := parse2(value, 17)
	nsec := parse9(value, 20)
	if (year | month | day | hour | min | sec | nsec) < 0 {
		return Time{}, ErrParse
	}
	off, ok := parseOffset(value, 29)
	if !ok {
		return Time{}, ErrParse
	}
	return Date(year, Month(month), day, hour, min, sec, nsec, offset+off), nil
}

// parseDateTime parses "YYYY-MM-DD HH:MM:SS".
func parseDateTime(value string, offset Offset) (Time, error) {
	// YYYY-MM-DD HH:MM:SS
	// 0123456789012345678
	if len(value) != 19 ||
		value[4] != '-' || value[7] != '-' || value[10] != ' ' ||
		value[13] != ':' || value[16] != ':' {
		return Time{}, ErrParse
	}
	year := parse4(value, 0)
	month := parse2(value, 5)
	day := parse2(value, 8)
	hour := parse2(value, 11)
	min := parse2(value, 14)
	sec := parse2(value, 17)
	if (year | month | day | hour | min | sec) < 0 {
		return Time{}, ErrParse
	}
	return Date(year, Month(month), day, hour, min, sec, 0, offset), nil
}

// parseDateOnly parses "YYYY-MM-DD".
func parseDateOnly(value string, offset Offset) (Time, error) {
	// YYYY-MM-DD
	// 0123456789
	if len(value) != 10 || value[4] != '-' || value[7] != '-' {
		return Time{}, ErrParse
	}
	year := parse4(value, 0)
	month := parse2(value, 5)
	day := parse2(value, 8)
	if (year | month | day) < 0 {
		return Time{}, ErrParse
	}
	return Date(year, Month(month), day, 0, 0, 0, 0, offset), nil
}

// parseTimeOnly parses "HH:MM:SS".
// Returns a time with date January 1, year 0.
func parseTimeOnly(value string, offset Offset) (Time, error) {
	// HH:MM:SS
	// 01234567
	if len(value) != 8 || value[2] != ':' || value[5] != ':' {
		return Time{}, ErrParse
	}
	hour := parse2(value, 0)
	min := parse2(value, 3)
	sec := parse2(value, 6)
	if (hour | min | sec) < 0 {
		return Time{}, ErrParse
	}
	return Date(0, Month(1), 1, hour, min, sec, 0, offset), nil
}

// parseOffset parses "Z", "+HH:MM", or "-HH:MM" at position i. The offset must
// end the value. Returns the offset in seconds.
func parseOffset(value string, i int) (Offset, bool) {
	rest := len(value) - i
	if rest == 1 && value[i] == 'Z' {
		return UTC, true
	}
	if rest != 6 || (value[i] != '+' && value[i] != '-') {
		return 0, false
	}
	if value[i+3] != ':' {
		return 0, false
	}
	h := parse2(value, i+1)
	m := parse2(value, i+4)
	if (h | m) < 0 {
		return 0, false
	}
	off := Offset(h*3600 + m*60)
	if value[i] == '-' {
		off = -off
	}
	return off, true
}

// parse2 parses a 2-digit decimal from s at position i. Returns -1 if invalid.
func parse2(s string, i int) int {
	d1 := int(s[i] - '0')
	d2 := int(s[i+1] - '0')
	if (d1 | d2) > 9 {
		return -1
	}
	return d1*10 + d2
}

// parse4 parses a 4-digit decimal from s at position i. Returns -1 if invalid.
func parse4(s string, i int) int {
	d1 := int(s[i] - '0')
	d2 := int(s[i+1] - '0')
	d3 := int(s[i+2] - '0')
	d4 := int(s[i+3] - '0')
	if (d1 | d2 | d3 | d4) > 9 {
		return -1
	}
	return d1*1000 + d2*100 + d3*10 + d4
}

// parse9 parses a 9-digit decimal from s at position i. Returns -1 if invalid.
func parse9(s string, i int) int {
	n := 0
	for j := range 9 {
		d := int(s[i+j] - '0')
		if d > 9 {
			return -1
		}
		n = n*10 + d
	}
	return n
}
