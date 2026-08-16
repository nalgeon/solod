package cmp_test

import (
	"solod.dev/so/cmp"
	"solod.dev/so/math"
	"solod.dev/so/testing"
)

// ID is a named type over a supported type.
type ID int

// Name is a named type over a supported type.
type Name string

// Point is a type that FuncFor does not support.
type Point struct {
	x, y int
}

// Triple is a type that FuncFor does not support.
type Triple [3]int

// The sweep alphabet holds a NUL byte and a high byte. The comparison must
// treat both as ordinary bytes.
const sweepAlpha = "\x00ab\xff"

// maxWord is the length of the longest word of the string sweep. The sweep
// takes every pair of words, so the words are short.
const maxWord = 3

// hexDigits holds the digits of a hexadecimal number.
const hexDigits = "0123456789abcdef"

// dump writes s into buf as hexadecimal and returns the result.
func dump(buf []byte, s string) string {
	for i := 0; i < len(s); i++ {
		buf[2*i] = hexDigits[s[i]>>4]
		buf[2*i+1] = hexDigits[s[i]&0xf]
	}
	return string(buf[:2*len(s)])
}

// wordCount returns the number of words of sweepAlpha with the length n.
func wordCount(n int) int {
	count := 1
	for range n {
		count *= len(sweepAlpha)
	}
	return count
}

// wordTotal returns the number of words of sweepAlpha with a length up to maxWord.
func wordTotal() int {
	total := 0
	for n := 0; n <= maxWord; n++ {
		total += wordCount(n)
	}
	return total
}

// wordAt writes the word number i into buf and returns the result. The shorter
// words come first, and every word with a length up to maxWord appears once.
// The caller must keep i below wordTotal().
func wordAt(buf []byte, i int) string {
	for n := 0; n <= maxWord; n++ {
		count := wordCount(n)
		if i < count {
			for k := 0; k < n; k++ {
				buf[k] = sweepAlpha[i%len(sweepAlpha)]
				i /= len(sweepAlpha)
			}
			return string(buf[:n])
		}
		i -= count
	}
	return ""
}

// The reference implementations. Every sweep checks the package against the
// simplest code that gives the wanted result.

// compareInt8Brute compares a and b with the plain operators.
func compareInt8Brute(a, b int8) int {
	if a < b {
		return -1
	}
	if a > b {
		return +1
	}
	return 0
}

// compareUint8Brute compares a and b with the plain operators.
func compareUint8Brute(a, b uint8) int {
	if a < b {
		return -1
	}
	if a > b {
		return +1
	}
	return 0
}

// compareStringBrute compares a and b byte by byte.
func compareStringBrute(a, b string) int {
	n := min(len(b), len(a))
	for i := 0; i < n; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return +1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return +1
	}
	return 0
}

func TestCompare_Int(t *testing.T) {
	if got := cmp.Compare(11, 22); got != -1 {
		t.Errorf("Compare(11, 22) = %d, want -1", got)
	}
	if got := cmp.Compare(11, 11); got != 0 {
		t.Errorf("Compare(11, 11) = %d, want 0", got)
	}
	if got := cmp.Compare(22, 11); got != +1 {
		t.Errorf("Compare(22, 11) = %d, want +1", got)
	}
	if got := cmp.Compare(-22, 11); got != -1 {
		t.Errorf("Compare(-22, 11) = %d, want -1", got)
	}
}

