// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag

import (
	stdbytes "bytes"
	stdflag "flag"
	stdio "io"
	"math"
	"testing"

	"solod.dev/so/io"
)

// mustPanic runs f and reports whether it panics with the expected message.
func mustPanic(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		switch msg := recover().(type) {
		case nil:
			t.Errorf("expected panic(%q), but did not panic", want)
		case string:
			if msg != want {
				t.Errorf("expected panic(%q), but got panic(%q)", want, msg)
			}
		default:
			t.Errorf("expected panic(%q), but got panic(%v)", want, msg)
		}
	}()
	f()
}

func TestVarPanicsOnBadName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"-foo", "flag '-foo' begins with -"},
		{"--foo", "flag '--foo' begins with -"},
		{"foo=bar", "flag 'foo=bar' contains ="},
		{"=", "flag '=' contains ="},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			fs.SetOutput(io.Discard)
			mustPanic(t, test.want, func() {
				var v stringValue
				fs.Var(&v, test.name, "usage")
			})
		})
	}
}

func TestVarPanicsOnRedefine(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetOutput(io.Discard)
	var v stringValue
	fs.Var(&v, "foo", "usage")
	mustPanic(t, "flag 'foo' redefined", func() {
		fs.Var(&v, "foo", "usage")
	})
}

func TestVarPanicsOnTooManyFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetOutput(io.Discard)
	values := make([]stringValue, MaxFlags+1)
	for i := range MaxFlags {
		fs.Var(&values[i], string(rune('a'+i/26))+string(rune('a'+i%26)), "usage")
	}
	if fs.NFlag() != MaxFlags {
		t.Fatalf("NFlag() = %d, want %d", fs.NFlag(), MaxFlags)
	}
	mustPanic(t, "too many flags defined", func() {
		fs.Var(&values[MaxFlags], "zz", "usage")
	})
}

func TestPanicOnError(t *testing.T) {
	fs := NewFlagSet("test", PanicOnError)
	fs.SetOutput(io.Discard)
	var b bool
	fs.BoolVar(&b, "bool", false, "bool value")

	tests := []struct {
		arg  string
		want error
	}{
		{"-nosuchflag", ErrNotFound},
		{"-=x", ErrSyntax},
		{"-bool=x", ErrParse},
		{"-help", ErrHelp},
	}
	for _, test := range tests {
		t.Run(test.arg, func(t *testing.T) {
			defer func() {
				got := recover()
				if got != test.want {
					t.Errorf("Parse(%q) panicked with %v, want %v", test.arg, got, test.want)
				}
			}()
			fs.Parse([]string{test.arg})
			t.Errorf("Parse(%q) did not panic", test.arg)
		})
	}
}

func FuzzParse(f *testing.F) {
	seeds := []string{
		"-b\x00-n\x0042\x00-s\x00hi\x00extra",
		"-b=true\x00-n=-1\x00-u=0x10\x00-f=2718e28",
		"--n\x000664\x00--s=a=b\x00--\x00-b",
		"-n\x00123456789012345678901",
		"-f\x001e1000",
		"-nosuchflag",
		"-=x",
		"---x",
		"-n",
		"-help",
		"-\x00-b",
		"",
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var args []string
		for part := range stdbytes.SplitSeq(data, []byte{0}) {
			args = append(args, string(part))
		}

		var sb bool
		var sn int
		var su uint
		var sf float64
		var ss string
		fs := NewFlagSet("test", ContinueOnError)
		fs.SetOutput(io.Discard)
		fs.BoolVar(&sb, "b", false, "bool value")
		fs.IntVar(&sn, "n", 0, "int value")
		fs.UintVar(&su, "u", 0, "uint value")
		fs.Float64Var(&sf, "f", 0, "float64 value")
		fs.StringVar(&ss, "s", "", "string value")
		soErr := fs.Parse(args)

		gfs := stdflag.NewFlagSet("test", stdflag.ContinueOnError)
		gfs.SetOutput(stdio.Discard)
		gb := gfs.Bool("b", false, "bool value")
		gn := gfs.Int("n", 0, "int value")
		gu := gfs.Uint("u", 0, "uint value")
		gf := gfs.Float64("f", 0, "float64 value")
		gs := gfs.String("s", "", "string value")
		goErr := gfs.Parse(args)

		if (soErr != nil) != (goErr != nil) {
			t.Fatalf("Parse(%q) = %v, Go = %v", args, soErr, goErr)
		}
		if sb != *gb {
			t.Errorf("Parse(%q): bool = %v, Go = %v", args, sb, *gb)
		}
		if sn != *gn {
			t.Errorf("Parse(%q): int = %d, Go = %d", args, sn, *gn)
		}
		if su != *gu {
			t.Errorf("Parse(%q): uint = %d, Go = %d", args, su, *gu)
		}
		// A NaN never equals itself, so compare the two NaNs apart.
		if sf != *gf && !(math.IsNaN(sf) && math.IsNaN(*gf)) {
			t.Errorf("Parse(%q): float64 = %v, Go = %v", args, sf, *gf)
		}
		if ss != *gs {
			t.Errorf("Parse(%q): string = %q, Go = %q", args, ss, *gs)
		}
		if soErr != nil {
			return
		}
		if fs.NArg() != gfs.NArg() {
			t.Fatalf("Parse(%q): NArg() = %d, Go = %d", args, fs.NArg(), gfs.NArg())
		}
		for i := range fs.NArg() {
			if fs.Arg(i) != gfs.Arg(i) {
				t.Errorf("Parse(%q): Arg(%d) = %q, Go = %q", args, i, fs.Arg(i), gfs.Arg(i))
			}
		}
	})
}
