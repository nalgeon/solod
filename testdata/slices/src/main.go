package main

type Pair struct {
	x int
	y int
}

func lenInt64(buf []byte) (int64, error) {
	n, _ := lenInt64Impl(buf)
	return n, nil
}

func lenInt64Impl(buf []byte) (int64, error) {
	return int64(len(buf)), nil
}

func sumSlice(s []int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func modifySlice(s []int) {
	s[0] = 99
	s[1] = 88
}

func acceptSlice(s []int) {
	_ = s
}

type Bytes []byte

func nilSlice() []int {
	return nil
}

func sumVariadic(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

type SliceHolder struct {
	nums []int
}

func (h SliceHolder) sum() int {
	s := 0
	for _, v := range h.nums {
		s += v
	}
	return s
}

func (h SliceHolder) get(i int) int {
	return h.nums[i]
}

type IntSlice []int

func main() {
	{
		// Slicing an array: all forms.
		nums := [...]int{1, 2, 3, 4, 5}

		s1 := nums[:]
		s1[1] = 200
		_ = s1

		s2 := nums[2:]
		_ = s2

		s3 := nums[:3]
		_ = s3

		s4 := nums[1:4]
		_ = s4

		n := copy(s4, s1) // n == 3
		_ = n
	}
	{
		// Slicing a string.
		str := "hello"
		s1 := str[:]
		if s1 != "hello" {
			panic("want s1 == hello")
		}

		s2 := str[2:]
		if s2 != "llo" {
			panic("want s2 == llo")
		}

		s3 := str[:3]
		if s3 != "hel" {
			panic("want s3 == hel")
		}

		s4 := str[1:4]
		if s4 != "ell" {
			panic("want s4 == ell")
		}
	}
	{
		// Slicing a slice: all forms.
		nums := []int{1, 2, 3, 4, 5}

		s1 := nums[:]
		if s1[0] != 1 || s1[4] != 5 {
			panic("want s1[0] == 1 && s1[4] == 5")
		}

		s2 := nums[2:]
		if s2[0] != 3 || s2[2] != 5 {
			panic("want s2[0] == 3 && s2[2] == 5")
		}

		s3 := nums[:3]
		if s3[0] != 1 || s3[2] != 3 {
			panic("want s3[0] == 1 && s3[2] == 3")
		}

		s4 := nums[1:4]
		if s4[0] != 2 || s4[2] != 4 {
			panic("want s4[0] == 2 && s4[2] == 4")
		}
	}
	{
		// Three-index slice expression.
		nums := []int{1, 2, 3, 4, 5}
		s := nums[1:3:4]
		if len(s) != 2 || cap(s) != 3 {
			panic("want len 2, cap 3")
		}
		if s[0] != 2 || s[1] != 3 {
			panic("want s[0] == 2 && s[1] == 3")
		}
	}
	{
		// Slice literals.
		var nils []int = nil
		if nils != nil {
			panic("want nils == nil")
		}
		if len(nils) != 0 {
			panic("want len(nils) == 0")
		}

		empty := []int{}
		if len(empty) != 0 {
			panic("want len(empty) == 0")
		}

		strSlice := []string{"a", "b", "c"}
		sLen := len(strSlice) // sLen == 3
		_ = sLen

		twoD := [][]int{
			{1, 2, 3},
			{4, 5, 6},
		}
		x := twoD[0][1] // x == 2
		_ = x
	}
	{
		// Make a slice: len only.
		s := make([]int, 4)
		s[0] = 1
		s[1] = 2
		s[2] = 3
		s[3] = 4
		if s[0] != 1 || s[3] != 4 {
			panic("want s[0]==1, s[3]==4")
		}
		if len(s) != 4 {
			panic("want len==4")
		}
	}
	{
		// Make a slice: len and cap.
		s := make([]int, 0, 8)
		if len(s) != 0 || cap(s) != 8 {
			panic("want len==0, cap==8")
		}
		s = append(s, 10)
		if len(s) != 1 || s[0] != 10 {
			panic("want len==1, s[0]==10")
		}
		if cap(s) != 8 {
			panic("want cap still 8")
		}
	}
	{
		// Make with string element type.
		s := make([]string, 3)
		s[0] = "hello"
		s[1] = "world"
		s[2] = "!"
		if s[0] != "hello" || s[2] != "!" {
			panic("want make string slice")
		}
	}
	{
		// Append: single value.
		s := make([]int, 0, 4)
		s = append(s, 1)
		s = append(s, 2)
		if len(s) != 2 || s[0] != 1 || s[1] != 2 {
			panic("want append single")
		}
	}
	{
		// Append: multiple values.
		s := make([]int, 0, 8)
		s = append(s, 1, 2, 3)
		if len(s) != 3 || s[0] != 1 || s[2] != 3 {
			panic("want append multi")
		}
	}
	{
		// Append: spread another slice.
		s := make([]int, 0, 8)
		s = append(s, 1, 2)
		other := []int{3, 4, 5}
		s = append(s, other...)
		if len(s) != 5 || s[2] != 3 || s[4] != 5 {
			panic("want append spread")
		}
	}
	{
		// Append: string to byte slice.
		b := make([]byte, 0, 8)
		s := "hello"
		b = append(b, s...)
		if len(b) != 5 {
			panic("len(b) != 5")
		}
		if string(b) != "hello" {
			panic("string(b) != hello")
		}
	}
	{
		// Append: strings.
		s := make([]string, 0, 4)
		s = append(s, "hello")
		s = append(s, "world")
		if len(s) != 2 || s[0] != "hello" || s[1] != "world" {
			panic("want append strings")
		}
	}
	{
		// Cap: literal slice.
		s := []int{1, 2, 3}
		if cap(s) != 3 {
			panic("want cap(literal)==3")
		}
	}
	{
		// Cap: make with len only.
		s := make([]int, 5)
		if cap(s) != 5 {
			panic("want cap(make)==5")
		}
	}
	{
		// Cap: make with len and cap.
		s := make([]int, 2, 10)
		if len(s) != 2 || cap(s) != 10 {
			panic("want len==2, cap==10")
		}
	}
	{
		// Cap: sub-slice shares capacity.
		s := make([]int, 5, 10)
		s2 := s[2:]
		if len(s2) != 3 || cap(s2) != 8 {
			panic("want sub-slice cap")
		}
	}
	{
		// Len: after append.
		s := make([]int, 0, 8)
		if len(s) != 0 {
			panic("want len==0 before append")
		}
		s = append(s, 1)
		if len(s) != 1 {
			panic("want len==1 after append")
		}
		s = append(s, 2, 3)
		if len(s) != 3 {
			panic("want len==3 after multi append")
		}
	}
	{
		// Len: in expression.
		s := []int{1, 2, 3, 4}
		n := len(s) + 1
		if n != 5 {
			panic("want len in expr")
		}
	}
	{
		// Pass and return slices.
		var buf [4]byte
		n, _ := lenInt64(buf[:])
		if n != 4 {
			panic("want 4")
		}
		n, _ = lenInt64([]byte{1, 2, 3})
		if n != 3 {
			panic("want 3")
		}
	}
	{
		// Pass slice to function: reads.
		s := []int{10, 20, 30}
		if sumSlice(s) != 60 {
			panic("want sumSlice==60")
		}
	}
	{
		// Pass slice to function: modification (reference semantics).
		s := []int{1, 2, 3}
		modifySlice(s)
		if s[0] != 99 || s[1] != 88 {
			panic("want modified slice")
		}
	}
	{
		// Variadic function: individual args.
		if sumVariadic(1, 2, 3) != 6 {
			panic("want variadic sum==6")
		}
	}
	{
		// Variadic function: spread slice.
		nums := []int{10, 20, 30}
		if sumVariadic(nums...) != 60 {
			panic("want variadic spread==60")
		}
	}
	{
		// Number operations on slice elements.
		s := []int{1, 2, 3}
		s[1] += 10
		s[1] -= 10
		s[1] *= 10
		s[1] /= 2
		s[1] %= 6
		s[1]++
		s[1]--
		if s[1] != 4 {
			panic("want 4")
		}
	}
	{
		// Bitwise operations on slice elements.
		s := []int{1, 2, 3}
		s[1] <<= 2
		s[1] >>= 1
		s[1] |= 0b1100
		s[1] &= 0b1111
		s[1] ^= 0b0101
		if s[1] != 9 {
			panic("want 9")
		}
	}
	{
		// Slice element in comparison.
		s := []int{10, 20, 30}
		if s[0] > s[1] {
			panic("want s[0] <= s[1]")
		}
		if s[2] < s[1] {
			panic("want s[2] >= s[1]")
		}
		if s[0] == s[1] {
			panic("want s[0] != s[1]")
		}
	}
	{
		// Slice element in arithmetic expression.
		s := []int{2, 3, 5}
		result := s[0]*s[1] + s[2]
		if result != 11 {
			panic("want 2*3+5==11")
		}
	}
	{
		// Copying a slice.
		s := make([]string, 3, 6)
		s[0] = "a"
		s[1] = "b"
		s[2] = "c"
		c := make([]string, len(s))
		copy(c, s)
		if c[0] != "a" || c[2] != "c" {
			panic("want c[0] == 'a' && c[2] == 'c'")
		}
	}
	{
		// Copy: return value (partial copy when dst is smaller).
		src := []int{1, 2, 3, 4, 5}
		dst := make([]int, 3)
		n := copy(dst, src)
		if n != 3 {
			panic("want copy returned 3")
		}
		if dst[0] != 1 || dst[2] != 3 {
			panic("want partial copy values")
		}
	}
	{
		// Copy: return value (partial copy when src is smaller).
		src := []int{1, 2}
		dst := make([]int, 5)
		n := copy(dst, src)
		if n != 2 {
			panic("want copy returned 2")
		}
		if dst[0] != 1 || dst[1] != 2 || dst[2] != 0 {
			panic("want partial copy src smaller")
		}
	}
	{
		// Copying a string to a byte slice.
		str := "hello"
		b := make([]byte, len(str))
		copy(b, str)
		if string(b) != "hello" {
			panic("want string(b) == 'hello'")
		}

		// Copying a string literal to a byte slice.
		b2 := make([]byte, 2)
		copy(b2, "ab")
		if string(b2) != "ab" {
			panic("want string(b2) == 'ab'")
		}
	}
	{
		// Element types: byte.
		s := []byte{0x41, 0x42, 0x43}
		if s[0] != 0x41 || s[2] != 0x43 {
			panic("want byte slice")
		}
	}
	{
		// Element types: bool.
		s := []bool{true, false, true}
		if !s[0] || s[1] || !s[2] {
			panic("want bool slice")
		}
	}
	{
		// Element types: float64.
		s := []float64{1.5, 2.5, 3.5}
		sum := s[0] + s[1] + s[2]
		if sum != 7.5 {
			panic("want float64 sum==7.5")
		}
	}
	{
		// Element types: rune.
		s := []rune{'a', 'b', 'c'}
		if s[0] != 'a' || s[2] != 'c' {
			panic("want rune slice")
		}
	}
	{
		// Element types: struct.
		s := []Pair{{1, 2}, {3, 4}}
		if s[0].x != 1 || s[1].y != 4 {
			panic("want struct slice")
		}
		s[0].x = 10
		if s[0].x != 10 {
			panic("want modified struct field")
		}
	}
	{
		// Element types: pointer.
		a := 42
		b := 99
		s := []*int{&a, &b}
		if *s[0] != 42 || *s[1] != 99 {
			panic("want pointer slice")
		}
		*s[0] = 100
		if a != 100 {
			panic("want modified through pointer")
		}
	}
	{
		// Element types: string.
		s := []string{"hello", "world", "!"}
		if s[0] != "hello" || s[2] != "!" {
			panic("want string slice values")
		}
		if len(s) != 3 {
			panic("want string slice len==3")
		}
	}
	{
		// 2D slice: access and modify.
		twoD := [][]int{
			{1, 2, 3},
			{4, 5, 6},
		}
		if twoD[0][0] != 1 || twoD[1][2] != 6 {
			panic("want 2D values")
		}
		twoD[0][1] = 20
		if twoD[0][1] != 20 {
			panic("want 2D modified")
		}
	}
	{
		// Nil slice: comparison. Nil works on either side.
		var s []int
		if s != nil {
			panic("want nil slice")
		}
		if nil != s {
			panic("want nil slice")
		}
		s = []int{1}
		if s == nil {
			panic("want non-nil slice")
		}
		if nil == s {
			panic("want non-nil slice")
		}
	}
	{
		// Nil slice: a capacity of 0 means nil.
		nums := []int{1, 2, 3}
		if nums[2:2:2] != nil {
			panic("want empty full slice expression == nil")
		}
		var es string
		if []byte(es) != nil {
			panic("want []byte(\"\") == nil")
		}
		if nums[:0] == nil {
			panic("want nums[:0] != nil")
		}
	}
	{
		// Nil slice: len and cap.
		var s []int
		if len(s) != 0 {
			panic("want nil len==0")
		}
		if cap(s) != 0 {
			panic("want nil cap==0")
		}
	}
	{
		// Nil slice: pass to function.
		acceptSlice(nil)
	}
	{
		// Nil slice: return from function.
		s := nilSlice()
		if s != nil {
			panic("want nil from function")
		}
	}
	{
		// Nil slice: convert nil to a slice.
		s := ([]byte)(nil)
		if s != nil || len(s) != 0 {
			panic("want ([]byte)(nil) == nil")
		}
		named := Bytes(nil)
		if named != nil || len(named) != 0 {
			panic("want Bytes(nil) == nil")
		}
		if len(string(([]byte)(nil))) != 0 {
			panic("want string(([]byte)(nil)) == \"\"")
		}
	}
	{
		// Slice assigned to another variable (shared backing).
		s1 := []int{1, 2, 3}
		s2 := s1
		s2[0] = 99
		if s1[0] != 99 {
			panic("want shared backing")
		}
	}
	{
		// Struct with slice field.
		h := SliceHolder{nums: []int{10, 20, 30}}
		if h.get(0) != 10 || h.get(2) != 30 {
			panic("want struct slice field get")
		}
		if h.sum() != 60 {
			panic("want struct slice field sum")
		}
	}
	{
		// Named slice type: literal.
		s := IntSlice{10, 20, 30}
		if s[0] != 10 || s[2] != 30 {
			panic("want named type literal")
		}
		if len(s) != 3 {
			panic("want named type len")
		}
	}
	{
		// Named slice type: make.
		s := make(IntSlice, 0, 4)
		s = append(s, 1, 2)
		if len(s) != 2 || s[0] != 1 || s[1] != 2 {
			panic("want named type make+append")
		}
	}
	{
		// Named slice type: range.
		s := IntSlice{1, 2, 3}
		sum := 0
		for _, v := range s {
			sum += v
		}
		if sum != 6 {
			panic("want named type range")
		}
	}
	{
		// For-range over slices: index only.
		s := []int{1, 2, 3}
		sum := 0
		for i := range s {
			sum += s[i]
		}
		if sum != 6 {
			panic("want sum == 6")
		}
	}
	{
		// For-range over slices: value only.
		s := []int{1, 2, 3}
		sum := 0
		for _, num := range s {
			sum += num
		}
		if sum != 6 {
			panic("want sum == 6")
		}
	}
	{
		// For-range over slices: index and value.
		s := []int{1, 2, 3}
		sum := 0
		for i, num := range s {
			_ = i
			sum += num
		}
		if sum != 6 {
			panic("want sum == 6")
		}
	}
	{
		// For-range: empty body.
		s := []int{1, 2, 3}
		for range s {
		}
	}
	{
		// For-range: assign (not define).
		s := []int{10, 20, 30}
		i := 0
		v := 0
		sum := 0
		for i, v = range s {
			sum += i + v
		}
		// (0+10) + (1+20) + (2+30) = 63
		if sum != 63 {
			panic("want range assign sum==63")
		}
		// The loop assigns to i and v, so both
		// keep the values of the last iteration.
		if i != 2 || v != 30 {
			panic("want i == 2 && v == 30")
		}
		// An empty slice leaves i and v alone.
		var empty []int
		for i, v = range empty {
		}
		if i != 2 || v != 30 {
			panic("want i == 2 && v == 30")
		}
	}
	{
		// For-range over string slice.
		s := []string{"a", "b", "c"}
		result := ""
		for _, v := range s {
			result += v
		}
		if result != "abc" {
			panic("want range string concat")
		}
	}
	{
		// For-range over struct slice.
		s := []Pair{{1, 2}, {3, 4}, {5, 6}}
		sum := 0
		for _, p := range s {
			sum += p.x + p.y
		}
		if sum != 21 {
			panic("want range struct sum==21")
		}
	}
	{
		// Slice from array: modification affects the array.
		arr := [...]int{1, 2, 3}
		s := arr[:]
		s[0] = 99
		if arr[0] != 99 {
			panic("want array modified via slice")
		}
	}
	{
		// Sub-slice: modification affects the original.
		s := []int{1, 2, 3, 4, 5}
		sub := s[1:4]
		sub[0] = 99
		if s[1] != 99 {
			panic("want original modified via sub-slice")
		}
	}
	{
		// Append after sub-slice.
		s := make([]int, 3, 6)
		s[0] = 1
		s[1] = 2
		s[2] = 3
		s = append(s, 4)
		if len(s) != 4 || s[3] != 4 {
			panic("want append after make")
		}
	}
	{
		// Make with cap, fill with append.
		s := make([]int, 0, 5)
		s = append(s, 10)
		s = append(s, 20, 30)
		s = append(s, 40, 50)
		if len(s) != 5 || cap(s) != 5 {
			panic("want filled to cap")
		}
		if s[0] != 10 || s[4] != 50 {
			panic("want filled values")
		}
	}
	{
		// Make with byte slice: zero-initialized.
		s := make([]byte, 4)
		if s[0] != 0 || s[3] != 0 {
			panic("want byte zero init")
		}
		s[0] = 0xFF
		if s[0] != 0xFF {
			panic("want byte set")
		}
	}
	{
		// Slice in if-init statement.
		s := []int{10, 20, 30}
		if v := s[1]; v == 20 {
			_ = v
		} else {
			panic("want if-init slice")
		}
	}
	{
		// Index with variable.
		s := []int{100, 200, 300}
		idx := 2
		if s[idx] != 300 {
			panic("want variable index")
		}
		s[idx] = 999
		if s[2] != 999 {
			panic("want variable index set")
		}
	}
	{
		// Index with expression.
		s := []int{100, 200, 300}
		if s[len(s)-1] != 300 {
			panic("want expr index")
		}
	}
	{
		// Len and cap in comparison.
		s := make([]int, 3, 8)
		if len(s) >= cap(s) {
			panic("want len < cap")
		}
		if cap(s) != 8 {
			panic("want cap==8")
		}
	}
	{
		// Clear slice.
		s := []int{1, 2, 3}
		clear(s)
		if s[0] != 0 || s[1] != 0 || s[2] != 0 {
			panic("want zeroed after clear")
		}
		if len(s) != 3 {
			panic("want len preserved after clear")
		}
	}
}