func TestCompare_IntWidths(t *testing.T) {
	// Every width has its own comparison function, so every width needs a check.
	// A wrong function compares the raw bytes and loses both the sign and the byte order.
	if got := cmp.Compare(int8(-1), int8(1)); got != -1 {
		t.Errorf("Compare(int8(-1), int8(1)) = %d, want -1", got)
	}
	if got := cmp.Compare(int16(-1), int16(1)); got != -1 {
		t.Errorf("Compare(int16(-1), int16(1)) = %d, want -1", got)
	}
	if got := cmp.Compare(int32(-1), int32(1)); got != -1 {
		t.Errorf("Compare(int32(-1), int32(1)) = %d, want -1", got)
	}
	if got := cmp.Compare(int64(-1), int64(1)); got != -1 {
		t.Errorf("Compare(int64(-1), int64(1)) = %d, want -1", got)
	}
	if got := cmp.Compare(uint8(1), uint8(2)); got != -1 {
		t.Errorf("Compare(uint8(1), uint8(2)) = %d, want -1", got)
	}
	if got := cmp.Compare(uint16(1), uint16(256)); got != -1 {
		t.Errorf("Compare(uint16(1), uint16(256)) = %d, want -1", got)
	}
	if got := cmp.Compare(uint32(1), uint32(65536)); got != -1 {
		t.Errorf("Compare(uint32(1), uint32(65536)) = %d, want -1", got)
	}
	if got := cmp.Compare(uint64(1), uint64(1)<<32); got != -1 {
		t.Errorf("Compare(uint64(1), uint64(1)<<32) = %d, want -1", got)
	}
	if got := cmp.Compare(uint(1), uint(2)); got != -1 {
		t.Errorf("Compare(uint(1), uint(2)) = %d, want -1", got)
	}
	if got := cmp.Compare(byte(1), byte(2)); got != -1 {
		t.Errorf("Compare(byte(1), byte(2)) = %d, want -1", got)
	}
	if got := cmp.Compare(rune('a'), rune('b')); got != -1 {
		t.Errorf("Compare(rune('a'), rune('b')) = %d, want -1", got)
	}
}

func TestCompare_Float(t *testing.T) {
	if got := cmp.Compare(1.0, 1.1); got != -1 {
		t.Errorf("Compare(1.0, 1.1) = %d, want -1", got)
	}
	if got := cmp.Compare(1.1, 1.1); got != 0 {
		t.Errorf("Compare(1.1, 1.1) = %d, want 0", got)
	}
	if got := cmp.Compare(1.1, 1.0); got != +1 {
		t.Errorf("Compare(1.1, 1.0) = %d, want +1", got)
	}
	if got := cmp.Compare(float32(-1.0), float32(1.0)); got != -1 {
		t.Errorf("Compare(float32(-1.0), float32(1.0)) = %d, want -1", got)
	}
	if got := cmp.Compare(float32(1.5), float32(1.5)); got != 0 {
		t.Errorf("Compare(float32(1.5), float32(1.5)) = %d, want 0", got)
	}
}

func TestCompare_FloatNaN(t *testing.T) {
	nan := math.NaN()
	if got := cmp.Compare(nan, nan); got != 0 {
		t.Errorf("Compare(NaN, NaN) = %d, want 0", got)
	}
	if got := cmp.Compare(nan, 1.0); got != -1 {
		t.Errorf("Compare(NaN, 1.0) = %d, want -1", got)
	}
	if got := cmp.Compare(1.0, nan); got != +1 {
		t.Errorf("Compare(1.0, NaN) = %d, want +1", got)
	}
	if got := cmp.Compare(nan, math.Inf(-1)); got != -1 {
		t.Errorf("Compare(NaN, -Inf) = %d, want -1", got)
	}
	if got := cmp.Compare(math.Inf(-1), nan); got != +1 {
		t.Errorf("Compare(-Inf, NaN) = %d, want +1", got)
	}
	if got := cmp.Compare(nan, math.Inf(1)); got != -1 {
		t.Errorf("Compare(NaN, Inf) = %d, want -1", got)
	}
	nan32 := float32(nan)
	if got := cmp.Compare(nan32, nan32); got != 0 {
		t.Errorf("Compare(float32 NaN, float32 NaN) = %d, want 0", got)
	}
	if got := cmp.Compare(nan32, float32(1.0)); got != -1 {
		t.Errorf("Compare(float32 NaN, float32(1.0)) = %d, want -1", got)
	}
}

