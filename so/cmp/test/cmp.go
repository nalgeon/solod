package main

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
}

func TestCompare_FloatZero(t *testing.T) {
	negzero := math.Copysign(0, -1)
	if got := cmp.Compare(negzero, 0.0); got != 0 {
		t.Errorf("Compare(-0.0, 0.0) = %d, want 0", got)
	}
	if got := cmp.Compare(0.0, negzero); got != 0 {
		t.Errorf("Compare(0.0, -0.0) = %d, want 0", got)
	}
	if got := cmp.Compare(negzero, 1.0); got != -1 {
		t.Errorf("Compare(-0.0, 1.0) = %d, want -1", got)
	}
	if cmp.Less(negzero, 0.0) {
		t.Error("Less(-0.0, 0.0) = true, want false")
	}
	if !cmp.Equal(negzero, 0.0) {
		t.Error("Equal(-0.0, 0.0) = false, want true")
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
