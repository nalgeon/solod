// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"solod.dev/so/flag"
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

func TestSet(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	names := []string{"bool", "int", "int64", "uint", "uint64", "float64", "string"}
	values := []string{"true", "11", "22", "33", "44", "1.5", "hello"}
	for i, name := range names {
		err := fs.Set(name, values[i])
		if err != nil {
			t.Errorf("Set(%s, %s) = %s, want nil", name, values[i], errName(err))
		}
	}

	if !v.b {
		t.Error("bool = false, want true")
	}
	if v.i != 11 {
		t.Errorf("int = %d, want 11", v.i)
	}
	if v.i64 != 22 {
		t.Errorf("int64 = %d, want 22", int(v.i64))
	}
	if v.u != 33 {
		t.Errorf("uint = %d, want 33", int(v.u))
	}
	if v.u64 != 44 {
		t.Errorf("uint64 = %d, want 44", int(v.u64))
	}
	if v.f != 1.5 {
		t.Errorf("float64 = %g, want 1.5", v.f)
	}
	if v.s != "hello" {
		t.Errorf("string = %s, want hello", v.s)
	}
}

func TestSetTwice(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	// The last Set wins.
	fs.Set("int", "1")
	fs.Set("int", "2")
	if v.i != 2 {
		t.Errorf("int = %d, want 2", v.i)
	}
}

func TestSetBoolForms(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var b bool
	fs.BoolVar(&b, "bool", false, "bool value")

	trues := []string{"1", "t", "T", "TRUE", "true", "True"}
	for _, s := range trues {
		b = false
		err := fs.Set("bool", s)
		if err != nil {
			t.Errorf("Set(bool, %s) = %s, want nil", s, errName(err))
		}
		if !b {
			t.Errorf("bool = false after Set(bool, %s), want true", s)
		}
	}

	falses := []string{"0", "f", "F", "FALSE", "false", "False"}
	for _, s := range falses {
		b = true
		err := fs.Set("bool", s)
		if err != nil {
			t.Errorf("Set(bool, %s) = %s, want nil", s, errName(err))
		}
		if b {
			t.Errorf("bool = true after Set(bool, %s), want false", s)
		}
	}
}

func TestSetIntBases(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var i int
	fs.IntVar(&i, "int", 0, "int value")

	// The prefix of the value picks the base.
	values := []string{"1234", "0664", "0o664", "0x1234", "0b101", "-5", "+7", "1_000"}
	want := []int{1234, 436, 436, 4660, 5, -5, 7, 1000}
	for k, s := range values {
		err := fs.Set("int", s)
		if err != nil {
			t.Errorf("Set(int, %s) = %s, want nil", s, errName(err))
			continue
		}
		if i != want[k] {
			t.Errorf("Set(int, %s): int = %d, want %d", s, i, want[k])
		}
	}
}

func TestSetUintBases(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var u uint
	fs.UintVar(&u, "uint", 0, "uint value")

	values := []string{"1234", "0664", "0x1234", "0b101"}
	want := []int{1234, 436, 4660, 5}
	for k, s := range values {
		err := fs.Set("uint", s)
		if err != nil {
			t.Errorf("Set(uint, %s) = %s, want nil", s, errName(err))
			continue
		}
		if int(u) != want[k] {
			t.Errorf("Set(uint, %s): uint = %d, want %d", s, int(u), want[k])
		}
	}

	// A uint has no sign.
	err := fs.Set("uint", "-1")
	if err != flag.ErrParse {
		t.Errorf("Set(uint, -1) = %s, want ErrParse", errName(err))
	}
}

func TestSetFloatForms(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var f float64
	fs.Float64Var(&f, "float64", 0, "float64 value")

	values := []string{"1", "-1.5", "2718e28", "1e-3", ".5"}
	want := []float64{1, -1.5, 2718e28, 1e-3, 0.5}
	for k, s := range values {
		err := fs.Set("float64", s)
		if err != nil {
			t.Errorf("Set(float64, %s) = %s, want nil", s, errName(err))
			continue
		}
		if f != want[k] {
			t.Errorf("Set(float64, %s): float64 = %g, want %g", s, f, want[k])
		}
	}
}

