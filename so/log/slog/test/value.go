package slog_test

import (
	"solod.dev/so/log/slog"
	"solod.dev/so/math"
	"solod.dev/so/testing"
	"solod.dev/so/time"
)

// intVals holds the int64 values of the tests.
var intVals = [...]int64{
	0, 1, -1, 42, -42, 255, 256, 65535, 65536,
	math.MaxInt32, math.MinInt32, math.MaxInt64, math.MinInt64,
}

// uintVals holds the uint64 values of the tests.
var uintVals = [...]uint64{
	0, 1, 42, 255, 256, 65535, 65536,
	math.MaxInt32, math.MaxUint32, math.MaxInt64, math.MaxUint64,
}

// durVals holds the durations of the tests, in nanoseconds.
var durVals = [...]int64{
	0, 1, -1, 1000, 1000000, 1000000000, -1500000000, 5400000000000,
	math.MaxInt64, math.MinInt64,
}

// secVals holds the instants of the tests, in Unix seconds. UnixNano must hold
// every instant, so the range of the list stays inside the years 1678 to 2262.
var secVals = [...]int64{
	0, 1, -1, 1000000000, 1700000000, -1700000000,
	-9000000000, 9000000000,
}

func TestStringValue(t *testing.T) {
	for i, s := range strVals {
		v := slog.StringValue(s)
		if v.Kind() != slog.KindString {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindString))
			continue
		}
		if v.String() != s {
			t.Errorf("#%d: String() = %s, want %s", i, v.String(), s)
		}
	}
}

func TestInt64Value(t *testing.T) {
	for i, n := range intVals {
		v := slog.Int64Value(n)
		if v.Kind() != slog.KindInt64 {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindInt64))
			continue
		}
		if v.Int64() != n {
			t.Errorf("#%d: Int64() = %d, want %d", i, v.Int64(), n)
		}
	}
}

func TestIntValue(t *testing.T) {
	vals := [...]int{0, 1, -1, 42, -42, 65536, math.MaxInt, math.MinInt}
	for i, n := range vals {
		v := slog.IntValue(n)
		if v.Kind() != slog.KindInt64 {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindInt64))
			continue
		}
		if v.Int() != n {
			t.Errorf("#%d: Int() = %d, want %d", i, v.Int(), n)
		}
		if v.Int64() != int64(n) {
			t.Errorf("#%d: Int64() = %d, want %d", i, v.Int64(), int64(n))
		}
	}
}

func TestUint64Value(t *testing.T) {
	for i, n := range uintVals {
		v := slog.Uint64Value(n)
		if v.Kind() != slog.KindUint64 {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindUint64))
			continue
		}
		if v.Uint64() != n {
			t.Errorf("#%d: Uint64() = %x, want %x", i, v.Uint64(), n)
		}
	}
}

func TestFloat64Value(t *testing.T) {
	vals := [...]float64{
		0, 1, -1, 0.5, -0.5, 3.14, 1e-300, 1e300,
		math.MaxFloat64, math.SmallestNonzeroFloat64,
		math.Copysign(0, -1), math.Inf(1), math.Inf(-1),
	}
	for i, f := range vals {
		v := slog.Float64Value(f)
		if v.Kind() != slog.KindFloat64 {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindFloat64))
			continue
		}
		// The bits must survive, so a negative zero must not become a zero.
		if math.Float64bits(v.Float64()) != math.Float64bits(f) {
			t.Errorf("#%d: Float64() = %g, want %g", i, v.Float64(), f)
		}
	}
}

func TestFloat64ValueNaN(t *testing.T) {
	v := slog.Float64Value(math.NaN())
	if v.Kind() != slog.KindFloat64 {
		t.Errorf("Kind() = %d, want %d", int(v.Kind()), int(slog.KindFloat64))
		return
	}
	if !math.IsNaN(v.Float64()) {
		t.Errorf("Float64() = %g, want NaN", v.Float64())
	}
}

func TestBoolValue(t *testing.T) {
	vals := [...]bool{true, false}
	for i, b := range vals {
		v := slog.BoolValue(b)
		if v.Kind() != slog.KindBool {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindBool))
			continue
		}
		if v.Bool() != b {
			t.Errorf("#%d: Bool() = %t, want %t", i, v.Bool(), b)
		}
	}
}

func TestTimeValue(t *testing.T) {
	nsecs := [...]int64{0, 1, 999999999, 500000000}
	for i, sec := range secVals {
		for j, nsec := range nsecs {
			want := time.Unix(sec, nsec)
			v := slog.TimeValue(want)
			if v.Kind() != slog.KindTime {
				t.Errorf("#%d.%d: Kind() = %d, want %d", i, j, int(v.Kind()), int(slog.KindTime))
				continue
			}
			got := v.Time()
			if !got.Equal(want) {
				t.Errorf("#%d.%d: Time() = %d.%09d, want %d.%09d",
					i, j, got.Unix(), got.Nanosecond(), want.Unix(), want.Nanosecond())
			}
			if got.UnixNano() != want.UnixNano() {
				t.Errorf("#%d.%d: UnixNano() = %d, want %d", i, j, got.UnixNano(), want.UnixNano())
			}
		}
	}
}

