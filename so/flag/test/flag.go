// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"solod.dev/so/flag"
	"solod.dev/so/io"
	"solod.dev/so/os"
	"solod.dev/so/testing"
)

func TestDefine(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if fs.NFlag() != 0 {
		t.Errorf("NFlag() = %d, want 0", fs.NFlag())
	}

	var v vals
	defineAll(&fs, &v)
	if fs.NFlag() != 7 {
		t.Errorf("NFlag() = %d, want 7", fs.NFlag())
	}

	var lst listValue
	fs.Var(&lst, "list", "list value")
	if fs.NFlag() != 8 {
		t.Errorf("NFlag() = %d, want 8", fs.NFlag())
	}
}

func TestDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var b bool
	fs.BoolVar(&b, "bool", true, "bool value")
	var i int
	fs.IntVar(&i, "int", -11, "int value")
	var i64 int64
	fs.Int64Var(&i64, "int64", -22, "int64 value")
	var u uint
	fs.UintVar(&u, "uint", 33, "uint value")
	var u64 uint64
	fs.Uint64Var(&u64, "uint64", 44, "uint64 value")
	var f float64
	fs.Float64Var(&f, "float64", 2.5, "float64 value")
	var s string
	fs.StringVar(&s, "string", "hi", "string value")

	if !b {
		t.Error("bool default = false, want true")
	}
	if i != -11 {
		t.Errorf("int default = %d, want -11", i)
	}
	if i64 != -22 {
		t.Errorf("int64 default = %d, want -22", int(i64))
	}
	if u != 33 {
		t.Errorf("uint default = %d, want 33", int(u))
	}
	if u64 != 44 {
		t.Errorf("uint64 default = %d, want 44", int(u64))
	}
	if f != 2.5 {
		t.Errorf("float64 default = %g, want 2.5", f)
	}
	if s != "hi" {
		t.Errorf("string default = %s, want hi", s)
	}
}

func TestLookup(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	fl := fs.Lookup("int")
	if fl == nil {
		t.Fatal("Lookup(int) = nil, want a flag")
		return
	}
	if fl.Name != "int" {
		t.Errorf("Lookup(int).Name = %s, want int", fl.Name)
	}
	if fl.Usage != "int value" {
		t.Errorf("Lookup(int).Usage = %s, want int value", fl.Usage)
	}
	if fl.Value == nil {
		t.Error("Lookup(int).Value = nil, want a value")
	}

	if fs.Lookup("nosuchflag") != nil {
		t.Error("Lookup(nosuchflag) != nil, want nil")
	}
	if fs.Lookup("") != nil {
		t.Error("Lookup() != nil, want nil")
	}
}

func TestFlagType(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)
	var lst listValue
	fs.Var(&lst, "list", "list value")
	var empty emptyValue
	fs.Var(&empty, "empty", "empty value")

	names := []string{"bool", "int", "int64", "uint", "uint64", "float64", "string", "list", "empty"}
	// A bool flag has no type name. An int64 reads as int, a uint64 as uint,
	// and a Value with an empty type name as value.
	types := []string{"", "int", "int", "uint", "uint", "float", "string", "list", "value"}
	for i, name := range names {
		got := fs.Lookup(name).Type()
		if got != types[i] {
			t.Errorf("Lookup(%s).Type() = %s, want %s", name, got, types[i])
		}
	}
}

func TestGet(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var v vals
	defineAll(&fs, &v)

	// Get returns the address of the variable behind the flag.
	if fs.Lookup("bool").Value.Get().(*bool) != &v.b {
		t.Error("bool Get() != &v.b")
	}
	if fs.Lookup("int").Value.Get().(*int) != &v.i {
		t.Error("int Get() != &v.i")
	}
	if fs.Lookup("int64").Value.Get().(*int64) != &v.i64 {
		t.Error("int64 Get() != &v.i64")
	}
	if fs.Lookup("uint").Value.Get().(*uint) != &v.u {
		t.Error("uint Get() != &v.u")
	}
	if fs.Lookup("uint64").Value.Get().(*uint64) != &v.u64 {
		t.Error("uint64 Get() != &v.u64")
	}
	if fs.Lookup("float64").Value.Get().(*float64) != &v.f {
		t.Error("float64 Get() != &v.f")
	}
	if fs.Lookup("string").Value.Get().(*string) != &v.s {
		t.Error("string Get() != &v.s")
	}

	// The address reads the value the flag set.
	err := fs.Set("int", "42")
	if err != nil {
		t.Errorf("Set(int, 42) = %s, want nil", errName(err))
	}
	p := fs.Lookup("int").Value.Get().(*int)
	if *p != 42 {
		t.Errorf("int Get() = %d, want 42", *p)
	}
}

