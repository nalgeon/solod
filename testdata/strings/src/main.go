package main

// Named byte and rune types.
type Byte byte
type Rune rune

// Emitted as adjacent string literals.
const quoted = string('"') + "ok" + string('"')

func main() {
	{
		// Empty string.
		s1 := ""
		if len(s1) != 0 || s1 != "" {
			panic("want empty string")
		}
		var s2 string
		if len(s2) != 0 || s2 != "" {
			panic("want empty string")
		}
	}
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
		var r rune
		for _, r = range str {
			_ = r
		}
		for i, r := range "go" {
			println("i =", i, "r =", r)
		}
		for range str {
		}
	}
	{
		// Loop over string runes with a declared key. The loop assigns to the
		// key, so the key keeps the index of the last rune.
		s := "go世"
		var i int
		for i = range s {
			_ = i
		}
		if i != 2 {
			panic("want i == 2")
		}
		var r rune
		for i, r = range s {
			_ = r
		}
		if i != 2 || r != '世' {
			panic("want i == 2 && r == 世")
		}
	}
	{
		// Continue in range-over-string loop.
		s := "hello"
		n := 0
		for _, c := range s {
			if c == 'l' {
				continue
			}
			n++
		}
		if n != 3 {
			panic("want n == 3")
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
	{
		// String addition.
		s1 := "Hello, "
		s2 := "世界!"
		s3 := s1 + s2
		if s3 != "Hello, 世界!" {
			panic("want s3 == Hello, 世界!")
		}
	}
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
		var b byte = 'A'
		if string(b) != "A" {
			panic("want string(b) == A")
		}
		var r rune = '世'
		if string(r) != "世" {
			panic("want string(r) == 世")
		}
		var b2 byte = 200
		if string(b2) != "È" {
			panic("want string(b2) == È")
		}
		var n int = 0x4e16
		if string(n) != "世" {
			panic("want string(n) == 世")
		}
		var u uint64 = 0x10001f4a9
		if string(u) != "\uFFFD" {
			panic("want string(u) == replacement char")
		}
		const c = string(1 << 100)
		if c != "\uFFFD" {
			panic("want c == replacement char")
		}
		if quoted != `"ok"` {
			panic("want quoted == \"ok\"")
		}
	}
	{
		// String conversion to slices of named byte and rune types.
		s1 := "1世3"
		bs := []Byte(s1)
		if bs[0] != '1' {
			panic("unexpected byte")
		}
		rs := []Rune(s1)
		if rs[1] != '世' {
			panic("unexpected rune")
		}
		if string(bs) != s1 {
			panic("want string(bs) == s1")
		}
		if string(rs) != s1 {
			panic("want string(rs) == s1")
		}
	}
}
