package main

import (
	"solod.dev/so/cmp"
	"solod.dev/so/math/bits"
	"solod.dev/so/math/rand"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

func descInt(a, b any) int {
	va := *a.(*int)
	vb := *b.(*int)
	return vb - va
}

var sortInts = [...]int{74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586}
var sortFloat64s = [...]float64{74.3, 59.0, 238.2, -784.0, 2.3, 9845.768, -959.7485, 905, 7.8, 7.8, 74.3, 59.0, 238.2, -784.0, 2.3}
var sortStrs = [...]string{"", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&", "***"}

func TestIsSorted(t *testing.T) {
	// False on unsorted data.
	if slices.IsSorted(sortInts[:]) {
		t.Error("unsorted ints")
	}
	if slices.IsSorted(sortStrs[:]) {
		t.Error("unsorted strs")
	}
	// True on sorted data.
	sorted := []int{1, 2, 3, 4, 5}
	if !slices.IsSorted(sorted) {
		t.Error("sorted ints")
	}
	sortedStrs := []string{"a", "b", "c"}
	if !slices.IsSorted(sortedStrs) {
		t.Error("sorted strs")
	}
	// True on an empty or single-element slice.
	var empty []int
	if !slices.IsSorted(empty) {
		t.Error("empty")
	}
	one := []int{7}
	if !slices.IsSorted(one) {
		t.Error("single element")
	}
}

func TestIsSortedFunc(t *testing.T) {
	// False on unsorted data.
	compare := cmp.FuncFor[int]()
	if slices.IsSortedFunc(sortInts[:], compare) {
		t.Error("unsorted ints")
	}
	// True on sorted data.
	sorted := []int{1, 2, 3, 4, 5}
	if !slices.IsSortedFunc(sorted, compare) {
		t.Error("sorted ints")
	}
}

func TestSort_Ints(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortInts[:])
	defer slices.Free(alloc, s)
	slices.Sort(s)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != -5467984 || s[12] != 9845 {
		t.Error("wrong values")
	}
}

func TestSort_Float64s(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortFloat64s[:])
	defer slices.Free(alloc, s)
	slices.Sort(s)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != -959.7485 || s[14] != 9845.768 {
		t.Error("wrong values")
	}
}

func TestSort_Strings(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortStrs[:])
	defer slices.Free(alloc, s)
	slices.Sort(s)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != "" || s[7] != "foo" {
		t.Error("wrong values")
	}
}

func TestSort_Empty(t *testing.T) {
	var empty []int
	slices.Sort(empty)
	one := []int{7}
	slices.Sort(one)
	if one[0] != 7 {
		t.Error("one: wrong value")
	}
	two := []int{2, 1}
	slices.Sort(two)
	if two[0] != 1 || two[1] != 2 {
		t.Error("two: wrong values")
	}
}

func TestSortFunc(t *testing.T) {
	// SortFunc (reverse order).
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortInts[:])
	defer slices.Free(alloc, s)
	slices.SortFunc(s, descInt)
	if !slices.IsSortedFunc(s, descInt) {
		t.Error("not sorted")
	}
	if s[0] != 9845 || s[12] != -5467984 {
		t.Error("wrong values")
	}
}

func TestSortFunc_Nil(t *testing.T) {
	// SortFunc with nil compare.
	type point struct{ x, y int }
	s := []point{{1, 2}, {3, 4}, {2, 3}}
	slices.SortFunc(s, nil)
	if !slices.IsSortedFunc(s, nil) {
		t.Error("not sorted")
	}
	if s[0].x != 1 || s[0].y != 2 {
		t.Error("wrong s[0]")
	}
	if s[1].x != 2 || s[1].y != 3 {
		t.Error("wrong s[1]")
	}
	if s[2].x != 3 || s[2].y != 4 {
		t.Error("wrong s[2]")
	}
}

func TestSortStableFunc_Ints(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortInts[:])
	defer slices.Free(alloc, s)
	compare := cmp.FuncFor[int]()
	slices.SortStableFunc(s, compare)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != -5467984 || s[12] != 9845 {
		t.Error("wrong values")
	}
}

func TestSortStableFunc_Float64s(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortFloat64s[:])
	defer slices.Free(alloc, s)
	compare := cmp.FuncFor[float64]()
	slices.SortStableFunc(s, compare)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != -959.7485 || s[14] != 9845.768 {
		t.Error("wrong values")
	}
}

func TestSortStableFunc_Strings(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortStrs[:])
	defer slices.Free(alloc, s)
	compare := cmp.FuncFor[string]()
	slices.SortStableFunc(s, compare)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != "" || s[7] != "foo" {
		t.Error("wrong values")
	}
}