func TestCompare_FloatInf(t *testing.T) {
	inf := math.Inf(1)
	ninf := math.Inf(-1)
	if got := cmp.Compare(inf, inf); got != 0 {
		t.Errorf("Compare(Inf, Inf) = %d, want 0", got)
	}
	if got := cmp.Compare(ninf, ninf); got != 0 {
		t.Errorf("Compare(-Inf, -Inf) = %d, want 0", got)
	}
	if got := cmp.Compare(inf, 1.0); got != +1 {
		t.Errorf("Compare(Inf, 1.0) = %d, want +1", got)
	}
	if got := cmp.Compare(1.0, inf); got != -1 {
		t.Errorf("Compare(1.0, Inf) = %d, want -1", got)
	}
	if got := cmp.Compare(ninf, inf); got != -1 {
		t.Errorf("Compare(-Inf, Inf) = %d, want -1", got)
	}
	if got := cmp.Compare(ninf, 1.0); got != -1 {
		t.Errorf("Compare(-Inf, 1.0) = %d, want -1", got)
	}
	if got := cmp.Compare(1.0, ninf); got != +1 {
		t.Errorf("Compare(1.0, -Inf) = %d, want +1", got)
	}
	inf32 := float32(inf)
	if got := cmp.Compare(inf32, float32(1.0)); got != +1 {
		t.Errorf("Compare(float32 Inf, float32(1.0)) = %d, want +1", got)
	}
	if got := cmp.Compare(-inf32, inf32); got != -1 {
		t.Errorf("Compare(float32 -Inf, float32 Inf) = %d, want -1", got)
	}
}

func TestCompare_FloatZero(t *testing.T) {
	negzero := math.Copysign(0, -1)
	if got := cmp.Compare(negzero, 0.0); got != 0 {
		t.Errorf("Compare(-0.0, 0.0) = %d, want 0", got)
	}
	if got := cmp.Compare(0.0, negzero); got != 0 {
		t.Errorf("Compare(0.0, -0.0) = %d, want 0", got)
	}
	if got := cmp.Compare(negzero, negzero); got != 0 {
		t.Errorf("Compare(-0.0, -0.0) = %d, want 0", got)
	}
	if got := cmp.Compare(0.0, 0.0); got != 0 {
		t.Errorf("Compare(0.0, 0.0) = %d, want 0", got)
	}
	if got := cmp.Compare(negzero, 1.0); got != -1 {
		t.Errorf("Compare(-0.0, 1.0) = %d, want -1", got)
	}
	if got := cmp.Compare(negzero, -1.0); got != +1 {
		t.Errorf("Compare(-0.0, -1.0) = %d, want +1", got)
	}
	if cmp.Less(negzero, 0.0) {
		t.Error("Less(-0.0, 0.0) = true, want false")
	}
	if cmp.Less(0.0, negzero) {
		t.Error("Less(0.0, -0.0) = true, want false")
	}
	if !cmp.Equal(negzero, 0.0) {
		t.Error("Equal(-0.0, 0.0) = false, want true")
	}
	// The bit patterns of -0.0 and 0.0 differ, so Equal must not use memcmp.
	negzero32 := float32(negzero)
	if got := cmp.Compare(negzero32, float32(0.0)); got != 0 {
		t.Errorf("Compare(float32 -0.0, float32 0.0) = %d, want 0", got)
	}
	if !cmp.Equal(negzero32, float32(0.0)) {
		t.Error("Equal(float32 -0.0, float32 0.0) = false, want true")
	}
}

func TestCompare_String(t *testing.T) {
	if got := cmp.Compare("a", "aa"); got != -1 {
		t.Errorf("Compare(a, aa) = %d, want -1", got)
	}
	if got := cmp.Compare("a", "a"); got != 0 {
		t.Errorf("Compare(a, a) = %d, want 0", got)
	}
	if got := cmp.Compare("aa", "a"); got != +1 {
		t.Errorf("Compare(aa, a) = %d, want +1", got)
	}
	if got := cmp.Compare("", ""); got != 0 {
		t.Errorf("Compare(empty, empty) = %d, want 0", got)
	}
	if got := cmp.Compare("", "a"); got != -1 {
		t.Errorf("Compare(empty, a) = %d, want -1", got)
	}
	// memcmp returns the byte difference, so Compare must normalize the result.
	if got := cmp.Compare("hello world", "hello aorld"); got != +1 {
		t.Errorf("Compare(hello world, hello aorld) = %d, want +1", got)
	}
	if got := cmp.Compare("hello aorld", "hello world"); got != -1 {
		t.Errorf("Compare(hello aorld, hello world) = %d, want -1", got)
	}
}

