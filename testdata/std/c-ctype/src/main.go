package main

import "github.com/nalgeon/solod/so/c/ctype"

func main() {
	if !ctype.IsAlnum('a') {
		panic("want IsAlnum = true")
	}
	if !ctype.IsAlpha('a') {
		panic("want IsAlpha = true")
	}
	if !ctype.IsBlank(' ') {
		panic("want IsBlank = true")
	}
	if !ctype.IsCntrl(0x1F) {
		panic("want IsCntrl = true")
	}
	if !ctype.IsDigit('7') {
		panic("want IsDigit = true")
	}
	if !ctype.IsPunct(',') {
		panic("want IsPunct = true")
	}
	if !ctype.IsSpace('\n') {
		panic("want IsSpace = true")
	}
	if !ctype.IsUpper('A') {
		panic("want IsUpper = true")
	}
	if !ctype.IsXDigit('B') {
		panic("want IsXDigit = true")
	}
	if ctype.ToLower('A') != 'a' {
		panic("want ToLower(A) = a")
	}
	if ctype.ToUpper('a') != 'A' {
		panic("want ToUpper(a) = A")
	}
}