// -- Sorter --

func TestSorter(t *testing.T) {
	// Compare, Less and Swap work on the element at the given index.
	s := []int{30, 10, 20}
	sorter := slices.NewSorter(s, cmp.FuncFor[int]())

	if sorter.Compare(0, 1) <= 0 {
		t.Error("Sorter.Compare: want s[0] > s[1]")
	}
	if sorter.Compare(1, 2) >= 0 {
		t.Error("Sorter.Compare: want s[1] < s[2]")
	}
	if sorter.Compare(0, 0) != 0 {
		t.Error("Sorter.Compare: want s[0] == s[0]")
	}
	if sorter.Less(0, 1) {
		t.Error("Sorter.Less: want false")
	}
	if !sorter.Less(1, 2) {
		t.Error("Sorter.Less: want true")
	}

	sorter.Swap(0, 1)
	if s[0] != 10 || s[1] != 30 {
		t.Error("Sorter.Swap: wrong values")
	}
}

func TestSorter_NilCompare(t *testing.T) {
	// A nil compare function falls back to a raw byte comparison.
	type point struct{ x, y int }
	s := []point{{2, 0}, {1, 0}}
	sorter := slices.NewSorter(s, nil)
	if sorter.Compare(0, 1) <= 0 {
		t.Error("Sorter.Compare: want s[0] > s[1]")
	}
	sorter.Swap(0, 1)
	if s[0].x != 1 || s[1].x != 2 {
		t.Error("Sorter.Swap: wrong values")
	}
}

func TestSortWith(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortInts[:])
	defer slices.Free(alloc, s)

	sorter := slices.NewSorter(s, cmp.FuncFor[int]())
	if slices.IsSortedWith(sorter) {
		t.Error("IsSortedWith: want false before the sort")
	}
	slices.SortWith(sorter)
	if !slices.IsSortedWith(sorter) {
		t.Error("IsSortedWith: want true after the sort")
	}
	if s[0] != -5467984 || s[12] != 9845 {
		t.Error("SortWith: wrong values")
	}
}

func TestSortStableWith(t *testing.T) {
	alloc := t.Allocator()
	s := slices.Clone(alloc, sortInts[:])
	defer slices.Free(alloc, s)

	sorter := slices.NewSorter(s, cmp.FuncFor[int]())
	slices.SortStableWith(sorter)
	if !slices.IsSortedWith(sorter) {
		t.Error("not sorted")
	}
	if s[0] != -5467984 || s[12] != 9845 {
		t.Error("wrong values")
	}
}

// -- Large inputs --

// A sort of 12 elements or less is an insertion sort. A stable sort of less
// than 20 elements is an insertion sort too. The tests below use larger inputs
// to reach the partitioning, the heap sort and the symmetric merge.

type intPair struct {
	a, b int
}

// intPairCmp compares two pairs on the first field only.
func intPairCmp(x, y any) int {
	vx := *x.(*intPair)
	vy := *y.(*intPair)
	return vx.a - vy.a
}

// inOrder reports whether the pairs with an equal first field
// kept their original order.
func inOrder(d []intPair) bool {
	lastA, lastB := -1, 0
	for i := range d {
		if lastA != d[i].a {
			lastA = d[i].a
			lastB = d[i].b
			continue
		}
		if d[i].b <= lastB {
			return false
		}
		lastB = d[i].b
	}
	return true
}

// initB records the original position of each pair in the second field.
func initB(d []intPair) {
	for i := range d {
		d[i].b = i
	}
}

func TestSort_LargeRandom(t *testing.T) {
	// A large random input drives the partitioning and the heap sort.
	const n = 10000
	alloc := t.Allocator()
	s := slices.Make[int](alloc, n)
	defer slices.Free(alloc, s)

	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	for i := range s {
		s[i] = r.IntN(100)
	}
	if slices.IsSorted(s) {
		t.Fatal("the input is already sorted")
		return
	}

	slices.Sort(s)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
}

func TestSort_LargeEqual(t *testing.T) {
	// Many equal elements drive the equal-element partitioning.
	const n = 1000
	alloc := t.Allocator()
	s := slices.Make[int](alloc, n)
	defer slices.Free(alloc, s)
	for i := range s {
		s[i] = 42
	}
	s[0] = 1
	s[n-1] = 0

	slices.Sort(s)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != 0 || s[1] != 1 || s[n-1] != 42 {
		t.Error("wrong values")
	}
}

