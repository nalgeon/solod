package main

type counter struct {
	val *int
}

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
		// Division by -1 overflows for the minimum value.
		var i32 int32 = -2147483648
		var neg int32 = -1
		if i32/neg != -2147483648 || i32%neg != 0 {
			panic("expected i32 == -2147483648 && rem == 0")
		}
		if i32/-1 != -2147483648 || i32%-1 != 0 {
			panic("expected the same for a constant divisor")
		}
		var i64 int64 = -9223372036854775808
		if i64/-1 != -9223372036854775808 || i64%-1 != 0 {
			panic("expected i64 == -9223372036854775808 && rem == 0")
		}
		var i8 int8 = -128
		if i8/-1 != -128 || i8%-1 != 0 {
			panic("expected i8 == -128 && rem == 0")
		}
		// An unsigned divisor never reaches the guard, because -1
		// converts to the maximum value of the type.
		var u1, u2 uint32 = 7, 4294967295
		if u1/u2 != 0 || u1%u2 != 7 {
			panic("expected u1/u2 == 0 && u1%u2 == 7")
		}
		q := int32(-2147483648)
		q /= -1
		if q != -2147483648 {
			panic("expected q == -2147483648")
		}
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

	{
		// String addition is supported for string literals (but not for variables).
		s := "hello" + " " + "world"
		_ = s
	}

	{
		// Bitwise operations.
		var b1, b2 = 0b1010, 0b1100
		b3 := ((b1 | b2) & (b1 & b2)) | (b1 ^ b2)
		b3 = b3 << 2
		b3 = b3 >> 1
		b3 <<= 2
		b3 >>= 1
		b3 = b3 &^ b1
		_ = b3
		b4 := 0b1010
		b4 |= 0b1100
		b4 &= 0b1100
		b4 ^= 0b1100
		// b4 &^= 0b1010 // not supported
		b5 := ^b4
		_ = b5
	}

	{
		// Arithmetic on a type narrower than int. C promotes the operands to
		// int and computes at that width, so every result needs a conversion
		// back to the narrow type.
		var n1, n2 byte = 3, 10
		n3 := n1 - n2
		_ = n3
		n4 := int(n1 - n2)
		_ = n4
		n5 := int(n1 * n2)
		_ = n5
		n6 := int(n1 << 6)
		_ = n6
		n7 := int(^n1)
		_ = n7
		n8 := int(-n1)
		_ = n8
		var s1, s2 int16 = 30000, 30000
		n9 := int(s1 + s2)
		_ = n9
	}

	{
		// Increment/decrement through a pointer.
		n := 1
		p := &n
		*p++
		*p--
		c := counter{val: &n}
		*c.val++
		*c.val--
		_ = n
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
		// Byte comparison.
		var b1, b2, b3 byte = 'a', 'b', 'c'
		e1 := ((b1 < b2) && (b2 > b3)) || (b1 == b3)
		_ = e1
		e2 := ((b1 <= b2) && (b2 >= b3)) || (b1 != b3)
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
