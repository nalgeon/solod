package main

func main() {
	str := "Hi 世界!"

	// Loop over bytes.
	for i := 0; i < len(str); i++ {
		chr := str[i]
		println("i =", i, "chr =", chr)
	}

	// Loop over runes.
	for i, r := range str {
		println("i =", i, "r =", r)
	}
	for i := range str {
		println("i =", i)
	}
	for _, r := range str {
		println("r =", r)
	}

	{
		// Compare strings.
		s1 := "hello"
		s2 := "world"
		if s1 == s2 || s1 == "hello" {
			println("ok")
		}
	}

	{
		// String conversion.
		s := "1世3"
		bs := []byte(s)
		if bs[0] != '1' {
			panic("unexpected byte")
		}
		rs := []rune(s)
		if rs[1] != '世' {
			panic("unexpected rune")
		}
	}

}