func TestSort_LargeReversed(t *testing.T) {
	// A reversed input drives the decreasing-order path.
	const n = 1000
	alloc := t.Allocator()
	s := slices.Make[int](alloc, n)
	defer slices.Free(alloc, s)
	for i := range s {
		s[i] = n - i
	}

	slices.Sort(s)
	if !slices.IsSorted(s) {
		t.Error("not sorted")
	}
	if s[0] != 1 || s[n-1] != n {
		t.Error("wrong values")
	}
}

func TestSortStable_Random(t *testing.T) {
	// A large random input drives the symmetric merge.
	const n, m = 1000, 100
	alloc := t.Allocator()
	s := slices.Make[intPair](alloc, n)
	defer slices.Free(alloc, s)

	pcg := rand.NewPCG(1, 2)
	r := rand.New(&pcg)
	for i := range s {
		s[i].a = r.IntN(m)
	}
	if slices.IsSortedFunc(s, intPairCmp) {
		t.Fatal("the input is already sorted")
		return
	}

	initB(s)
	slices.SortStableFunc(s, intPairCmp)
	if !slices.IsSortedFunc(s, intPairCmp) {
		t.Error("not sorted")
	}
	if !inOrder(s) {
		t.Error("not stable")
	}
}

func TestSortStable_Sorted(t *testing.T) {
	// An already sorted input must keep its order.
	const n, m = 1000, 100
	alloc := t.Allocator()
	s := slices.Make[intPair](alloc, n)
	defer slices.Free(alloc, s)
	for i := range s {
		s[i].a = i / (n / m)
	}

	initB(s)
	slices.SortStableFunc(s, intPairCmp)
	if !slices.IsSortedFunc(s, intPairCmp) {
		t.Error("not sorted")
	}
	if !inOrder(s) {
		t.Error("not stable")
	}
}

func TestSortStable_Reversed(t *testing.T) {
	// A reversed input drives the rotation inside the merge.
	const n = 1000
	alloc := t.Allocator()
	s := slices.Make[intPair](alloc, n)
	defer slices.Free(alloc, s)
	for i := range s {
		s[i].a = n - i
	}

	initB(s)
	slices.SortStableFunc(s, intPairCmp)
	if !slices.IsSortedFunc(s, intPairCmp) {
		t.Error("not sorted")
	}
	if !inOrder(s) {
		t.Error("not stable")
	}
}

// -- Adversary --

// The adversary below is based on the antiquicksort program by M. Douglas
// McIlroy. See https://www.cs.dartmouth.edu/~doug/mdmspe.pdf for more info.
// It makes every pivot choice a bad one, which takes a plain quicksort to
// O(n^2) comparisons. The test asserts that the sort stays at O(n*log(n)).
//
// The sort defeats the adversary with breakPatterns, which shuffles the input
// with a random value. The heap sort runs only after breakPatterns fails many
// times in a row, so no input reaches the heap sort. The Go slices package does
// not test the heap sort either.
var (
	// advGas is the value of an element with no order yet.
	// It is higher than every other value.
	advGas int
	// advSolid is the count of the elements with an order.
	advSolid int
	// advCandidate is the current guess at the pivot.
	// It holds the address of a slice element, so a swap does not change it.
	advCandidate any
	// advCmps is the count of the comparisons.
	advCmps int
)

func advCompare(a, b any) int {
	advCmps++
	pa := a.(*int)
	pb := b.(*int)

	// Both elements have no order yet, so give one of them an order.
	// Keep the pivot candidate free to stay the largest element.
	if *pa == advGas && *pb == advGas {
		if a == advCandidate {
			*pa = advSolid
		} else {
			*pb = advSolid
		}
		advSolid++
	}

	// The element with no order becomes the next pivot candidate.
	if *pa == advGas {
		advCandidate = a
	} else if *pb == advGas {
		advCandidate = b
	}

	return *pa - *pb
}

func TestSort_Adversary(t *testing.T) {
	const n = 10000
	alloc := t.Allocator()
	s := slices.Make[int](alloc, n)
	defer slices.Free(alloc, s)

	advGas = n - 1
	advSolid = 0
	advCandidate = nil
	advCmps = 0
	for i := range s {
		s[i] = advGas
	}

	slices.SortFunc(s, advCompare)

	// The adversary gives element i the value i, so a sorted
	// result holds the values 0 to n-1 in order.
	for i := range s {
		if s[i] != i {
			t.Fatal("not sorted")
			return
		}
	}

	// The factor 4 comes from the Go test. A quadratic sort
	// needs about n*n/2 comparisons, which is far above this bound.
	maxCmps := n * bits.Len(uint(n)) * 4
	if advCmps > maxCmps {
		t.Error("too many comparisons")
	}
}