func TestDurationValue(t *testing.T) {
	for i, n := range durVals {
		d := time.Duration(n)
		v := slog.DurationValue(d)
		if v.Kind() != slog.KindDuration {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(v.Kind()), int(slog.KindDuration))
			continue
		}
		if v.Duration() != d {
			t.Errorf("#%d: Duration() = %d, want %d", i, int64(v.Duration()), n)
		}
	}
}

// TestKinds checks that every kind has a value of its own.
func TestKinds(t *testing.T) {
	kinds := [...]slog.Kind{
		slog.KindAny, slog.KindBool, slog.KindDuration, slog.KindFloat64,
		slog.KindInt64, slog.KindString, slog.KindTime, slog.KindUint64,
	}
	for i := range len(kinds) {
		for j := range len(kinds) {
			if (kinds[i] == kinds[j]) != (i == j) {
				t.Errorf("kind #%d and kind #%d are both %d", i, j, int(kinds[i]))
			}
		}
	}
}

func TestAttrString(t *testing.T) {
	for i, s := range strVals {
		a := slog.String("key", s)
		if a.Key != "key" {
			t.Errorf("#%d: Key = %s, want key", i, a.Key)
			continue
		}
		if a.Value.Kind() != slog.KindString {
			t.Errorf("#%d: Kind() = %d, want %d", i, int(a.Value.Kind()), int(slog.KindString))
			continue
		}
		if a.Value.String() != s {
			t.Errorf("#%d: Value = %s, want %s", i, a.Value.String(), s)
		}
	}
}

func TestAttrInt(t *testing.T) {
	vals := [...]int{0, 1, -1, 42, math.MaxInt, math.MinInt}
	for i, n := range vals {
		a := slog.Int("n", n)
		if a.Key != "n" {
			t.Errorf("#%d: Key = %s, want n", i, a.Key)
			continue
		}
		if a.Value.Int() != n {
			t.Errorf("#%d: Value = %d, want %d", i, a.Value.Int(), n)
		}
	}
}

func TestAttrInt64(t *testing.T) {
	for i, n := range intVals {
		a := slog.Int64("n", n)
		if a.Key != "n" {
			t.Errorf("#%d: Key = %s, want n", i, a.Key)
			continue
		}
		if a.Value.Int64() != n {
			t.Errorf("#%d: Value = %d, want %d", i, a.Value.Int64(), n)
		}
	}
}

func TestAttrUint64(t *testing.T) {
	for i, n := range uintVals {
		a := slog.Uint64("n", n)
		if a.Key != "n" {
			t.Errorf("#%d: Key = %s, want n", i, a.Key)
			continue
		}
		if a.Value.Uint64() != n {
			t.Errorf("#%d: Value = %x, want %x", i, a.Value.Uint64(), n)
		}
	}
}

func TestAttrFloat64(t *testing.T) {
	vals := [...]float64{0, 1, -1, 9.5, -0.25, 1e300}
	for i, f := range vals {
		a := slog.Float64("f", f)
		if a.Key != "f" {
			t.Errorf("#%d: Key = %s, want f", i, a.Key)
			continue
		}
		if a.Value.Float64() != f {
			t.Errorf("#%d: Value = %g, want %g", i, a.Value.Float64(), f)
		}
	}
}

func TestAttrBool(t *testing.T) {
	vals := [...]bool{true, false}
	for i, b := range vals {
		a := slog.Bool("ok", b)
		if a.Key != "ok" {
			t.Errorf("#%d: Key = %s, want ok", i, a.Key)
			continue
		}
		if a.Value.Bool() != b {
			t.Errorf("#%d: Value = %t, want %t", i, a.Value.Bool(), b)
		}
	}
}

func TestAttrTime(t *testing.T) {
	for i, sec := range secVals {
		want := time.Unix(sec, 0)
		a := slog.Time("at", want)
		if a.Key != "at" {
			t.Errorf("#%d: Key = %s, want at", i, a.Key)
			continue
		}
		if !a.Value.Time().Equal(want) {
			t.Errorf("#%d: Value = %d, want %d", i, a.Value.Time().Unix(), sec)
		}
	}
}

func TestAttrDuration(t *testing.T) {
	for i, n := range durVals {
		d := time.Duration(n)
		a := slog.Duration("took", d)
		if a.Key != "took" {
			t.Errorf("#%d: Key = %s, want took", i, a.Key)
			continue
		}
		if a.Value.Duration() != d {
			t.Errorf("#%d: Value = %d, want %d", i, int64(a.Value.Duration()), n)
		}
	}
}
