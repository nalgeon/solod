package main

func copyBuf(buf []byte) (int64, error) {
	n1, _ := copyImpl(buf)
	n2, _ := copyImpl([]byte{})
	return n1 + n2, nil
}

func copyImpl(buf []byte) (int64, error) {
	return int64(10 + len(buf)), nil
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
		n, _ := copyBuf(buf[:])
		if n != 24 {
			panic("want 24")
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