func TestGetUserDefined(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var lst listValue
	fs.Var(&lst, "list", "list value")

	got := fs.Lookup("list").Value.Get().(*listValue)
	if got != &lst {
		t.Error("list Get() != &lst")
	}
}

func TestZeroFlagSet(t *testing.T) {
	// The zero FlagSet has no name and uses ContinueOnError.
	var fs flag.FlagSet
	if fs.Name() != "" {
		t.Errorf("Name() = %s, want empty", fs.Name())
	}
	if fs.ErrorHandling() != flag.ContinueOnError {
		t.Errorf("ErrorHandling() = %d, want %d", int(fs.ErrorHandling()), int(flag.ContinueOnError))
	}
	if fs.NFlag() != 0 {
		t.Errorf("NFlag() = %d, want 0", fs.NFlag())
	}
	if fs.NArg() != 0 {
		t.Errorf("NArg() = %d, want 0", fs.NArg())
	}
	if len(fs.Args()) != 0 {
		t.Errorf("len(Args()) = %d, want 0", len(fs.Args()))
	}
	if fs.Parsed() {
		t.Error("Parsed() = true, want false")
	}
	if fs.Lookup("x") != nil {
		t.Error("Lookup(x) != nil, want nil")
	}

	// A zero FlagSet parses.
	var b bool
	fs.BoolVar(&b, "b", false, "b value")
	err := fs.Parse([]string{"-b"})
	if err != nil {
		t.Errorf("Parse() = %s, want nil", errName(err))
	}
	if !b {
		t.Error("b = false, want true")
	}
}

func TestGetters(t *testing.T) {
	fs := flag.NewFlagSet("flag set", flag.ContinueOnError)
	if fs.Name() != "flag set" {
		t.Errorf("Name() = %s, want flag set", fs.Name())
	}
	if fs.ErrorHandling() != flag.ContinueOnError {
		t.Errorf("ErrorHandling() = %d, want %d", int(fs.ErrorHandling()), int(flag.ContinueOnError))
	}
	if fs.Output() != io.Writer(os.Stderr) {
		t.Error("Output() != os.Stderr")
	}

	// Init replaces the name and the error handling of the set.
	fs.Init("gopher", flag.ExitOnError)
	fs.SetOutput(os.Stdout)
	if fs.Name() != "gopher" {
		t.Errorf("Name() = %s, want gopher", fs.Name())
	}
	if fs.ErrorHandling() != flag.ExitOnError {
		t.Errorf("ErrorHandling() = %d, want %d", int(fs.ErrorHandling()), int(flag.ExitOnError))
	}
	if fs.Output() != io.Writer(os.Stdout) {
		t.Error("Output() != os.Stdout")
	}

	// A nil output falls back to stderr.
	fs.SetOutput(nil)
	if fs.Output() != io.Writer(os.Stderr) {
		t.Error("Output() != os.Stderr after SetOutput(nil)")
	}
}

func TestErrorHandlingValues(t *testing.T) {
	if flag.ContinueOnError != 0 {
		t.Errorf("ContinueOnError = %d, want 0", int(flag.ContinueOnError))
	}
	if flag.ExitOnError != 1 {
		t.Errorf("ExitOnError = %d, want 1", int(flag.ExitOnError))
	}
	if flag.PanicOnError != 2 {
		t.Errorf("PanicOnError = %d, want 2", int(flag.PanicOnError))
	}
}