func TestCompare_Named(t *testing.T) {
	var i1 ID = 11
	var i2 ID = 22
	if got := cmp.Compare(i1, i2); got != -1 {
		t.Errorf("Compare(ID(11), ID(22)) = %d, want -1", got)
	}
	if got := cmp.Compare(i1, i1); got != 0 {
		t.Errorf("Compare(ID(11), ID(11)) = %d, want 0", got)
	}
	var n1 Name = "hello"
	var n2 Name = "world"
	if got := cmp.Compare(n1, n2); got != -1 {
		t.Errorf("Compare(Name(hello), Name(world)) = %d, want -1", got)
	}
	if got := cmp.Compare(n1, n1); got != 0 {
		t.Errorf("Compare(Name(hello), Name(hello)) = %d, want 0", got)
	}
	// Less and Equal take the same path, so a named type needs a check there too.
	if !cmp.Less(i1, i2) {
		t.Error("Less(ID(11), ID(22)) = false")
	}
	if cmp.Less(i2, i1) {
		t.Error("Less(ID(22), ID(11)) = true")
	}
	if !cmp.Equal(i1, i1) {
		t.Error("Equal(ID(11), ID(11)) = false")
	}
	if cmp.Equal(i1, i2) {
		t.Error("Equal(ID(11), ID(22)) = true")
	}
	if !cmp.Less(n1, n2) {
		t.Error("Less(Name(hello), Name(world)) = false")
	}
	if !cmp.Equal(n1, n1) {
		t.Error("Equal(Name(hello), Name(hello)) = false")
	}
	if cmp.Equal(n1, n2) {
		t.Error("Equal(Name(hello), Name(world)) = true")
	}
}

func TestCompare_HighByte(t *testing.T) {
	// The operands differ in a high byte only. A comparison at the wrong width
	// reads the low byte of both operands and reports them equal.
	if got := cmp.Compare(uint16(0x0001), uint16(0x0101)); got != -1 {
		t.Errorf("Compare(0x0001, 0x0101) = %d, want -1", got)
	}
	if got := cmp.Compare(uint32(0x00000001), uint32(0x01000001)); got != -1 {
		t.Errorf("Compare(0x00000001, 0x01000001) = %d, want -1", got)
	}
	if got := cmp.Compare(uint64(0x0000000000000001), uint64(0x0100000000000001)); got != -1 {
		t.Errorf("Compare(0x0000000000000001, 0x0100000000000001) = %d, want -1", got)
	}
	if got := cmp.Compare(int32(0x00000001), int32(0x01000001)); got != -1 {
		t.Errorf("Compare(int32 0x00000001, int32 0x01000001) = %d, want -1", got)
	}
	if cmp.Equal(uint32(0x00000001), uint32(0x01000001)) {
		t.Error("Equal(0x00000001, 0x01000001) = true")
	}
	// float32 and float64 hold the same value in a different number of bytes.
	if got := cmp.Compare(float32(1e30), float32(1e-30)); got != +1 {
		t.Errorf("Compare(float32 1e30, float32 1e-30) = %d, want +1", got)
	}
	if got := cmp.Compare(1e300, 1e-300); got != +1 {
		t.Errorf("Compare(1e300, 1e-300) = %d, want +1", got)
	}
}

func TestEqual(t *testing.T) {
	if cmp.Equal(11, 22) {
		t.Error("Equal(11, 22) = true")
	}
	if !cmp.Equal(11, 11) {
		t.Error("Equal(11, 11) = false")
	}
	if cmp.Equal("hello", "world") {
		t.Error("Equal(hello, world) = true")
	}
	if !cmp.Equal("hello", "hello") {
		t.Error("Equal(hello, hello) = false")
	}
	if !cmp.Equal(1.5, 1.5) {
		t.Error("Equal(1.5, 1.5) = false")
	}
	if cmp.Equal(uint16(1), uint16(256)) {
		t.Error("Equal(uint16(1), uint16(256)) = true")
	}
}

