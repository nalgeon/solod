// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode_test

import (
	"fmt"

	"github.com/nalgeon/solod/so/unicode"
)

// Functions starting with "Is" can be used to inspect which table of range a
// rune belongs to. Note that runes may fit into more than one range.
func Example_is() {
	// constant with mixed type runes
	const mixed = "\b5Ὂg̀9! ℃ᾭG"
	for _, c := range mixed {
		fmt.Printf("For %q:\n", c)
		if unicode.IsControl(c) {
			fmt.Println("\tis control rune")
		}
		if unicode.IsDigit(c) {
			fmt.Println("\tis digit rune")
		}
		if unicode.IsLetter(c) {
			fmt.Println("\tis letter rune")
		}
		if unicode.IsLower(c) {
			fmt.Println("\tis lower case rune")
		}
		if unicode.IsSpace(c) {
			fmt.Println("\tis space rune")
		}
		if unicode.IsTitle(c) {
			fmt.Println("\tis title case rune")
		}
		if unicode.IsUpper(c) {
			fmt.Println("\tis upper case rune")
		}
	}

	// Output:
	// For '\b':
	// 	is control rune
	// For '5':
	// 	is digit rune
	// For 'Ὂ':
	// 	is letter rune
	// 	is upper case rune
	// For 'g':
	// 	is letter rune
	// 	is lower case rune
	// For '̀':
	// For '9':
	// 	is digit rune
	// For '!':
	// For ' ':
	// 	is space rune
	// For '℃':
	// For 'ᾭ':
	// 	is letter rune
	// 	is title case rune
	// For 'G':
	// 	is letter rune
	// 	is upper case rune
}

func ExampleTo() {
	const lcG = 'g'
	fmt.Printf("%#U\n", unicode.To(unicode.UpperCase, lcG))
	fmt.Printf("%#U\n", unicode.To(unicode.LowerCase, lcG))
	fmt.Printf("%#U\n", unicode.To(unicode.TitleCase, lcG))

	const ucG = 'G'
	fmt.Printf("%#U\n", unicode.To(unicode.UpperCase, ucG))
	fmt.Printf("%#U\n", unicode.To(unicode.LowerCase, ucG))
	fmt.Printf("%#U\n", unicode.To(unicode.TitleCase, ucG))

	// Output:
	// U+0047 'G'
	// U+0067 'g'
	// U+0047 'G'
	// U+0047 'G'
	// U+0067 'g'
	// U+0047 'G'
}

func ExampleToLower() {
	const ucG = 'G'
	fmt.Printf("%#U\n", unicode.ToLower(ucG))

	// Output:
	// U+0067 'g'
}

func ExampleToTitle() {
	const ucG = 'g'
	fmt.Printf("%#U\n", unicode.ToTitle(ucG))

	// Output:
	// U+0047 'G'
}

func ExampleToUpper() {
	const ucG = 'g'
	fmt.Printf("%#U\n", unicode.ToUpper(ucG))

	// Output:
	// U+0047 'G'
}

func ExampleIsDigit() {
	fmt.Printf("%t\n", unicode.IsDigit('৩'))
	fmt.Printf("%t\n", unicode.IsDigit('A'))
	// Output:
	// true
	// false
}

func ExampleIsLetter() {
	fmt.Printf("%t\n", unicode.IsLetter('A'))
	fmt.Printf("%t\n", unicode.IsLetter('7'))
	// Output:
	// true
	// false
}

func ExampleIsLower() {
	fmt.Printf("%t\n", unicode.IsLower('a'))
	fmt.Printf("%t\n", unicode.IsLower('A'))
	// Output:
	// true
	// false
}

func ExampleIsUpper() {
	fmt.Printf("%t\n", unicode.IsUpper('A'))
	fmt.Printf("%t\n", unicode.IsUpper('a'))
	// Output:
	// true
	// false
}

func ExampleIsTitle() {
	fmt.Printf("%t\n", unicode.IsTitle('ǅ'))
	fmt.Printf("%t\n", unicode.IsTitle('a'))
	// Output:
	// true
	// false
}

func ExampleIsSpace() {
	fmt.Printf("%t\n", unicode.IsSpace(' '))
	fmt.Printf("%t\n", unicode.IsSpace('\n'))
	fmt.Printf("%t\n", unicode.IsSpace('\t'))
	fmt.Printf("%t\n", unicode.IsSpace('a'))
	// Output:
	// true
	// true
	// true
	// false
}