func TestSetString(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var s string
	fs.StringVar(&s, "string", "0", "string value")

	// A string flag accepts every value.
	values := []string{"", "x", "-1", "a b c", "="}
	for _, val := range values {
		err := fs.Set("string", val)
		if err != nil {
			t.Errorf("Set(string, %s) = %s, want nil", val, errName(err))
		}
		if s != val {
			t.Errorf("string = %s, want %s", s, val)
		}
	}
}

func TestSetNotFound(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	err := fs.Set("nosuchflag", "1")
	if err != flag.ErrNotFound {
		t.Errorf("Set(nosuchflag, 1) = %s, want ErrNotFound", errName(err))
	}
}

func TestSetParseError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	// A string flag cannot fail, so it is not in the list.
	names := []string{"bool", "int", "int64", "uint", "uint64", "float64"}
	for _, name := range names {
		err := fs.Set(name, "x")
		if err != flag.ErrParse {
			t.Errorf("Set(%s, x) = %s, want ErrParse", name, errName(err))
		}
		err = fs.Set(name, "")
		if err != flag.ErrParse {
			t.Errorf("Set(%s, ) = %s, want ErrParse", name, errName(err))
		}
	}
}

func TestSetRangeError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	names := []string{"int", "int64", "uint", "uint64", "float64"}
	values := []string{
		"123456789012345678901",
		"123456789012345678901",
		"123456789012345678901",
		"123456789012345678901",
		"1e1000",
	}
	for i, name := range names {
		err := fs.Set(name, values[i])
		if err != flag.ErrRange {
			t.Errorf("Set(%s, %s) = %s, want ErrRange", name, values[i], errName(err))
		}
	}
}

func TestSetIntWidth(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var i int
	fs.IntVar(&i, "int", 0, "int value")
	var u uint
	fs.UintVar(&u, "uint", 0, "uint value")

	// An int flag holds the width of the int of the target, not 64 bits.
	iErr := fs.Set("int", "2147483648")
	uErr := fs.Set("uint", "4294967296")
	if strconv.IntSize == 32 {
		if iErr != flag.ErrRange {
			t.Errorf("Set(int, 2147483648) = %s, want ErrRange", errName(iErr))
		}
		if uErr != flag.ErrRange {
			t.Errorf("Set(uint, 4294967296) = %s, want ErrRange", errName(uErr))
		}
		return
	}
	if iErr != nil {
		t.Errorf("Set(int, 2147483648) = %s, want nil", errName(iErr))
	}
	if int64(i) != int32Overflow {
		t.Errorf("int = %d, want 2147483648", i)
	}
	if uErr != nil {
		t.Errorf("Set(uint, 4294967296) = %s, want nil", errName(uErr))
	}
	if uint64(u) != uint32Overflow {
		t.Errorf("uint = %d, want 4294967296", int(u))
	}
}

func TestSetUserDefined(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var lst listValue
	fs.Var(&lst, "list", "list value")

	// Set goes to the Value, in call order.
	fs.Set("list", "a")
	fs.Set("list", "b")
	if lst.n != 2 {
		t.Fatalf("list holds %d values, want 2", lst.n)
		return
	}
	if lst.items[0] != "a" || lst.items[1] != "b" {
		t.Errorf("list = [%s %s], want [a b]", lst.items[0], lst.items[1])
	}
}

func TestSetUserDefinedError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var fail failValue
	fs.Var(&fail, "fail", "fail value")

	// Set returns the error of the Value as it is.
	err := fs.Set("fail", "x")
	if err != errFail {
		t.Errorf("Set(fail, x) = %s, want errFail", errName(err))
	}
	if fail.calls != 1 {
		t.Errorf("Set called %d times, want 1", fail.calls)
	}
}