func TestEqual_NaN(t *testing.T) {
	nan := math.NaN()
	if !cmp.Equal(nan, nan) {
		t.Error("Equal(NaN, NaN) = false")
	}
	if cmp.Equal(nan, 1.0) {
		t.Error("Equal(NaN, 1.0) = true")
	}
}

func TestEqual_Unsupported(t *testing.T) {
	// FuncFor does not support these types, so Equal compares the raw bytes.
	// Equal does not accept a Triple, because C cannot copy an array.
	p1 := Point{1, 2}
	p2 := Point{1, 3}
	if cmp.Equal(p1, p2) {
		t.Error("Equal(Point{1, 2}, Point{1, 3}) = true")
	}
	if !cmp.Equal(p1, p1) {
		t.Error("Equal(Point{1, 2}, Point{1, 2}) = false")
	}
	if cmp.Equal(true, false) {
		t.Error("Equal(true, false) = true")
	}
	if !cmp.Equal(true, true) {
		t.Error("Equal(true, true) = false")
	}
	// A pointer compares by address.
	v1, v2 := 1, 1
	p3, p4 := &v1, &v2
	if cmp.Equal(p3, p4) {
		t.Error("Equal(&v1, &v2) = true")
	}
	if !cmp.Equal(p3, p3) {
		t.Error("Equal(&v1, &v1) = false")
	}
	p4 = &v1
	if !cmp.Equal(p3, p4) {
		t.Error("Equal(&v1, &v1) through two variables = false")
	}
}

func TestLess(t *testing.T) {
	if !cmp.Less(11, 22) {
		t.Error("Less(11, 22) = false")
	}
	if cmp.Less(22, 11) {
		t.Error("Less(22, 11) = true")
	}
	if cmp.Less(11, 11) {
		t.Error("Less(11, 11) = true")
	}
	if !cmp.Less("hello", "world") {
		t.Error("Less(hello, world) = false")
	}
	if cmp.Less("world", "hello") {
		t.Error("Less(world, hello) = true")
	}
	if !cmp.Less(1.0, 1.1) {
		t.Error("Less(1.0, 1.1) = false")
	}
	if cmp.Less(1.1, 1.0) {
		t.Error("Less(1.1, 1.0) = true")
	}
	if !cmp.Less(int8(-1), int8(1)) {
		t.Error("Less(int8(-1), int8(1)) = false")
	}
}

func TestLess_NaN(t *testing.T) {
	nan := math.NaN()
	if !cmp.Less(nan, 1.0) {
		t.Error("Less(NaN, 1.0) = false")
	}
	if cmp.Less(1.0, nan) {
		t.Error("Less(1.0, NaN) = true")
	}
	if cmp.Less(nan, nan) {
		t.Error("Less(NaN, NaN) = true")
	}
}

