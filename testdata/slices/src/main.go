package main

func lenInt64(buf []byte) (int64, error) {
	n, _ := lenInt64Impl(buf)
	return n, nil
}

func lenInt64Impl(buf []byte) (int64, error) {
	return int64(len(buf)), nil
}

func main() {
	{
		// Slicing an array.
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
		// Slicing a slice.
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
		// Slice literals.
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
		// Make a slice.
		s := make([]int, 4)
		s[0] = 1
		s[1] = 2
		s[2] = 3
		s[3] = 4
		_ = s
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
		// s[1] &^= 0b1010  // not supported
		if s[1] != 9 {
			panic("want 9")
		}
	}
}
