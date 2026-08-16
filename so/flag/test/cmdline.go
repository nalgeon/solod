// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"solod.dev/so/flag"
	"solod.dev/so/os"
	"solod.dev/so/strings"
	"solod.dev/so/testing"
)

// The flags of this file go to the command line set, which the whole program
// shares. Their names carry a prefix of their own, so they never collide with
// the flags of another package.

func TestCommandLine(t *testing.T) {
	var b bool
	flag.BoolVar(&b, "cl_bool", false, "bool value")
	var i int
	flag.IntVar(&i, "cl_int", 0, "int value")
	var i64 int64
	flag.Int64Var(&i64, "cl_int64", 0, "int64 value")
	var u uint
	flag.UintVar(&u, "cl_uint", 0, "uint value")
	var u64 uint64
	flag.Uint64Var(&u64, "cl_uint64", 0, "uint64 value")
	var f float64
	flag.Float64Var(&f, "cl_float64", 0, "float64 value")
	var s string
	flag.StringVar(&s, "cl_string", "0", "string value")
	var lst listValue
	flag.Var(&lst, "cl_list", "list value")

	// The first wrapper call builds the command line set.
	if flag.CommandLine == nil {
		t.Fatal("CommandLine = nil, want a flag set")
		return
	}
	if flag.CommandLine.ErrorHandling() != flag.ExitOnError {
		t.Errorf("ErrorHandling() = %d, want %d",
			int(flag.CommandLine.ErrorHandling()), int(flag.ExitOnError))
	}
	// The name of the set is the name of the program. The test runner takes no
	// argc and argv, so os.Args is empty here and the name is empty too.
	want := ""
	if len(os.Args) > 0 {
		want = os.Args[0]
	}
	if flag.CommandLine.Name() != want {
		t.Errorf("Name() = %s, want %s", flag.CommandLine.Name(), want)
	}

	// Every wrapper defines its flag on the command line set.
	names := []string{
		"cl_bool", "cl_int", "cl_int64", "cl_uint",
		"cl_uint64", "cl_float64", "cl_string", "cl_list",
	}
	for _, name := range names {
		if flag.CommandLine.Lookup(name) == nil {
			t.Errorf("Lookup(%s) = nil, want a flag", name)
		}
	}
}

func TestCommandLineParse(t *testing.T) {
	var b bool
	flag.BoolVar(&b, "clp_bool", false, "bool value")
	var i int
	flag.IntVar(&i, "clp_int", 0, "int value")

	// The command line set exits on error, so the arguments must be valid.
	err := flag.CommandLine.Parse([]string{"-clp_bool", "-clp_int", "7", "extra"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if !b {
		t.Error("clp_bool = false, want true")
	}
	if i != 7 {
		t.Errorf("clp_int = %d, want 7", i)
	}

	// Args reads the arguments of the command line set.
	args := flag.Args()
	if len(args) != 1 || args[0] != "extra" {
		t.Fatalf("Args() holds %d values, want 1", len(args))
		return
	}
}

func TestCommandLineUsage(t *testing.T) {
	// A wrapper call builds the command line set, so CommandLine is not nil
	// after this one.
	var b bool
	flag.BoolVar(&b, "clu_bool", false, "clu bool value")

	var buf [4096]byte
	out := strings.FixedBuilder(buf[:0])
	flag.CommandLine.SetOutput(&out)

	// Usage writes the header and the flags to the output of the set.
	flag.Usage()
	flag.CommandLine.SetOutput(nil)

	got := out.String()
	if !strings.HasPrefix(got, "Usage") {
		t.Errorf("Usage() = %s, want a header that starts with Usage", got)
	}
	if !strings.Contains(got, "  -clu_bool\n    \tclu bool value\n") {
		t.Errorf("Usage() = %s, want a mention of -clu_bool", got)
	}
}
