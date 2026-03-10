package main

func main() {
	{
		// String literals.
		s := "Hello, 世界!"
		if len(s) != 7+3+3+1 {
			panic("want len(s) == 14")
		}
	}

	{
		// Loop over string bytes.
		str := "Hi 世界!"

		for i := 0; i < len(str); i++ {
			chr := str[i]
			println("i =", i, "chr =", chr)
		}
	}

	{
		// Loop over string runes.
		str := "Hi 世界!"
		for i, r := range str {
			println("i =", i, "r =", r)
		}
		for i := range str {
			println("i =", i)
		}
		for _, r := range str {
			println("r =", r)
		}
		for i, r := range "go" {
			println("i =", i, "r =", r)
		}
		for range str {
		}
	}

	{
		// Compare strings.
		s1 := "hello"
		s2 := "world"
		if s1 == s2 || s1 == "hello" {
			println("ok")
		}
	}

	// {
	// 	// String addition is not supported.
	// 	s1 := "Hello, "
	// 	s2 := "世界!"
	// 	s3 := s1 + s2
	// 	if s3 != "Hello, 世界!" {
	// 		panic("want s3 == Hello, 世界!")
	// 	}
	// }

	{
		// String conversion to byte and rune slices, and vice versa.
		s1 := "1世3"
		bs := []byte(s1)
		if bs[0] != '1' {
			panic("unexpected byte")
		}
		rs := []rune(s1)
		if rs[1] != '世' {
			panic("unexpected rune")
		}
		s2 := string(bs)
		if s2 != s1 {
			panic("want s2 == s1")
		}
		s3 := string(rs)
		if s3 != s1 {
			panic("want s3 == s1")
		}
	}
}
