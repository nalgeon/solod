package slices_test

import (
	"solod.dev/so/math"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

func TestMake(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Make[int](alloc, 3)
	defer slices.Free(alloc, s)
	s[0] = 11
	s[1] = 22
	s[2] = 33
	if len(s) != 3 || cap(s) != 3 {
		t.Error("unexpected len/cap")
	}
	if s[0] != 11 || s[1] != 22 || s[2] != 33 {
		t.Error("unexpected values")
	}
}

func TestMakeCap(t *testing.T) {
	alloc := t.Allocator()
	s := slices.MakeCap[int](alloc, 2, 8)
	defer slices.Free(alloc, s)
	if len(s) != 2 || cap(s) != 8 {
		t.Error("unexpected len/cap")
	}
	if s[0] != 0 || s[1] != 0 {
		t.Error("not zeroed")
	}
}

func TestAppend(t *testing.T) {
	// Append within capacity.
	alloc := t.Allocator()
	s := slices.MakeCap[int](alloc, 0, 8)
	s = slices.Append(alloc, s, 10, 20, 30)
	if len(s) != 3 || s[0] != 10 || s[1] != 20 || s[2] != 30 {
		t.Error("unexpected values")
	}
	slices.Free(alloc, s)
}

func TestAppend_Grow(t *testing.T) {
	// Append that triggers growth.
	alloc := t.Allocator()
	s := slices.MakeCap[int](alloc, 0, 2)
	s = slices.Append(alloc, s, 1, 2)
	s = slices.Append(alloc, s, 3, 4, 5)
	if len(s) != 5 || s[0] != 1 || s[4] != 5 {
		t.Error("unexpected values")
	}
	slices.Free(alloc, s)
}

func TestAppend_Nil(t *testing.T) {
	// Append to nil slice.
	alloc := t.Allocator()
	var s []int
	s = slices.Append(alloc, s, 10, 20, 30)
	if len(s) != 3 || s[0] != 10 || s[1] != 20 || s[2] != 30 {
		t.Error("unexpected values")
	}
	slices.Free(alloc, s)
}

func TestExtend(t *testing.T) {
	// Extend from another slice.
	alloc := t.Allocator()
	s := slices.MakeCap[int](alloc, 0, 8)
	other := []int{100, 200, 300}
	s = slices.Extend(alloc, s, other)
	if len(s) != 3 || s[0] != 100 || s[2] != 300 {
		t.Error("unexpected values")
	}
	slices.Free(alloc, s)
}

func TestExtend_Nil(t *testing.T) {
	// Extend a nil slice.
	alloc := t.Allocator()
	var s []int
	other := []int{10, 20, 30}
	s = slices.Extend(alloc, s, other)
	if len(s) != 3 || s[0] != 10 || s[1] != 20 || s[2] != 30 {
		t.Error("unexpected values")
	}
	slices.Free(alloc, s)
}

func TestExtend_Empty(t *testing.T) {
	// An empty slice has a nil data pointer.
	// Extend must skip the copy instead of passing the nil pointer on.
	alloc := t.Allocator()
	var empty []int

	s := slices.MakeCap[int](alloc, 0, 4)
	s = slices.Extend(alloc, s, empty)
	if len(s) != 0 || cap(s) != 4 {
		t.Error("unexpected len/cap")
	}

	s = slices.Extend(alloc, s, []int{7, 8})
	s = slices.Extend(alloc, s, empty)
	if len(s) != 2 || s[0] != 7 || s[1] != 8 {
		t.Error("unexpected values")
	}
	slices.Free(alloc, s)

	// Both slices empty.
	var dst []int
	dst = slices.Extend(alloc, dst, empty)
	if len(dst) != 0 {
		t.Error("unexpected len")
	}
	slices.Free(alloc, dst)
}

func TestClone(t *testing.T) {
	alloc := t.Allocator()
	s1 := []int{11, 22, 33}
	s2 := slices.Clone(alloc, s1)
	defer slices.Free(alloc, s2)
	if len(s2) != 3 || cap(s2) != 3 {
		t.Error("unexpected len/cap")
	}
	s2[0] = 99
	if s1[0] != 11 || s2[0] != 99 {
		t.Error("unexpected values")
	}
}

func TestClone_Empty(t *testing.T) {
	// An empty slice has a nil data pointer.
	// Clone must skip the copy instead of passing the nil pointer on.
	alloc := t.Allocator()

	var nilSlice []int
	c1 := slices.Clone(alloc, nilSlice)
	if len(c1) != 0 || cap(c1) != 0 {
		t.Error("unexpected len/cap")
	}
	slices.Free(alloc, c1)

	empty := []int{}
	c2 := slices.Clone(alloc, empty)
	if len(c2) != 0 || cap(c2) != 0 {
		t.Error("unexpected len/cap")
	}
	slices.Free(alloc, c2)
}

func TestEqual(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{1, 2, 3}
	s3 := []int{1, 2, 4}
	s4 := []int{1, 2}
	s5 := []int{}
	var s6 []int = nil
	if !slices.Equal(s1, s2) {
		t.Error("want s1 == s2")
	}
	if slices.Equal(s1, s3) {
		t.Error("want s1 != s3")
	}
	if slices.Equal(s1, s4) {
		t.Error("want s1 != s4")
	}
	if !slices.Equal(s5, s6) {
		t.Error("want empty and nil slices equal")
	}
	if slices.Equal(s1, s6) {
		t.Error("want non-empty and nil slices not equal")
	}
}

func TestEqual_Strings(t *testing.T) {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"a", "b", "c"}
	s3 := []string{"a", "b", "d"}
	if !slices.Equal(s1, s2) {
		t.Error("want s1 == s2")
	}
	if slices.Equal(s1, s3) {
		t.Error("want s1 != s3")
	}
}

