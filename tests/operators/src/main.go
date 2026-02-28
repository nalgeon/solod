package main

func main() {
	{
		// Integer arithmetics.
		var a, b, c int = 11, 22, 33
		d := b/a + (a-c)*a + c%b
		d += 10
		d -= 10
		d *= 10
		d /= 2
		d %= 5
		d++
		d--
		_ = d
	}

	{
		// Floating-point arithmetics.
		var x, y, z float64 = 1.1, 2.2, 3.3
		f := x/y + (y-z)*x
		f += 1.0
		f -= 1.0
		f *= 2.0
		f /= 2.0
		f++
		f--
		_ = f
	}

	// {
	// 	// String concatenation is not supported.
	// 	var s1, s2, s3 string = "hello", " ", "world"
	// 	s4 := s1 + s2 + s3
	// 	_ = s4
	// }

	{
		// Bitwise operations.
		var b1, b2 = 0b1010, 0b1100
		b3 := ((b1 | b2) & (b1 & b2)) | (b1 ^ b2)
		b3 = b3 << 2
		b3 = b3 >> 1
		b3 = b3 &^ b1
		_ = b3
		b4 := 0b1010
		b4 |= 0b1100
		b4 &= 0b1100
		b4 ^= 0b1100
		// b4 &^= 0b1010 // not supported
		_ = b4
	}

	{
		// Logical operations.
		var a, b, c bool = true, false, true
		d := ((a && b) || (b || c)) && !a
		_ = d
	}

	{
		// Number comparison.
		x, y, z := 10, 20, 30
		e1 := ((x < y) && (y > z)) || (x == z)
		_ = e1
		e2 := ((x <= y) && (y >= z)) || (x != z)
		_ = e2
	}

	{
		// Rune comparison.
		r1, r2, r3 := 'a', 'b', '本'
		e1 := ((r1 < r2) && (r2 > r3)) || (r1 == r3)
		_ = e1
		e2 := ((r1 <= r2) && (r2 >= r3)) || (r1 != r3)
		_ = e2
	}

	{
		// String comparison.
		s1, s2, s3 := "hello", "world", "hello"
		e1 := ((s1 < s2) || (s1 > s3)) && ((s1 == s3) || (s2 != s3))
		_ = e1
		e2 := ((s1 <= s2) && (s1 >= s3)) || (s1 != s3)
		_ = e2
	}
}