func TestFuncFor_Supported(t *testing.T) {
	// Every check assigns the result first. GCC rejects a direct comparison
	// of a function address to nil.
	fn := cmp.FuncFor[int]()
	if fn == nil {
		t.Error("FuncFor[int]() = nil")
	}
	fn = cmp.FuncFor[int8]()
	if fn == nil {
		t.Error("FuncFor[int8]() = nil")
	}
	fn = cmp.FuncFor[int16]()
	if fn == nil {
		t.Error("FuncFor[int16]() = nil")
	}
	fn = cmp.FuncFor[int32]()
	if fn == nil {
		t.Error("FuncFor[int32]() = nil")
	}
	fn = cmp.FuncFor[int64]()
	if fn == nil {
		t.Error("FuncFor[int64]() = nil")
	}
	fn = cmp.FuncFor[uint]()
	if fn == nil {
		t.Error("FuncFor[uint]() = nil")
	}
	fn = cmp.FuncFor[uint8]()
	if fn == nil {
		t.Error("FuncFor[uint8]() = nil")
	}
	fn = cmp.FuncFor[uint16]()
	if fn == nil {
		t.Error("FuncFor[uint16]() = nil")
	}
	fn = cmp.FuncFor[uint32]()
	if fn == nil {
		t.Error("FuncFor[uint32]() = nil")
	}
	fn = cmp.FuncFor[uint64]()
	if fn == nil {
		t.Error("FuncFor[uint64]() = nil")
	}
	fn = cmp.FuncFor[float32]()
	if fn == nil {
		t.Error("FuncFor[float32]() = nil")
	}
	fn = cmp.FuncFor[float64]()
	if fn == nil {
		t.Error("FuncFor[float64]() = nil")
	}
	fn = cmp.FuncFor[string]()
	if fn == nil {
		t.Error("FuncFor[string]() = nil")
	}
	fn = cmp.FuncFor[ID]()
	if fn == nil {
		t.Error("FuncFor[ID]() = nil")
	}
	fn = cmp.FuncFor[Name]()
	if fn == nil {
		t.Error("FuncFor[Name]() = nil")
	}
}

func TestCompare_IntBounds(t *testing.T) {
	// The extreme values of each width. A comparison at the wrong width, or a
	// signed comparison of an unsigned value, gives a different result here.
	if got := cmp.Compare(int8(math.MinInt8), int8(math.MaxInt8)); got != -1 {
		t.Errorf("Compare(MinInt8, MaxInt8) = %d, want -1", got)
	}
	if got := cmp.Compare(int16(math.MinInt16), int16(math.MaxInt16)); got != -1 {
		t.Errorf("Compare(MinInt16, MaxInt16) = %d, want -1", got)
	}
	if got := cmp.Compare(int32(math.MinInt32), int32(math.MaxInt32)); got != -1 {
		t.Errorf("Compare(MinInt32, MaxInt32) = %d, want -1", got)
	}
	if got := cmp.Compare(int64(math.MinInt64), int64(math.MaxInt64)); got != -1 {
		t.Errorf("Compare(MinInt64, MaxInt64) = %d, want -1", got)
	}
	if got := cmp.Compare(int64(math.MaxInt64), int64(math.MinInt64)); got != +1 {
		t.Errorf("Compare(MaxInt64, MinInt64) = %d, want +1", got)
	}
	if got := cmp.Compare(math.MinInt, math.MaxInt); got != -1 {
		t.Errorf("Compare(MinInt, MaxInt) = %d, want -1", got)
	}
	// A subtraction of the two operands overflows here, so the comparison
	// must not subtract.
	if got := cmp.Compare(int64(math.MinInt64), int64(1)); got != -1 {
		t.Errorf("Compare(MinInt64, 1) = %d, want -1", got)
	}
	if got := cmp.Compare(int64(math.MaxInt64), int64(-1)); got != +1 {
		t.Errorf("Compare(MaxInt64, -1) = %d, want +1", got)
	}
	// The high bit of an unsigned value is a value bit, not a sign bit.
	if got := cmp.Compare(uint8(math.MaxUint8), uint8(0)); got != +1 {
		t.Errorf("Compare(MaxUint8, 0) = %d, want +1", got)
	}
	if got := cmp.Compare(uint16(math.MaxUint16), uint16(0)); got != +1 {
		t.Errorf("Compare(MaxUint16, 0) = %d, want +1", got)
	}
	if got := cmp.Compare(uint32(math.MaxUint32), uint32(0)); got != +1 {
		t.Errorf("Compare(MaxUint32, 0) = %d, want +1", got)
	}
	if got := cmp.Compare(uint64(math.MaxUint64), uint64(0)); got != +1 {
		t.Errorf("Compare(MaxUint64, 0) = %d, want +1", got)
	}
	if got := cmp.Compare(uint64(math.MaxUint64), uint64(math.MaxInt64)); got != +1 {
		t.Errorf("Compare(MaxUint64, MaxInt64) = %d, want +1", got)
	}
	if got := cmp.Compare(math.MaxUint, uint(0)); got != +1 {
		t.Errorf("Compare(MaxUint, 0) = %d, want +1", got)
	}
}

