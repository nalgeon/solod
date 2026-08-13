// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt_test

import (
	"solod.dev/so/fmt"
	"solod.dev/so/strings"
)

func ExamplePrint() {
	const name, age = "Kim", "22"
	fmt.Print(name, " is ", age, " years old.\n")

	// It is conventional not to worry about any
	// error returned by Print.

	// Output:
	// Kim is 22 years old.
}

func ExamplePrintln() {
	const name, age = "Kim", "22"
	fmt.Println(name, "is", age, "years old.")

	// It is conventional not to worry about any
	// error returned by Println.

	// Output:
	// Kim is 22 years old.
}

func ExamplePrintf() {
	const name, age = "Kim", 22
	fmt.Printf("%s is %d years old.\n", name, age)

	// It is conventional not to worry about any
	// error returned by Printf.

	// Output:
	// Kim is 22 years old.
}

func ExampleSprintf() {
	const name, age = "Kim", 22
	buf := make([]byte, 64)
	s := fmt.Sprintf(buf, "%s is %d years old.\n", name, age)
	fmt.Print(s)

	// Output:
	// Kim is 22 years old.
}

func ExampleFprintf() {
	const name, age = "Kim", 22
	var sb strings.Builder
	n, err := fmt.Fprintf(&sb, "%s is %d years old.\n", name, age)

	// The n and err return values from Fprintf are
	// those returned by the underlying io.Writer.
	if err != nil {
		panic(err)
	}
	fmt.Print(sb.String())
	fmt.Printf("%d bytes written.\n", n)
	sb.Free()

	// Output:
	// Kim is 22 years old.
	// 21 bytes written.
}

func ExampleSscanf() {
	var name string
	var age int
	n, err := fmt.Sscanf("Kim is 22 years old", "%s is %d years old", &name, &age)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d: %s, %d\n", n, name, age)

	// Output:
	// 2: Kim, 22
}