func TestEqual_Floats(t *testing.T) {
	// Equal compares floats with cmp.Equal, so a NaN equals a NaN.
	nan := math.NaN()
	s1 := []float64{1, 2, nan}
	s2 := []float64{1, 2, nan}
	if !slices.Equal(s1, s2) {
		t.Error("want NaN == NaN")
	}

	// And -0.0 equals 0.0.
	s3 := []float64{0.0}
	s4 := []float64{math.Copysign(0.0, -1)}
	if !slices.Equal(s3, s4) {
		t.Error("want -0.0 == 0.0")
	}

	s5 := []float64{1, 2, 3}
	if slices.Equal(s1, s5) {
		t.Error("want s1 != s5")
	}
}

func TestEqual_Structs(t *testing.T) {
	type point struct {
		x, y int
	}
	s1 := []point{{1, 2}, {3, 4}}
	s2 := []point{{1, 2}, {3, 4}}
	s3 := []point{{1, 2}, {3, 5}}
	if !slices.Equal(s1, s2) {
		t.Error("want s1 == s2")
	}
	if slices.Equal(s1, s3) {
		t.Error("want s1 != s3")
	}
}

func TestIndex(t *testing.T) {
	ints := []int{10, 20, 30, 20}
	if slices.Index(ints, 20) != 1 {
		t.Error("Index(ints, 20) != 1")
	}
	if slices.Index(ints, 40) != -1 {
		t.Error("Index(ints, 40) != -1")
	}
	strs := []string{"a", "b", "c", "b"}
	if slices.Index(strs, "b") != 1 {
		t.Error("Index(strs, b) != 1")
	}
	if slices.Index(strs, "d") != -1 {
		t.Error("Index(strs, d) != -1")
	}
	var empty []int
	if slices.Index(empty, 1) != -1 {
		t.Error("Index(empty, 1) != -1")
	}
}

func TestContains(t *testing.T) {
	ints := []int{10, 20, 30, 20}
	if !slices.Contains(ints, 20) {
		t.Error("Contains(ints, 20) != true")
	}
	if slices.Contains(ints, 40) {
		t.Error("Contains(ints, 40) != false")
	}
	strs := []string{"a", "b", "c", "b"}
	if !slices.Contains(strs, "b") {
		t.Error("Contains(strs, b) != true")
	}
	if slices.Contains(strs, "d") {
		t.Error("Contains(strs, d) != false")
	}
}

func TestMinMax_Ints(t *testing.T) {
	ints := []int{3, 1, 4, 1, 5, 9}
	if slices.Min(ints) != 1 {
		t.Error("wrong min value")
	}
	if slices.Max(ints) != 9 {
		t.Error("wrong max value")
	}
	one := []int{7}
	if slices.Min(one) != 7 || slices.Max(one) != 7 {
		t.Error("wrong min/max single value")
	}
}

func TestMinMax_Strings(t *testing.T) {
	strs := []string{"banana", "apple", "cherry"}
	if slices.Min(strs) != "apple" {
		t.Error("wrong min value")
	}
	if slices.Max(strs) != "cherry" {
		t.Error("wrong max value")
	}
}

func TestMinMaxFunc(t *testing.T) {
	// MinFunc and MaxFunc return the first of the equal elements.
	pairs := []intPair{{1, 0}, {2, 1}, {1, 2}, {2, 3}}
	gotMin := slices.MinFunc(pairs, intPairCmp)
	if gotMin.a != 1 || gotMin.b != 0 {
		t.Error("wrong min element")
	}
	gotMax := slices.MaxFunc(pairs, intPairCmp)
	if gotMax.a != 2 || gotMax.b != 1 {
		t.Error("wrong max element")
	}
}
