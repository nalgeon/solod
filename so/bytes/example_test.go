package bytes_test

import (
	"os"

	"solod.dev/so/bytes"
	"solod.dev/so/fmt"
	"solod.dev/so/unicode"
)

func ExampleBuffer() {
	var b bytes.Buffer // A Buffer needs no initialization.
	b.Write([]byte("Hello "))
	fmt.Fprintf(&b, "world!")
	b.WriteTo(os.Stdout)
	// Output: Hello world!
}

func ExampleBuffer_Bytes() {
	buf := bytes.Buffer{}
	buf.Write([]byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'})
	os.Stdout.Write(buf.Bytes())
	// Output: hello world
}

func ExampleBuffer_Cap() {
	buf1 := bytes.NewBuffer(nil, make([]byte, 10))
	buf2 := bytes.NewBuffer(nil, make([]byte, 0, 10))
	fmt.Printf("%d\n", buf1.Cap())
	fmt.Printf("%d\n", buf2.Cap())
	// Output:
	// 10
	// 10
}

func ExampleBuffer_Grow() {
	var b bytes.Buffer
	b.Grow(64)
	bb := b.Bytes()
	b.Write([]byte("64 bytes or fewer"))
	fmt.Printf("%q", bb[:b.Len()])
	// Output: "64 bytes or fewer"
}

func ExampleBuffer_Len() {
	var b bytes.Buffer
	b.Grow(64)
	b.Write([]byte("abcde"))
	fmt.Printf("%d", b.Len())
	// Output: 5
}

func ExampleBuffer_Next() {
	var b bytes.Buffer
	b.Grow(64)
	b.Write([]byte("abcde"))
	fmt.Printf("%s\n", b.Next(2))
	fmt.Printf("%s\n", b.Next(2))
	fmt.Printf("%s", b.Next(2))
	// Output:
	// ab
	// cd
	// e
}

func ExampleBuffer_Read() {
	var b bytes.Buffer
	b.Grow(64)
	b.Write([]byte("abcde"))
	rdbuf := make([]byte, 1)
	n, err := b.Read(rdbuf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d\n", n)
	fmt.Println(b.String())
	fmt.Println(string(rdbuf))
	// Output:
	// 1
	// bcde
	// a
}

func ExampleBuffer_ReadByte() {
	var b bytes.Buffer
	b.Grow(64)
	b.Write([]byte("abcde"))
	c, err := b.ReadByte()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d\n", c)
	fmt.Println(b.String())
	// Output:
	// 97
	// bcde
}

func ExampleClone() {
	b := []byte("abc")
	clone := bytes.Clone(nil, b)
	fmt.Printf("%s\n", clone)
	clone[0] = 'd'
	fmt.Printf("%s\n", b)
	fmt.Printf("%s\n", clone)
	// Output:
	// abc
	// abc
	// dbc
}

func ExampleCompare() {
	// Interpret Compare's result by comparing it to zero.
	var a, b []byte
	if bytes.Compare(a, b) < 0 {
		// a less b
	}
	if bytes.Compare(a, b) <= 0 {
		// a less or equal b
	}
	if bytes.Compare(a, b) > 0 {
		// a greater b
	}
	if bytes.Compare(a, b) >= 0 {
		// a greater or equal b
	}

	// Prefer Equal to Compare for equality comparisons.
	if bytes.Equal(a, b) {
		// a equal b
	}
	if !bytes.Equal(a, b) {
		// a not equal b
	}
}

func ExampleContains() {
	fmt.Printf("%t\n", bytes.Contains([]byte("seafood"), []byte("foo")))
	fmt.Printf("%t\n", bytes.Contains([]byte("seafood"), []byte("bar")))
	fmt.Printf("%t\n", bytes.Contains([]byte("seafood"), []byte("")))
	fmt.Printf("%t\n", bytes.Contains([]byte(""), []byte("")))
	// Output:
	// true
	// false
	// true
	// true
}

func ExampleCount() {
	fmt.Printf("%d\n", bytes.Count([]byte("cheese"), []byte("e")))
	fmt.Printf("%d\n", bytes.Count([]byte("five"), []byte(""))) // before & after each rune
	// Output:
	// 3
	// 5
}

func ExampleCut() {
	show := func(s, sep string) {
		res := bytes.Cut([]byte(s), []byte(sep))
		fmt.Printf("Cut(%q, %q) = %q, %q, %v\n", s, sep, res.Before, res.After, res.Found)
	}
	show("Gopher", "Go")
	show("Gopher", "ph")
	show("Gopher", "er")
	show("Gopher", "Badger")
	// Output:
	// Cut("Gopher", "Go") = "", "pher", true
	// Cut("Gopher", "ph") = "Go", "er", true
	// Cut("Gopher", "er") = "Goph", "", true
	// Cut("Gopher", "Badger") = "Gopher", "", false
}

func ExampleEqual() {
	fmt.Printf("%t\n", bytes.Equal([]byte("Go"), []byte("Go")))
	fmt.Printf("%t\n", bytes.Equal([]byte("Go"), []byte("C++")))
	// Output:
	// true
	// false
}

func ExampleHasPrefix() {
	fmt.Printf("%t\n", bytes.HasPrefix([]byte("Gopher"), []byte("Go")))
	fmt.Printf("%t\n", bytes.HasPrefix([]byte("Gopher"), []byte("C")))
	fmt.Printf("%t\n", bytes.HasPrefix([]byte("Gopher"), []byte("")))
	// Output:
	// true
	// false
	// true
}

func ExampleHasSuffix() {
	fmt.Printf("%t\n", bytes.HasSuffix([]byte("Amigo"), []byte("go")))
	fmt.Printf("%t\n", bytes.HasSuffix([]byte("Amigo"), []byte("O")))
	fmt.Printf("%t\n", bytes.HasSuffix([]byte("Amigo"), []byte("Ami")))
	fmt.Printf("%t\n", bytes.HasSuffix([]byte("Amigo"), []byte("")))
	// Output:
	// true
	// false
	// false
	// true
}

func ExampleIndex() {
	fmt.Printf("%d\n", bytes.Index([]byte("chicken"), []byte("ken")))
	fmt.Printf("%d\n", bytes.Index([]byte("chicken"), []byte("dmr")))
	// Output:
	// 4
	// -1
}

func ExampleIndexByte() {
	fmt.Printf("%d\n", bytes.IndexByte([]byte("chicken"), byte('k')))
	fmt.Printf("%d\n", bytes.IndexByte([]byte("chicken"), byte('g')))
	// Output:
	// 4
	// -1
}

func ExampleJoin() {
	s := [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	fmt.Printf("%s", bytes.Join(nil, s, []byte(", ")))
	// Output: foo, bar, baz
}

func ExampleMap() {
	rot13 := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		}
		return r
	}
	fmt.Printf("%s\n", bytes.Map(nil, rot13, []byte("'Twas brillig and the slithy gopher...")))
	// Output:
	// 'Gjnf oevyyvt naq gur fyvgul tbcure...
}

func ExampleReader_Len() {
	r1 := bytes.NewReader([]byte("Hi!"))
	fmt.Printf("%d\n", r1.Len())
	r2 := bytes.NewReader([]byte("こんにちは!"))
	fmt.Printf("%d\n", r2.Len())
	// Output:
	// 3
	// 16
}

func ExampleReplace() {
	old := []byte("oink oink oink")
	fmt.Printf("%s\n", bytes.Replace(nil, old, []byte("k"), []byte("ky"), 2))
	fmt.Printf("%s\n", bytes.Replace(nil, old, []byte("oink"), []byte("moo"), -1))
	// Output:
	// oinky oinky oink
	// moo moo moo
}

func ExampleRunes() {
	rs := bytes.Runes(nil, []byte("go gopher"))
	for _, r := range rs {
		fmt.Printf("%#U\n", r)
	}
	// Output:
	// U+0067 'g'
	// U+006F 'o'
	// U+0020 ' '
	// U+0067 'g'
	// U+006F 'o'
	// U+0070 'p'
	// U+0068 'h'
	// U+0065 'e'
	// U+0072 'r'
}

func ExampleSplit() {
	fmt.Printf("%q\n", bytes.Split(nil, []byte("a,b,c"), []byte(",")))
	fmt.Printf("%q\n", bytes.Split(nil, []byte("a man a plan a canal panama"), []byte("a ")))
	fmt.Printf("%q\n", bytes.Split(nil, []byte(" xyz "), []byte("")))
	fmt.Printf("%q\n", bytes.Split(nil, []byte(""), []byte("Bernardo O'Higgins")))
	// Output:
	// ["a" "b" "c"]
	// ["" "man " "plan " "canal panama"]
	// [" " "x" "y" "z" " "]
	// [""]
}

func ExampleSplitN() {
	fmt.Printf("%q\n", bytes.SplitN(nil, []byte("a,b,c"), []byte(","), 2))
	z := bytes.SplitN(nil, []byte("a,b,c"), []byte(","), 0)
	fmt.Printf("%q (nil = %v)\n", z, z == nil)
	// Output:
	// ["a" "b,c"]
	// [] (nil = false)
}

func ExampleTrim() {
	fmt.Printf("[%q]", bytes.Trim([]byte(" !!! Achtung! Achtung! !!! "), "! "))
	// Output: ["Achtung! Achtung"]
}

func ExampleTrimFunc() {
	fmt.Println(string(bytes.TrimFunc([]byte("go-gopher!"), unicode.IsLetter)))
	fmt.Println(string(bytes.TrimFunc([]byte("\"go-gopher!\""), unicode.IsLetter)))
	fmt.Println(string(bytes.TrimFunc([]byte("1234go-gopher!567"), unicode.IsDigit)))
	// Output:
	// -gopher!
	// "go-gopher!"
	// go-gopher!
}

func ExampleTrimLeft() {
	fmt.Print(string(bytes.TrimLeft([]byte("453gopher8257"), "0123456789")))
	// Output:
	// gopher8257
}

func ExampleTrimPrefix() {
	var b = []byte("Goodbye,, world!")
	b = bytes.TrimPrefix(b, []byte("Goodbye,"))
	b = bytes.TrimPrefix(b, []byte("See ya,"))
	fmt.Printf("Hello%s", b)
	// Output: Hello, world!
}

func ExampleTrimSpace() {
	fmt.Printf("%s", bytes.TrimSpace([]byte(" \t\n a lone gopher \n\t\r\n")))
	// Output: a lone gopher
}

func ExampleTrimSuffix() {
	var b = []byte("Hello, goodbye, etc!")
	b = bytes.TrimSuffix(b, []byte("goodbye, etc!"))
	b = bytes.TrimSuffix(b, []byte("gopher"))
	b = append(b, bytes.TrimSuffix([]byte("world!"), []byte("x!"))...)
	os.Stdout.Write(b)
	// Output: Hello, world!
}

func ExampleTrimRight() {
	fmt.Print(string(bytes.TrimRight([]byte("453gopher8257"), "0123456789")))
	// Output:
	// 453gopher
}

func ExampleToLower() {
	fmt.Printf("%s", bytes.ToLower(nil, []byte("Gopher")))
	// Output: gopher
}

func ExampleToUpper() {
	fmt.Printf("%s", bytes.ToUpper(nil, []byte("Gopher")))
	// Output: GOPHER
}
