package slog

import (
	stdstrconv "strconv"
	stdstrings "strings"
	"testing"
	stdtime "time"

	"github.com/nalgeon/be"
	"solod.dev/so/math"
	"solod.dev/so/strings"
	"solod.dev/so/time"
)

// The accessors of Value. Each one accepts a single kind and panics on any
// other kind.
const (
	accString = iota
	accInt
	accInt64
	accUint64
	accFloat64
	accBool
	accTime
	accDuration
	nAccessors
)

// accNames holds the name of each accessor.
var accNames = [nAccessors]string{
	"String", "Int", "Int64", "Uint64", "Float64", "Bool", "Time", "Duration",
}

// accKinds holds the kind each accessor accepts.
var accKinds = [nAccessors]Kind{
	KindString, KindInt64, KindInt64, KindUint64,
	KindFloat64, KindBool, KindTime, KindDuration,
}

// allKinds holds every kind.
var allKinds = []Kind{
	KindAny, KindBool, KindDuration, KindFloat64,
	KindInt64, KindString, KindTime, KindUint64,
}

// valueOfKind returns a value of the kind.
func valueOfKind(k Kind) Value {
	switch k {
	case KindBool:
		return BoolValue(true)
	case KindDuration:
		return DurationValue(time.Second)
	case KindFloat64:
		return Float64Value(1.5)
	case KindInt64:
		return Int64Value(42)
	case KindString:
		return StringValue("s")
	case KindTime:
		return TimeValue(time.Unix(1, 0))
	case KindUint64:
		return Uint64Value(7)
	}
	return Value{}
}

// callAccessor calls the accessor on v.
func callAccessor(v Value, acc int) {
	switch acc {
	case accString:
		_ = v.String()
	case accInt:
		_ = v.Int()
	case accInt64:
		_ = v.Int64()
	case accUint64:
		_ = v.Uint64()
	case accFloat64:
		_ = v.Float64()
	case accBool:
		_ = v.Bool()
	case accTime:
		_ = v.Time()
	case accDuration:
		_ = v.Duration()
	}
}

func TestValueAccessorPanic(t *testing.T) {
	for _, kind := range allKinds {
		for acc := range nAccessors {
			checkAccessor(t, kind, acc)
		}
	}
}

// checkAccessor calls the accessor on a value of the kind and checks the
// panic.
func checkAccessor(t *testing.T, kind Kind, acc int) {
	t.Helper()
	want := accKinds[acc] != kind
	defer func() {
		got := recover() != nil
		if got != want {
			t.Errorf("kind %d, %s(): panic = %t, want %t", int(kind), accNames[acc], got, want)
		}
	}()
	callAccessor(valueOfKind(kind), acc)
}

func TestTimeValuePanic(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		defer func() {
			be.True(t, recover() != nil)
		}()
		TimeValue(time.Time{})
	})

	t.Run("overflow", func(t *testing.T) {
		defer func() {
			be.True(t, recover() != nil)
		}()
		TimeValue(time.Date(3000, time.January, 1, 0, 0, 0, 0, time.UTC))
	})
}

func TestNewPanic(t *testing.T) {
	defer func() {
		be.True(t, recover() != nil)
	}()
	New(nil)
}

func TestFloatBits(t *testing.T) {
	vals := []float64{
		0, 1, -1, 0.5, 3.14, 1e300, 1e-300,
		math.MaxFloat64, math.SmallestNonzeroFloat64,
		math.Copysign(0, -1), math.Inf(1), math.Inf(-1), math.NaN(),
	}
	for _, f := range vals {
		bits := math.Float64bits(f)
		be.Equal(t, float64bits(f), bits)
		// NaN is not equal to itself, so compare the bits.
		be.Equal(t, math.Float64bits(float64frombits(bits)), bits)
	}
	// Every NaN payload must survive the round trip.
	for i := range uint64(64) {
		bits := math.Float64bits(math.NaN()) | i
		be.Equal(t, float64bits(float64frombits(bits)), bits)
	}
}

func TestDefaultLazyInit(t *testing.T) {
	saved := defaultLogger
	defer func() { defaultLogger = saved }()

	defaultOnce.Init()
	defaultLogger = nil
	d := Default()
	be.True(t, d != nil)
	be.True(t, d.Enabled(LevelInfo))
	be.True(t, !d.Enabled(LevelDebug))
	be.True(t, Default() == d)
}

func TestSetDefaultConsumesOnce(t *testing.T) {
	// Check that SetDefault stops the lazy init, so a later
	// top-level call keeps the logger of the caller.
	saved := defaultLogger
	defer func() { defaultLogger = saved }()

	defaultOnce.Init()
	defaultLogger = nil

	var sb strings.Builder
	defer sb.Free()
	h := NewTextHandler(&sb, LevelInfo)
	l := New(&h)
	SetDefault(&l)

	Info("hello")
	be.True(t, Default() == &l)
	be.True(t, strings.Contains(sb.String(), "INFO hello\n"))
}

