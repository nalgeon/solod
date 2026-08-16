// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"solod.dev/so/flag"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// defaultOutput is the PrintDefaults output of the flag set of
// TestPrintDefaults. A flag of one letter keeps its usage on the same line.
const defaultOutput = `  -A	for bootstrapping, allow 'any' type
  -Alongflagname
    	disable bounds checking
  -C	a boolean defaulting to true (default true)
  -D string
    	set relative path for local imports
  -E string
    	issue 23543 (default "0")
  -F float
    	a non-zero number (default 2.7)
  -G float
    	a float that defaults to zero
  -N int
    	a non-zero int (default 27)
  -U uint
    	an unsigned int
  -V list
    	a list of strings (default [a b])
  -W value
    	a value with no type name
  -Z int
    	an int that defaults to zero
`

func TestPrintDefaults(t *testing.T) {
	fs := flag.NewFlagSet("print defaults test", flag.ContinueOnError)
	var buf [1024]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)

	var a bool
	fs.BoolVar(&a, "A", false, "for bootstrapping, allow 'any' type")
	var alongflagname bool
	fs.BoolVar(&alongflagname, "Alongflagname", false, "disable bounds checking")
	var c bool
	fs.BoolVar(&c, "C", true, "a boolean defaulting to true (default true)")
	var d string
	fs.StringVar(&d, "D", "", "set relative path for local imports")
	var e string
	fs.StringVar(&e, "E", "0", "issue 23543 (default \"0\")")
	var f float64
	fs.Float64Var(&f, "F", 2.7, "a non-zero number (default 2.7)")
	var g float64
	fs.Float64Var(&g, "G", 0, "a float that defaults to zero")
	var n int
	fs.IntVar(&n, "N", 27, "a non-zero int (default 27)")
	var u uint
	fs.UintVar(&u, "U", 0, "an unsigned int")
	var lst listValue
	fs.Var(&lst, "V", "a list of strings (default [a b])")
	var empty emptyValue
	fs.Var(&empty, "W", "a value with no type name")
	var z int
	fs.IntVar(&z, "Z", 0, "an int that defaults to zero")

	fs.PrintDefaults()
	got := b.String()
	if got != defaultOutput {
		t.Errorf("PrintDefaults() = \n%s\nwant\n%s", got, defaultOutput)
	}
}

func TestPrintDefaultsEmpty(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [64]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)

	// A set with no flags prints nothing.
	fs.PrintDefaults()
	if b.String() != "" {
		t.Errorf("PrintDefaults() = %s, want empty", b.String())
	}
}

func TestPrintDefaultsOrder(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)

	// The flags print in definition order, not in name order.
	var z bool
	fs.BoolVar(&z, "z", false, "z value")
	var a bool
	fs.BoolVar(&a, "a", false, "a value")

	fs.PrintDefaults()
	want := "  -z\tz value\n  -a\ta value\n"
	if b.String() != want {
		t.Errorf("PrintDefaults() = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestUsage(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)
	var n int
	fs.IntVar(&n, "n", 0, "n value")

	// A named set prints its name in the header.
	fs.Usage()
	want := "Usage of test:\n  -n int\n    \tn value\n"
	if b.String() != want {
		t.Errorf("Usage() = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestUsageNoName(t *testing.T) {
	var fs flag.FlagSet
	var buf [64]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)

	// A set with no name prints a header with no name.
	fs.Usage()
	if b.String() != "Usage:\n" {
		t.Errorf("Usage() = %s, want Usage:", b.String())
	}
}

func TestFailUnknown(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)

	fs.Parse([]string{"-nosuchflag"})
	want := "flag -nosuchflag provided but not defined\nUsage of test:\n"
	if b.String() != want {
		t.Errorf("output = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestFailBadSyntax(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)

	fs.Parse([]string{"-=x"})
	want := "bad flag syntax: -=x\nUsage of test:\n"
	if b.String() != want {
		t.Errorf("output = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestFailNeedsArgument(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)
	var n int
	fs.IntVar(&n, "n", 0, "n value")

	fs.Parse([]string{"-n"})
	want := "flag -n needs an argument\nUsage of test:\n  -n int\n    \tn value\n"
	if b.String() != want {
		t.Errorf("output = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestFailBadValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)
	var n int
	fs.IntVar(&n, "n", 0, "n value")

	fs.Parse([]string{"-n=x"})
	want := "invalid value for flag -n: x\nUsage of test:\n  -n int\n    \tn value\n"
	if b.String() != want {
		t.Errorf("output = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestFailBadBool(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)
	var v bool
	fs.BoolVar(&v, "v", false, "v value")

	fs.Parse([]string{"-v=x"})
	want := "invalid boolean value for flag -v: x\nUsage of test:\n  -v\tv value\n"
	if b.String() != want {
		t.Errorf("output = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestHelpOutput(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)
	var n int
	fs.IntVar(&n, "n", 0, "n value")

	// The help flag prints the usage and no error message.
	fs.Parse([]string{"-help"})
	want := "Usage of test:\n  -n int\n    \tn value\n"
	if b.String() != want {
		t.Errorf("output = \n%s\nwant\n%s", b.String(), want)
	}
}

func TestOutputIsWritten(t *testing.T) {
	var fs flag.FlagSet
	var buf [128]byte
	b := strings.FixedBuilder(buf[:0])
	fs.SetOutput(&b)
	fs.Init("test", flag.ContinueOnError)

	// SetOutput takes the messages away from stderr.
	fs.Parse([]string{"-unknown"})
	if !strings.Contains(b.String(), "-unknown") {
		t.Errorf("output = %s, want a mention of -unknown", b.String())
	}
}
