// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"solod.dev/so/flag"
	"solod.dev/so/io"
	"solod.dev/so/testing"
)

func TestParse(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if fs.Parsed() {
		t.Error("Parsed() = true before Parse")
	}

	var v vals
	defineAll(&fs, &v)
	var b2 bool
	fs.BoolVar(&b2, "bool2", false, "bool2 value")

	extra := "one-extra-argument"
	args := []string{
		"-bool",
		"-bool2=true",
		"--int", "22",
		"--int64", "0x23",
		"-uint", "24",
		"--uint64", "25",
		"-string", "hello",
		"-float64", "2718e28",
		extra,
	}
	err := fs.Parse(args)
	if err != nil {
		t.Fatalf("Parse() = %s, want nil", errName(err))
		return
	}
	if !fs.Parsed() {
		t.Error("Parsed() = false after Parse")
	}

	if !v.b {
		t.Error("bool = false, want true")
	}
	if !b2 {
		t.Error("bool2 = false, want true")
	}
	if v.i != 22 {
		t.Errorf("int = %d, want 22", v.i)
	}
	if v.i64 != 0x23 {
		t.Errorf("int64 = %d, want 35", int(v.i64))
	}
	if v.u != 24 {
		t.Errorf("uint = %d, want 24", int(v.u))
	}
	if v.u64 != 25 {
		t.Errorf("uint64 = %d, want 25", int(v.u64))
	}
	if v.f != 2718e28 {
		t.Errorf("float64 = %g, want 2718e28", v.f)
	}
	if v.s != "hello" {
		t.Errorf("string = %s, want hello", v.s)
	}

	if fs.NArg() != 1 {
		t.Fatalf("NArg() = %d, want 1", fs.NArg())
		return
	}
	if fs.Arg(0) != extra {
		t.Errorf("Arg(0) = %s, want %s", fs.Arg(0), extra)
	}
	if len(fs.Args()) != 1 || fs.Args()[0] != extra {
		t.Errorf("Args() = [%s], want [%s]", fs.Args()[0], extra)
	}
}

func TestParseDashForms(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var s string
	fs.StringVar(&s, "string", "", "string value")

	// One dash and two dashes are equal, and the value is either the rest of
	// the argument after an equals sign or the next argument.
	one := []string{"-string", "a"}
	two := []string{"--string", "b"}
	oneEq := []string{"-string=c"}
	twoEq := []string{"--string=d"}
	want := []string{"a", "b", "c", "d"}

	fs.Parse(one)
	if s != want[0] {
		t.Errorf("-string a: string = %s, want a", s)
	}
	fs.Parse(two)
	if s != want[1] {
		t.Errorf("--string b: string = %s, want b", s)
	}
	fs.Parse(oneEq)
	if s != want[2] {
		t.Errorf("-string=c: string = %s, want c", s)
	}
	fs.Parse(twoEq)
	if s != want[3] {
		t.Errorf("--string=d: string = %s, want d", s)
	}
}

