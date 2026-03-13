package main

import "github.com/nalgeon/solod/so/unicode/utf8"

func main() {
	{
		// DecodeLastRune.
		b := []byte("Hello, 世界")
		r, size := utf8.DecodeLastRune(b)
		if r != '界' || size != 3 {
			panic("DecodeLastRune failed")
		}
	}
	{
		// DecodeLastRuneInString.
		str := "Hello, 世界"
		r, size := utf8.DecodeLastRuneInString(str)
		if r != '界' || size != 3 {
			panic("DecodeLastRuneInString failed")
		}
	}
	{
		// DecodeRune.
		b := []byte("Hello, 世界")
		r, size := utf8.DecodeRune(b)
		if r != 'H' || size != 1 {
			panic("DecodeRune failed")
		}
	}
	{
		// DecodeRuneInString.
		str := "Hello, 世界"
		r, size := utf8.DecodeRuneInString(str)
		if r != 'H' || size != 1 {
			panic("DecodeRuneInString failed")
		}
	}
	{
		// EncodeRune.
		buf := make([]byte, 3)
		n := utf8.EncodeRune(buf, '界')
		if n != 3 || string(buf) != "界" {
			panic("EncodeRune failed")
		}
	}
	{
		// RuneCount.
		n := utf8.RuneCount([]byte("Hello, 世界"))
		if n != 9 {
			panic("RuneCount failed")
		}
	}
	{
		// RuneCountInString.
		n := utf8.RuneCountInString("Hello, 世界")
		if n != 9 {
			panic("RuneCountInString failed")
		}
	}
	{
		// RuneLen.
		n := utf8.RuneLen('界')
		if n != 3 {
			panic("RuneLen failed")
		}
	}
	{
		// ValidString.
		if !utf8.ValidString("Hello, 世界") {
			panic("ValidString failed")
		}
	}
	{
		// AppendRune.
		buf := make([]byte, 7, 10)
		copy(buf, []byte("Hello, "))
		buf = utf8.AppendRune(buf, '界')
		if string(buf) != "Hello, 界" {
			panic("AppendRune failed")
		}
	}
}