// -- Fuzzing.

// The kinds the fuzzer builds an attr of.
const nFuzzKinds = 7

// maxFuzzAttrs is the number of the attrs the fuzzer builds.
const maxFuzzAttrs = 8

// The bounds of the instant of a fuzzed record. UnixNano holds every instant
// inside the bounds, and Go and So write the same year for it.
const (
	minFuzzSec = -8000000000
	maxFuzzSec = 8000000000
)

// cursor reads the values of a fuzzer from a byte string.
type cursor struct {
	b []byte
}

// empty reports whether the cursor has no byte left.
func (c *cursor) empty() bool { return len(c.b) == 0 }

// byte reads one byte. It gives 0 at the end of the input.
func (c *cursor) byte() byte {
	if len(c.b) == 0 {
		return 0
	}
	v := c.b[0]
	c.b = c.b[1:]
	return v
}

// uint64 reads eight bytes, big endian.
func (c *cursor) uint64() uint64 {
	var v uint64
	for range 8 {
		v = v<<8 | uint64(c.byte())
	}
	return v
}

// str reads a length-prefixed string of up to 255 bytes.
func (c *cursor) str() string {
	n := min(int(c.byte()), len(c.b))
	s := string(c.b[:n])
	c.b = c.b[n:]
	return s
}

// fuzzAttr builds an attr from the cursor, and returns the attr and the text
// a handler must write for it.
func fuzzAttr(c *cursor) (Attr, string) {
	kind := int(c.byte()) % nFuzzKinds
	key := c.str()
	switch kind {
	case 0:
		v := c.str()
		return String(key, v), key + "=" + quoteRef(v)
	case 1:
		v := int64(c.uint64())
		return Int64(key, v), key + "=" + stdstrconv.FormatInt(v, 10)
	case 2:
		v := c.uint64()
		return Uint64(key, v), key + "=" + stdstrconv.FormatUint(v, 10)
	case 3:
		v := math.Float64frombits(c.uint64())
		return Float64(key, v), key + "=" + stdstrconv.FormatFloat(v, 'g', -1, 64)
	case 4:
		v := c.byte()%2 == 0
		return Bool(key, v), key + "=" + stdstrconv.FormatBool(v)
	case 5:
		sec := clampSec(int64(c.uint64()))
		v := time.Unix(sec, 0)
		return Time(key, v), key + "=" + stdtime.Unix(sec, 0).UTC().Format(stdtime.RFC3339)
	}
	v := time.Duration(int64(c.uint64()))
	return Duration(key, v), key + "=" + stdtime.Duration(v).String()
}

// clampSec brings a Unix second inside the bounds of a fuzzed instant.
func clampSec(sec int64) int64 {
	const span = maxFuzzSec - minFuzzSec + 1
	return minFuzzSec + (sec%span+span)%span
}

// quoteRef returns the text a handler must write for a string value.
func quoteRef(s string) string {
	quote := len(s) == 0
	for i := range len(s) {
		if s[i] == ' ' || s[i] == '"' || s[i] == '=' {
			quote = true
		}
	}
	if quote {
		return `"` + s + `"`
	}
	return s
}

// levelRef returns the name a handler must write for a level.
func levelRef(level Level) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	}
	return "UNKNOWN"
}

func FuzzHandle(f *testing.F) {
	f.Add([]byte("\x00\x03msg\x00\x01k\x05value"))
	f.Add([]byte("\x04\x2a\x01a\x01\x01b\x00\x00\x00\x00\x00\x00\x00\x2a"))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		c := &cursor{b: data}
		level := Level(int8(c.byte()))
		sec := clampSec(int64(c.uint64()))
		msg := c.str()

		var attrs []Attr
		var want stdstrings.Builder
		want.WriteString(stdtime.Unix(sec, 0).UTC().Format(stdtime.RFC3339) +
			" " + levelRef(level) + " " + msg)
		for !c.empty() && len(attrs) < maxFuzzAttrs {
			a, text := fuzzAttr(c)
			attrs = append(attrs, a)
			want.WriteString(" " + text)
		}
		want.WriteString("\n")

		var sb strings.Builder
		defer sb.Free()
		h := NewTextHandler(&sb, LevelDebug)
		r := Record{Time: time.Unix(sec, 0), Message: msg, Level: level, Attrs: attrs}
		if err := h.Handle(r); err != nil {
			t.Fatalf("Handle() = %v", err)
		}
		if got := sb.String(); got != want.String() {
			t.Errorf("Handle() wrote %q, want %q", got, want.String())
		}
	})
}
