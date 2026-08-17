package slices

import (
	gocmp "cmp"

	"solod.dev/so/c"
	"solod.dev/so/cmp"
	"solod.dev/so/math/bits"
	"solod.dev/so/mem"
)

// Sorter provides comparison and swapping
// operations for a slice of any type.
type Sorter struct {
	slice   Slice
	esize   int
	compare cmp.Func
}

// NewSorter creates a Sorter for a given slice with a custom compare function.
// If compare is nil, compares by raw byte value (memcmp).
//
//so:inline
func NewSorter[T any](s []T, compare cmp.Func) Sorter {
	return Sorter{
		slice:   Header(s),
		esize:   c.Sizeof[T](),
		compare: compare,
	}
}

// Compare compares the elements at indices i and j.
// Returns a negative value if s[i] < s[j], zero if they are equal,
// and a positive value if s[i] > s[j].
func (s Sorter) Compare(i, j int) int {
	a := c.PtrAdd(s.slice.ptr, i*s.esize)
	b := c.PtrAdd(s.slice.ptr, j*s.esize)
	if s.compare != nil {
		return s.compare(a, b)
	}
	return mem.Compare(a, b, s.esize)
}

// Less reports whether the element at index i
// should sort before the element at index j.
func (s Sorter) Less(i, j int) bool {
	return s.Compare(i, j) < 0
}

// Swap swaps the elements at indices i and j.
func (s Sorter) Swap(i, j int) {
	a := c.PtrAdd(s.slice.ptr, i*s.esize)
	b := c.PtrAdd(s.slice.ptr, j*s.esize)
	mem.SwapByte(a, b, s.esize)
}

// Sort sorts a slice of any ordered type in ascending order.
//
//so:inline
func Sort[T gocmp.Ordered](x []T) {
	_s := NewSorter(x, cmp.FuncFor[T]())
	SortWith(_s)
}

// SortFunc sorts the slice x in ascending order as determined by the cmp
// function. This sort is not guaranteed to be stable.
// cmp(a, b) should return a negative number when a < b, a positive number when
// a > b and zero when a == b or a and b are incomparable in the sense of
// a strict weak ordering.
//
// SortFunc requires that cmp is a strict weak ordering.
// See https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings.
// The function should return 0 for incomparable items.
//
//so:inline
func SortFunc[T any](x []T, compare cmp.Func) {
	_s := NewSorter(x, compare)
	SortWith(_s)
}

// SortWith sorts the slice using the provided Sorter.
func SortWith(s Sorter) {
	limit := bits.Len(uint(s.slice.len))
	pdqsort_func(s, 0, s.slice.len, limit)
}

// SortStableFunc sorts the slice x while keeping the original order of equal
// elements, using cmp to compare elements in the same way as [SortFunc].
//
//so:inline
func SortStableFunc[T any](x []T, compare cmp.Func) {
	_s := NewSorter(x, compare)
	SortStableWith(_s)
}

// SortStableWith sorts the slice using the provided Sorter
// while keeping the original order of equal elements.
func SortStableWith(s Sorter) {
	stable_func(s, s.slice.len)
}

// IsSorted reports whether x is sorted in ascending order.
//
//so:inline
func IsSorted[T gocmp.Ordered](x []T) bool {
	_s := NewSorter(x, cmp.FuncFor[T]())
	return IsSortedWith(_s)
}

// IsSortedFunc reports whether x is sorted in ascending order, with cmp as the
// comparison function as defined by [SortFunc].
//
//so:inline
func IsSortedFunc[T any](x []T, compare cmp.Func) bool {
	_s := NewSorter(x, compare)
	return IsSortedWith(_s)
}

// IsSortedWith reports whether the slice is sorted
// according to the provided Sorter.
func IsSortedWith(s Sorter) bool {
	for i := s.slice.len - 1; i > 0; i-- {
		if s.Compare(i, i-1) < 0 {
			return false
		}
	}
	return true
}

// Min returns the minimal value in x. It panics if x is empty.
// For floating-point numbers, Min propagates NaNs (any NaN value in x
// forces the output to be NaN).
//
//so:inline
func Min[T gocmp.Ordered](x []T) T {
	return MinFunc(x, cmp.FuncFor[T]())
}

// MinFunc returns the minimal value in x, using cmp to compare elements.
// It panics if x is empty. If there is more than one minimal element
// according to the cmp function, MinFunc returns the first one.
// If cmp is nil, compares by raw byte value (memcmp).
//
//so:inline
func MinFunc[T any](x []T, compare cmp.Func) T {
	_x, _cmp := x, compare
	if len(_x) < 1 {
		panic("slices: empty list")
	}
	_esize := c.Sizeof[T]()
	_m := _x[0]
	for _j := 1; _j < len(_x); _j++ {
		_xj := _x[_j]
		var _c int
		if _cmp != nil {
			_c = _cmp(&_xj, &_m)
		} else {
			_c = mem.Compare(&_xj, &_m, _esize)
		}
		if _c < 0 {
			_m = _xj
		}
	}
	return _m
}

// Max returns the maximal value in x. It panics if x is empty.
// For floating-point numbers, Max ignores NaNs. Max returns a NaN
// only if every element in x is a NaN. This differs from Min,
// which returns a NaN if any element in x is a NaN.
//
//so:inline
func Max[T gocmp.Ordered](x []T) T {
	return MaxFunc(x, cmp.FuncFor[T]())
}

// MaxFunc returns the maximal value in x, using cmp to compare elements.
// It panics if x is empty. If there is more than one maximal element
// according to the cmp function, MaxFunc returns the first one.
// If cmp is nil, compares by raw byte value (memcmp).
//
//so:inline
func MaxFunc[T any](x []T, compare cmp.Func) T {
	_x, _cmp := x, compare
	if len(_x) < 1 {
		panic("slices: empty list")
	}
	_esize := c.Sizeof[T]()
	_m := _x[0]
	for _j := 1; _j < len(_x); _j++ {
		_xj := _x[_j]
		var _c int
		if _cmp != nil {
			_c = _cmp(&_xj, &_m)
		} else {
			_c = mem.Compare(&_xj, &_m, _esize)
		}
		if _c > 0 {
			_m = _xj
		}
	}
	return _m
}