func TestCompare_Int8Sweep(t *testing.T) {
	// Every int8 pair. Compare, Less and Equal must agree with each other and
	// with the plain operators, and Compare must give exactly -1, 0 or +1.
	for a := math.MinInt8; a <= math.MaxInt8; a++ {
		for b := math.MinInt8; b <= math.MaxInt8; b++ {
			x, y := int8(a), int8(b)
			want := compareInt8Brute(x, y)
			if got := cmp.Compare(x, y); got != want {
				t.Errorf("Compare(%d, %d) = %d, want %d", a, b, got, want)
				return
			}
			if got := cmp.Less(x, y); got != (want < 0) {
				t.Errorf("Less(%d, %d) = %t, want %t", a, b, got, want < 0)
				return
			}
			if got := cmp.Equal(x, y); got != (want == 0) {
				t.Errorf("Equal(%d, %d) = %t, want %t", a, b, got, want == 0)
				return
			}
			// The order is antisymmetric.
			if got := cmp.Compare(y, x); got != -want {
				t.Errorf("Compare(%d, %d) = %d, want %d", b, a, got, -want)
				return
			}
		}
	}
}

func TestCompare_Uint8Sweep(t *testing.T) {
	// Every uint8 pair. A signed comparison gives a different result for a
	// value with the high bit set.
	for a := 0; a <= int(math.MaxUint8); a++ {
		for b := 0; b <= int(math.MaxUint8); b++ {
			x, y := uint8(a), uint8(b)
			want := compareUint8Brute(x, y)
			if got := cmp.Compare(x, y); got != want {
				t.Errorf("Compare(%d, %d) = %d, want %d", a, b, got, want)
				return
			}
			if got := cmp.Less(x, y); got != (want < 0) {
				t.Errorf("Less(%d, %d) = %t, want %t", a, b, got, want < 0)
				return
			}
			if got := cmp.Equal(x, y); got != (want == 0) {
				t.Errorf("Equal(%d, %d) = %t, want %t", a, b, got, want == 0)
				return
			}
			if got := cmp.Compare(y, x); got != -want {
				t.Errorf("Compare(%d, %d) = %d, want %d", b, a, got, -want)
				return
			}
		}
	}
}

func TestCompare_StringSweep(t *testing.T) {
	// Every pair of short words. A string compares by byte value first and by
	// length second, and a NUL byte does not end a word.
	var abuf, bbuf [maxWord]byte
	total := wordTotal()
	for ai := range total {
		a := wordAt(abuf[:], ai)
		for bi := range total {
			b := wordAt(bbuf[:], bi)
			want := compareStringBrute(a, b)
			got := cmp.Compare(a, b)
			if got != want || cmp.Less(a, b) != (want < 0) || cmp.Equal(a, b) != (want == 0) {
				var d1, d2 [2 * maxWord]byte
				t.Errorf("Compare(%s, %s) = %d, want %d",
					dump(d1[:], a), dump(d2[:], b), got, want)
				return
			}
		}
	}
}

