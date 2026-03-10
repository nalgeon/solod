package main

type array [3]int

type box struct {
	nums [3]int
}

func change(a [3]int) {
	a[0] = 42
}

func newBox() box {
	return box{
		nums: [3]int{11, 22, 33},
	}
}

type arange struct {
	lo uint8
	hi uint8
}

var aranges = [16]arange{
	0: {0x10, 0x20},
	1: {0x30, 0x40},
	2: {0x50, 0x60},
}

func main() {
	{
		// Array literals.
		var a [5]int
		_ = a

		a[4] = 100
		x := a[4]
		_ = x

		l := len(a)
		_ = l

		b := [5]int{1, 2, 3, 4, 5}
		_ = b

		c := [...]int{1, 2, 3, 4, 5}
		_ = c

		d := [...]int{100, 3: 400, 500}
		_ = d
	}
	{
		// Array length is fixed and part of the type.
		var a = [3]int{1, 2, 3}
		if len(a) != 3 {
			panic("want len(a) == 3")
		}
		_ = a
		var b = [3]int{1, 2, 3}
		if b != a {
			panic("want b == a")
		}
		var c = [3]int{3, 2, 1}
		if c == a {
			panic("want c != a")
		}
		if c != [3]int{3, 2, 1} {
			panic("want c == {3, 2, 1}")
		}
	}
	{
		// Arrays decay to pointers when passed to functions.
		a := [3]int{1, 2, 3}
		change(a)
		if a[0] != 42 {
			panic("want a[0] == 42")
		}
	}
	{
		// Arrays can be struct fields.
		b := newBox()
		if b.nums[1] != 22 {
			panic("want b.nums[1] == 22")
		}
	}
	{
		// Array-to-array assignment.
		a := [3]int{1, 2, 3}
		b := [3]int{0, 0, 0}
		b = a
		if b[0] != 1 || b[2] != 3 {
			panic("want b == {1, 2, 3}")
		}

		var c [3]int
		c = [3]int{1, 2, 3}
		if c[0] != 1 || c[2] != 3 {
			panic("want c == {1, 2, 3}")
		}

		d := c
		if d[0] != 1 || d[2] != 3 {
			panic("want d == {1, 2, 3}")
		}
	}
	{
		// Arrays can be named types.
		var a array
		a[1] = 42
		if a[1] != 42 {
			panic("want a[1] == 42")
		}
	}
	{
		// Array pointers.
		a := [3]int{1, 2, 3}
		p := &a
		if (*p) != a {
			panic("want p == a")
		}
	}
	{
		// Variable-length arrays are not possible, because
		// Go's type checker resolves n to a constant.
		const n = 3
		_ = n
		a := [n]int{}
		if a[0] != 0 || a[1] != 0 || a[2] != 0 {
			panic("want a == {0, 0, 0}")
		}
		a[0] = 42
		if a[0] != 42 {
			panic("want a[0] == 42")
		}
	}
	{
		// Multi-dimensional arrays.
		var twoD [2][3]int32
		for i := range 2 {
			for j := range 3 {
				twoD[i][j] = int32(i*10 + j + 1)
			}
		}
		if twoD[0][0] != 1 || twoD[1][2] != 13 {
			panic("want twoD == {{1, 2, 3}, {11, 12, 13}}")
		}
		twoD = [2][3]int32{
			{1, 2, 3},
			{11, 12, 13},
		}
		if twoD[0][0] != 1 || twoD[1][2] != 13 {
			panic("want twoD == {{1, 2, 3}, {11, 12, 13}}")
		}
	}
	{
		// For-range over arrays.
		a := [3]int{1, 2, 3}
		sum := 0
		for i := range a {
			sum += a[i]
		}
		if sum != 6 {
			panic("want sum == 6")
		}
		sum = 0
		for _, num := range a {
			sum += num
		}
		if sum != 6 {
			panic("want sum == 6")
		}
		sum = 0
		for i, num := range a {
			_ = i
			sum += num
		}
		if sum != 6 {
			panic("want sum == 6")
		}
		for range a {
		}
	}
	{
		// Array comparisons.
		a := [3]int{1, 2, 3}
		var b [3]int
		b[0] = 1
		b[1] = 2
		b[2] = 3
		if a != b {
			panic("want a == b")
		}
		c := [3]int{3, 2, 1}
		if a == c {
			panic("want a != c")
		}
	}
	_ = aranges
}