func TestParseBool(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var b bool
	fs.BoolVar(&b, "bool", false, "bool value")

	// A bool flag needs no argument.
	fs.Parse([]string{"-bool"})
	if !b {
		t.Error("-bool: bool = false, want true")
	}

	// An equals sign gives it one.
	fs.Parse([]string{"-bool=false"})
	if b {
		t.Error("-bool=false: bool = true, want false")
	}
	fs.Parse([]string{"-bool=1"})
	if !b {
		t.Error("-bool=1: bool = false, want true")
	}

	// A bool flag never takes the next argument.
	b = false
	err := fs.Parse([]string{"-bool", "false"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if !b {
		t.Error("-bool false: bool = false, want true")
	}
	if fs.NArg() != 1 || fs.Arg(0) != "false" {
		t.Errorf("-bool false: NArg() = %d, Arg(0) = %s, want 1 and false", fs.NArg(), fs.Arg(0))
	}
}

func TestParseTerminator(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	// A double dash stops the flags, and the arguments after it stay.
	err := fs.Parse([]string{"-bool", "--", "-int", "5"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if !v.b {
		t.Error("bool = false, want true")
	}
	if v.i != 0 {
		t.Errorf("int = %d, want 0", v.i)
	}
	if fs.NArg() != 2 || fs.Arg(0) != "-int" || fs.Arg(1) != "5" {
		t.Errorf("Args() = [%s %s], want [-int 5]", fs.Arg(0), fs.Arg(1))
	}

	// The terminator itself is not an argument.
	fs.Parse([]string{"--"})
	if fs.NArg() != 0 {
		t.Errorf("NArg() = %d, want 0", fs.NArg())
	}

	// The second terminator is an argument.
	fs.Parse([]string{"--", "--"})
	if fs.NArg() != 1 || fs.Arg(0) != "--" {
		t.Errorf("Args() = [%s], want [--]", fs.Arg(0))
	}
}

func TestParseSingleDash(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var b bool
	fs.BoolVar(&b, "bool", false, "bool value")

	// A single dash is an argument, and it stops the flags.
	err := fs.Parse([]string{"-bool", "-", "-bool"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if fs.NArg() != 2 || fs.Arg(0) != "-" || fs.Arg(1) != "-bool" {
		t.Errorf("Args() = [%s %s], want [- -bool]", fs.Arg(0), fs.Arg(1))
	}
}

func TestParseNonFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	// The first argument that is not a flag stops the flags.
	err := fs.Parse([]string{"-bool", "arg", "-int", "5"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if !v.b {
		t.Error("bool = false, want true")
	}
	if v.i != 0 {
		t.Errorf("int = %d, want 0", v.i)
	}
	if fs.NArg() != 3 {
		t.Errorf("NArg() = %d, want 3", fs.NArg())
	}
}

func TestParseEqualsValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var s string
	fs.StringVar(&s, "string", "x", "string value")

	// The value is the text after the first equals sign.
	fs.Parse([]string{"-string="})
	if s != "" {
		t.Errorf("-string=: string = %s, want empty", s)
	}
	fs.Parse([]string{"-string=a=b"})
	if s != "a=b" {
		t.Errorf("-string=a=b: string = %s, want a=b", s)
	}
	fs.Parse([]string{"-string=-1"})
	if s != "-1" {
		t.Errorf("-string=-1: string = %s, want -1", s)
	}
}

func TestParseNegativeValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var i int
	fs.IntVar(&i, "int", 0, "int value")

	// The argument after a flag is its value, even with a leading dash.
	err := fs.Parse([]string{"-int", "-5"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if i != -5 {
		t.Errorf("-int -5: int = %d, want -5", i)
	}

	fs.Parse([]string{"-int=-6"})
	if i != -6 {
		t.Errorf("-int=-6: int = %d, want -6", i)
	}
}

func TestParseRepeated(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var i int
	fs.IntVar(&i, "int", 0, "int value")

	// The last value of a repeated flag wins.
	fs.Parse([]string{"-int", "1", "-int", "2", "-int=3"})
	if i != 3 {
		t.Errorf("int = %d, want 3", i)
	}
}

func TestParseEmpty(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	err := fs.Parse(nil)
	if err != nil {
		t.Errorf("Parse(nil) = %s, want nil", errName(err))
	}
	if !fs.Parsed() {
		t.Error("Parsed() = false after Parse")
	}
	if fs.NArg() != 0 {
		t.Errorf("NArg() = %d, want 0", fs.NArg())
	}

	err = fs.Parse([]string{})
	if err != nil {
		t.Errorf("Parse([]) = %s, want nil", errName(err))
	}
	if fs.NArg() != 0 {
		t.Errorf("NArg() = %d, want 0", fs.NArg())
	}
}

func TestParseArgIndex(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var b bool
	fs.BoolVar(&b, "bool", false, "bool value")

	fs.Parse([]string{"-bool", "a", "b", "c"})
	if fs.NArg() != 3 {
		t.Fatalf("NArg() = %d, want 3", fs.NArg())
		return
	}
	want := []string{"a", "b", "c"}
	for i, s := range want {
		if fs.Arg(i) != s {
			t.Errorf("Arg(%d) = %s, want %s", i, fs.Arg(i), s)
		}
	}
	// An index out of range gives an empty string.
	if fs.Arg(-1) != "" {
		t.Errorf("Arg(-1) = %s, want empty", fs.Arg(-1))
	}
	if fs.Arg(3) != "" {
		t.Errorf("Arg(3) = %s, want empty", fs.Arg(3))
	}
}

func TestParseUnknown(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var v vals
	defineAll(&fs, &v)

	err := fs.Parse([]string{"-nosuchflag"})
	if err != flag.ErrNotFound {
		t.Errorf("Parse(-nosuchflag) = %s, want ErrNotFound", errName(err))
	}

	// A known flag before the unknown one is set.
	err = fs.Parse([]string{"-bool", "-nosuchflag"})
	if err != flag.ErrNotFound {
		t.Errorf("Parse(-bool -nosuchflag) = %s, want ErrNotFound", errName(err))
	}
	if !v.b {
		t.Error("bool = false, want true")
	}
}

func TestParseBadSyntax(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var v vals
	defineAll(&fs, &v)

	// A name that starts with a dash or an equals sign is bad syntax.
	bad := []string{"-=x", "---x", "--=x", "-==", "---"}
	for _, arg := range bad {
		args := []string{arg}
		err := fs.Parse(args)
		if err != flag.ErrSyntax {
			t.Errorf("Parse(%s) = %s, want ErrSyntax", arg, errName(err))
		}
	}
}

func TestParseNeedsArgument(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var v vals
	defineAll(&fs, &v)

	// Every flag but a bool needs a value.
	names := []string{"int", "int64", "uint", "uint64", "float64", "string"}
	for _, name := range names {
		args := []string{"-" + name}
		err := fs.Parse(args)
		if err != flag.ErrSyntax {
			t.Errorf("Parse(-%s) = %s, want ErrSyntax", name, errName(err))
		}
	}
}

func TestParseValueError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var v vals
	defineAll(&fs, &v)

	// A string flag cannot fail, so it is not in the list.
	names := []string{"bool", "int", "int64", "uint", "uint64", "float64"}
	for _, name := range names {
		eq := []string{"-" + name + "=x"}
		err := fs.Parse(eq)
		if err != flag.ErrParse {
			t.Errorf("Parse(-%s=x) = %s, want ErrParse", name, errName(err))
		}
	}

	// The value of the next argument fails the same way.
	next := []string{"-int", "x"}
	err := fs.Parse(next)
	if err != flag.ErrParse {
		t.Errorf("Parse(-int x) = %s, want ErrParse", errName(err))
	}
}

func TestParseRangeError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var v vals
	defineAll(&fs, &v)

	bad := []string{
		"-int=123456789012345678901",
		"-int64=123456789012345678901",
		"-uint=123456789012345678901",
		"-uint64=123456789012345678901",
		"-float64=1e1000",
	}
	for _, arg := range bad {
		args := []string{arg}
		err := fs.Parse(args)
		if err != flag.ErrRange {
			t.Errorf("Parse(%s) = %s, want ErrRange", arg, errName(err))
		}
	}
}

func TestParseHelp(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var b bool
	fs.BoolVar(&b, "bool", false, "bool value")

	// An undefined help flag stops the parse with ErrHelp.
	help := []string{"-help", "-h", "--help", "--h"}
	for _, arg := range help {
		args := []string{arg}
		err := fs.Parse(args)
		if err != flag.ErrHelp {
			t.Errorf("Parse(%s) = %s, want ErrHelp", arg, errName(err))
		}
	}

	// A flag before the help flag is set.
	b = false
	err := fs.Parse([]string{"-bool", "-help"})
	if err != flag.ErrHelp {
		t.Errorf("Parse(-bool -help) = %s, want ErrHelp", errName(err))
	}
	if !b {
		t.Error("bool = false, want true")
	}
}

func TestParseHelpDefined(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var help bool
	fs.BoolVar(&help, "help", false, "help flag")
	var h bool
	fs.BoolVar(&h, "h", false, "h flag")

	// A defined help flag wins over the message.
	err := fs.Parse([]string{"-help", "-h"})
	if err != nil {
		t.Errorf("Parse(-help -h) = %s, want nil", errName(err))
	}
	if !help {
		t.Error("help = false, want true")
	}
	if !h {
		t.Error("h = false, want true")
	}
}

func TestParseUserDefined(t *testing.T) {
	var fs flag.FlagSet
	fs.Init("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var lst listValue
	fs.Var(&lst, "v", "usage")

	// A user Value takes the next argument, and Set runs in command line order.
	err := fs.Parse([]string{"-v", "1", "-v", "2", "-v=3"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if lst.n != 3 {
		t.Fatalf("list holds %d values, want 3", lst.n)
		return
	}
	want := []string{"1", "2", "3"}
	for i, s := range want {
		if lst.items[i] != s {
			t.Errorf("list[%d] = %s, want %s", i, lst.items[i], s)
		}
	}
}

func TestParseUserDefinedError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var fail failValue
	fs.Var(&fail, "fail", "fail value")

	// Parse returns the error of the Value as it is.
	err := fs.Parse([]string{"-fail", "x"})
	if err != errFail {
		t.Errorf("Parse(-fail x) = %s, want errFail", errName(err))
	}
	if fail.calls != 1 {
		t.Errorf("Set called %d times, want 1", fail.calls)
	}

	// An unset Value never fails the parse.
	err = fs.Parse(nil)
	if err != nil {
		t.Errorf("Parse(nil) = %s, want nil", errName(err))
	}
	if fail.calls != 1 {
		t.Errorf("Set called %d times, want 1", fail.calls)
	}
}