func TestCompare_TotalOrder(t *testing.T) {
	// An insertion sort with Less must give a sequence that Compare accepts as
	// ordered. The values hold every float special case, so the sort passes
	// only if the comparison is a total order.
	var x [8]float64
	x[0] = 1.0
	x[1] = 0.0
	x[2] = math.Copysign(0, -1)
	x[3] = math.Inf(1)
	x[4] = math.Inf(-1)
	x[5] = math.NaN()
	x[6] = -1.0
	x[7] = math.NaN()
	for i := 1; i < len(x); i++ {
		for j := i; j > 0 && cmp.Less(x[j], x[j-1]); j-- {
			tmp := x[j]
			x[j] = x[j-1]
			x[j-1] = tmp
		}
	}
	for i := range len(x) - 1 {
		if cmp.Less(x[i+1], x[i]) {
			t.Errorf("Less sort mismatch at %d", i)
		}
		if cmp.Compare(x[i], x[i+1]) > 0 {
			t.Errorf("Compare sort mismatch at %d", i)
		}
	}
	// The NaNs sort first, and -Inf follows them.
	if !cmp.Equal(x[0], x[0]) {
		t.Error("x[0] is not a NaN")
	}
	if got := cmp.Compare(x[0], math.Inf(-1)); got != -1 {
		t.Errorf("Compare(x[0], -Inf) = %d, want -1", got)
	}
	if got := cmp.Compare(x[2], math.Inf(-1)); got != 0 {
		t.Errorf("Compare(x[2], -Inf) = %d, want 0", got)
	}
	if got := cmp.Compare(x[len(x)-1], math.Inf(1)); got != 0 {
		t.Errorf("Compare(last, Inf) = %d, want 0", got)
	}
}

func TestCompare_Transitive(t *testing.T) {
	// The order is transitive over the string sweep: a < b and b < c give a < c.
	var abuf, bbuf, cbuf [maxWord]byte
	total := wordTotal()
	for ai := range total {
		a := wordAt(abuf[:], ai)
		for bi := range total {
			b := wordAt(bbuf[:], bi)
			if !cmp.Less(a, b) {
				continue
			}
			for ci := range total {
				c := wordAt(cbuf[:], ci)
				if cmp.Less(b, c) && !cmp.Less(a, c) {
					var d1, d2, d3 [2 * maxWord]byte
					t.Errorf("Less(%s, %s) and Less(%s, %s), but not Less(%s, %s)",
						dump(d1[:], a), dump(d2[:], b),
						dump(d2[:], b), dump(d3[:], c),
						dump(d1[:], a), dump(d3[:], c))
					return
				}
			}
		}
	}
}

func TestFuncFor_Call(t *testing.T) {
	// slices calls the returned function with pointers to the values.
	fn := cmp.FuncFor[int]()
	x, y := 11, 22
	if got := fn(&x, &y); got != -1 {
		t.Errorf("FuncFor[int]()(11, 22) = %d, want -1", got)
	}
	if got := fn(&y, &x); got != +1 {
		t.Errorf("FuncFor[int]()(22, 11) = %d, want +1", got)
	}
	if got := fn(&x, &x); got != 0 {
		t.Errorf("FuncFor[int]()(11, 11) = %d, want 0", got)
	}
	// Each type gets a function of its own, and the value keeps its type.
	ufn := cmp.FuncFor[uint8]()
	u1, u2 := uint8(math.MaxUint8), uint8(0)
	if got := ufn(&u1, &u2); got != +1 {
		t.Errorf("FuncFor[uint8]()(255, 0) = %d, want +1", got)
	}
	sfn := cmp.FuncFor[string]()
	s1, s2 := "a", "aa"
	if got := sfn(&s1, &s2); got != -1 {
		t.Errorf("FuncFor[string]()(a, aa) = %d, want -1", got)
	}
	if got := sfn(&s2, &s1); got != +1 {
		t.Errorf("FuncFor[string]()(aa, a) = %d, want +1", got)
	}
	// A named type gets the function of its underlying type.
	ifn := cmp.FuncFor[ID]()
	var id1 ID = 11
	var id2 ID = 22
	if got := ifn(&id1, &id2); got != -1 {
		t.Errorf("FuncFor[ID]()(11, 22) = %d, want -1", got)
	}
}

func TestFuncFor_Unsupported(t *testing.T) {
	// slices.NewSorter and Equal rely on the nil result to fall back to memcmp.
	fn := cmp.FuncFor[bool]()
	if fn != nil {
		t.Error("FuncFor[bool]() != nil")
	}
	fn = cmp.FuncFor[Point]()
	if fn != nil {
		t.Error("FuncFor[Point]() != nil")
	}
	fn = cmp.FuncFor[Triple]()
	if fn != nil {
		t.Error("FuncFor[Triple]() != nil")
	}
}
