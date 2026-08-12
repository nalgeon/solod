package slog

import "solod.dev/so/time"

// Kind represents the type of a Value.
type Kind int

const (
	KindAny Kind = iota // reserved for future use
	KindBool
	KindDuration
	KindFloat64
	KindInt64
	KindString
	KindTime
	KindUint64
)

// Value holds a value of a supported type.
type Value struct {
	kind Kind
	num  uint64 // int, uint, bool, duration, time (UnixNano)
	str  string // string
}

// -- Constructors.

// StringValue returns a new [Value] for a string.
func StringValue(v string) Value {
	return Value{kind: KindString, str: v}
}

// IntValue returns a [Value] for an int.
func IntValue(v int) Value {
	return Int64Value(int64(v))
}

// Int64Value returns a [Value] for an int64.
func Int64Value(v int64) Value {
	return Value{kind: KindInt64, num: uint64(v)}
}

// Uint64Value returns a [Value] for a uint64.
func Uint64Value(v uint64) Value {
	return Value{kind: KindUint64, num: v}
}

// Float64Value returns a [Value] for a floating-point number.
func Float64Value(v float64) Value {
	return Value{kind: KindFloat64, num: float64bits(v)}
}

// BoolValue returns a [Value] for a bool.
func BoolValue(v bool) Value {
	u := uint64(0)
	if v {
		u = 1
	}
	return Value{kind: KindBool, num: u}
}

// TimeValue returns a [Value] for a [time.Time].
// It discards the monotonic portion.
func TimeValue(v time.Time) Value {
	if v.IsZero() {
		// UnixNano on the zero time is undefined.
		panic("slog: TimeValue called with zero time")
	}
	nsec := v.UnixNano()
	t := time.Unix(0, nsec)
	if !v.Equal(t) {
		panic("slog: TimeValue cannot be represented as UnixNano")
	}
	// UnixNano correctly represents the time, so use a zero-alloc representation.
	return Value{kind: KindTime, num: uint64(nsec)}
}

// DurationValue returns a [Value] for a [time.Duration].
func DurationValue(v time.Duration) Value {
	return Value{kind: KindDuration, num: uint64(v.Nanoseconds())}
}

// -- Accessors.

// Kind returns the kind of the value.
func (v Value) Kind() Kind {
	return v.kind
}

// String returns the value as a string. Panics if v is not a string.
func (v Value) String() string {
	if v.kind != KindString {
		panic("slog: Value.String called on non-string")
	}
	return v.str
}

// Int returns the value as an int. Panics if v is not a signed integer.
func (v Value) Int() int {
	return int(v.Int64())
}

// Int64 returns the value as an int64. Panics if v is not a signed integer.
func (v Value) Int64() int64 {
	if v.kind != KindInt64 {
		panic("slog: Value.Int64 called on non-int64")
	}
	return int64(v.num)
}

// Uint64 returns the value as a uint64. Panics if v is not an unsigned integer.
func (v Value) Uint64() uint64 {
	if v.kind != KindUint64 {
		panic("slog: Value.Uint64 called on non-uint64")
	}
	return v.num
}

// Float64 returns the value as a float64. Panics if v is not a float64.
func (v Value) Float64() float64 {
	if v.kind != KindFloat64 {
		panic("slog: Value.Float64 called on non-float64")
	}
	return v.float()
}

func (v Value) float() float64 {
	return float64frombits(v.num)
}

// Bool returns the value as a bool. Panics if v is not a bool.
func (v Value) Bool() bool {
	if v.kind != KindBool {
		panic("slog: Value.Bool called on non-bool")
	}
	return v.bool()
}

func (v Value) bool() bool {
	return v.num == 1
}

// Time returns the value as a time.Time. Panics if v is not a valid UnixNano time.
func (v Value) Time() time.Time {
	if v.kind != KindTime {
		panic("slog: Value.Time called on non-time")
	}
	return v.time()
}

func (v Value) time() time.Time {
	return time.Unix(0, int64(v.num))
}

// Duration returns the value as a time.Duration. Panics if v is not a duration.
func (v Value) Duration() time.Duration {
	if v.kind != KindDuration {
		panic("slog: Value.Duration called on non-duration")
	}
	return v.duration()
}

func (v Value) duration() time.Duration {
	return time.Duration(int64(v.num))
}

// Attr is a key-value pair.
type Attr struct {
	Key   string
	Value Value
}

// String returns an Attr for a string value.
func String(key, value string) Attr {
	return Attr{Key: key, Value: StringValue(value)}
}

// Int returns an Attr for an int value.
func Int(key string, value int) Attr {
	return Attr{Key: key, Value: IntValue(value)}
}

// Int64 returns an Attr for an int64 value.
func Int64(key string, value int64) Attr {
	return Attr{Key: key, Value: Int64Value(value)}
}

// Uint64 returns an Attr for a uint64 value.
func Uint64(key string, value uint64) Attr {
	return Attr{Key: key, Value: Uint64Value(value)}
}

// Float64 returns an Attr for a float64 value.
func Float64(key string, value float64) Attr {
	return Attr{Key: key, Value: Float64Value(value)}
}

// Bool returns an Attr for a bool value.
func Bool(key string, value bool) Attr {
	return Attr{Key: key, Value: BoolValue(value)}
}

// Time returns an Attr for a time.Time value.
func Time(key string, value time.Time) Attr {
	return Attr{Key: key, Value: TimeValue(value)}
}

// Duration returns an Attr for a time.Duration value.
func Duration(key string, value time.Duration) Attr {
	return Attr{Key: key, Value: DurationValue(value)}
}
