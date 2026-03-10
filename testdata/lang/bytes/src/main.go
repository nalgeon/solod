package main

func main() {
	{
		// Byte literals.
		var b1, b2, b3 byte = 'a', 'b', 'c'
		if b1 != 'a' || b2 != 'b' || b3 != 'c' {
			panic("unexpected byte")
		}
	}
	{
		// Rune literals.
		var r1, r2, r3 rune = '世', '界', '!'
		if r1 != '世' || r2 != '界' || r3 != '!' {
			panic("unexpected rune")
		}
	}
	{
		// Byte slices and strings.
		b := []byte{'h', 'e', 'l', 'l', 'o'}
		s := string(b)
		if s != "hello" {
			panic("want s == hello")
		}
	}
	{
		// Rune slices and strings.
		r := []rune{'世', '界'}
		s := string(r)
		if s != "世界" {
			panic("want s == 世界")
		}
	}
}
