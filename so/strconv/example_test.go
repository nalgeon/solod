// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv_test

import (
	"solod.dev/so/fmt"
	"solod.dev/so/strconv"
)

func ExampleAppendBool() {
	b := []byte("bool:")
	b = strconv.AppendBool(b, true)
	fmt.Println(string(b))

	// Output:
	// bool:true
}

func ExampleAppendFloat() {
	b32 := []byte("float32:")
	b32 = strconv.AppendFloat(b32, 3.1415926535, 'E', -1, 32)
	fmt.Println(string(b32))

	b64 := []byte("float64:")
	b64 = strconv.AppendFloat(b64, 3.1415926535, 'E', -1, 64)
	fmt.Println(string(b64))

	// Output:
	// float32:3.1415927E+00
	// float64:3.1415926535E+00
}

func ExampleAppendInt() {
	b10 := []byte("int (base 10):")
	b10 = strconv.AppendInt(b10, -42, 10)
	fmt.Println(string(b10))

	b16 := []byte("int (base 16):")
	b16 = strconv.AppendInt(b16, -42, 16)
	fmt.Println(string(b16))

	// Output:
	// int (base 10):-42
	// int (base 16):-2a
}

func ExampleAppendUint() {
	b10 := []byte("uint (base 10):")
	b10 = strconv.AppendUint(b10, 42, 10)
	fmt.Println(string(b10))

	b16 := []byte("uint (base 16):")
	b16 = strconv.AppendUint(b16, 42, 16)
	fmt.Println(string(b16))

	// Output:
	// uint (base 10):42
	// uint (base 16):2a
}

func ExampleAtoi() {
	v := "10"
	if s, err := strconv.Atoi(v); err == nil {
		fmt.Printf("%T, %v", s, s)
	}

	// Output:
	// int, 10
}

func ExampleFormatBool() {
	v := true
	s := strconv.FormatBool(v)
	fmt.Printf("%T, %v\n", s, s)

	// Output:
	// string, true
}

func ExampleFormatFloat() {
	v := 3.1415926535
	buf := make([]byte, 64)

	s32 := strconv.FormatFloat(buf, v, 'E', -1, 32)
	fmt.Printf("%T, %v\n", s32, s32)

	s64 := strconv.FormatFloat(buf, v, 'E', -1, 64)
	fmt.Printf("%T, %v\n", s64, s64)

	// fmt.Println uses these arguments to print floats
	fmt64 := strconv.FormatFloat(buf, v, 'g', -1, 64)
	fmt.Printf("%T, %v\n", fmt64, fmt64)

	// Output:
	// string, 3.1415927E+00
	// string, 3.1415926535E+00
	// string, 3.1415926535
}

func ExampleFormatInt() {
	v := int64(-42)
	buf := make([]byte, 64)

	s10 := strconv.FormatInt(buf, v, 10)
	fmt.Printf("%T, %v\n", s10, s10)

	s16 := strconv.FormatInt(buf, v, 16)
	fmt.Printf("%T, %v\n", s16, s16)

	// Output:
	// string, -42
	// string, -2a
}

func ExampleFormatUint() {
	v := uint64(42)
	buf := make([]byte, 64)

	s10 := strconv.FormatUint(buf, v, 10)
	fmt.Printf("%T, %v\n", s10, s10)

	s16 := strconv.FormatUint(buf, v, 16)
	fmt.Printf("%T, %v\n", s16, s16)

	// Output:
	// string, 42
	// string, 2a
}

func ExampleItoa() {
	i := 10
	buf := make([]byte, 64)
	s := strconv.Itoa(buf, i)
	fmt.Printf("%T, %v\n", s, s)

	// Output:
	// string, 10
}

func ExampleParseBool() {
	v := "true"
	if s, err := strconv.ParseBool(v); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}

	// Output:
	// bool, true
}

func ExampleParseFloat() {
	v := "3.1415926535"
	if s, err := strconv.ParseFloat(v, 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat(v, 64); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat("NaN", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	// ParseFloat is case insensitive
	if s, err := strconv.ParseFloat("nan", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat("inf", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat("+Inf", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat("-Inf", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat("-0", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseFloat("+0", 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}

	// Output:
	// float64, 3.1415927410125732
	// float64, 3.1415926535
	// float64, NaN
	// float64, NaN
	// float64, +Inf
	// float64, +Inf
	// float64, -Inf
	// float64, -0
	// float64, 0
}

func ExampleParseInt() {
	v32 := "-354634382"
	if s, err := strconv.ParseInt(v32, 10, 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseInt(v32, 16, 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}

	v64 := "-3546343826724305832"
	if s, err := strconv.ParseInt(v64, 10, 64); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseInt(v64, 16, 64); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}

	// Output:
	// int64, -354634382
	// int64, -3546343826724305832
}

func ExampleParseUint() {
	v := "42"
	if s, err := strconv.ParseUint(v, 10, 32); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	if s, err := strconv.ParseUint(v, 10, 64); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}

	// Output:
	// uint64, 42
	// uint64, 42
}
