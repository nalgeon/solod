// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flag_test

import (
	"solod.dev/so/errors"
	"solod.dev/so/flag"
)

// errFail is the error of the Value that fails on purpose.
var errFail = errors.New("flag_test: set failed")

// errName returns the name of a flag package error.
func errName(err error) string {
	switch err {
	case nil:
		return "nil"
	case flag.ErrHelp:
		return "ErrHelp"
	case flag.ErrNotFound:
		return "ErrNotFound"
	case flag.ErrParse:
		return "ErrParse"
	case flag.ErrRange:
		return "ErrRange"
	case flag.ErrSyntax:
		return "ErrSyntax"
	case errFail:
		return "errFail"
	}
	return "other"
}

// vals holds one variable for every flag type of the package.
type vals struct {
	b   bool
	i   int
	i64 int64
	u   uint
	u64 uint64
	f   float64
	s   string
}

// defineAll defines one flag of every type, all with a zero default value.
func defineAll(fs *flag.FlagSet, v *vals) {
	fs.BoolVar(&v.b, "bool", false, "bool value")
	fs.IntVar(&v.i, "int", 0, "int value")
	fs.Int64Var(&v.i64, "int64", 0, "int64 value")
	fs.UintVar(&v.u, "uint", 0, "uint value")
	fs.Uint64Var(&v.u64, "uint64", 0, "uint64 value")
	fs.Float64Var(&v.f, "float64", 0, "float64 value")
	fs.StringVar(&v.s, "string", "0", "string value")
}

// The values that overflow a 32 bit int and a 32 bit uint. They are variables,
// not constants, because a constant of this width makes the C compiler reject
// the comparison against a 32 bit int with -Wtype-limits.
var (
	int32Overflow  int64  = 2147483648
	uint32Overflow uint64 = 4294967296
)

// maxItems is the number of values a listValue holds.
const maxItems = 8

// listValue is a Value that collects every value it is set to.
type listValue struct {
	items [maxItems]string
	n     int
}

func (v *listValue) Set(s string) error {
	v.items[v.n] = s
	v.n++
	return nil
}

func (v *listValue) Get() any { return v }

func (v *listValue) Type() string { _ = v; return "list" }

// emptyValue is a Value with an empty type name.
type emptyValue struct {
	s string
}

func (v *emptyValue) Set(s string) error {
	v.s = s
	return nil
}

func (v *emptyValue) Get() any { return v }

func (v *emptyValue) Type() string { _ = v; return "" }

// failValue is a Value that fails on every Set.
type failValue struct {
	calls int
}

func (v *failValue) Set(s string) error {
	_ = s
	v.calls++
	return errFail
}

func (v *failValue) Get() any { return v }

func (v *failValue) Type() string { _ = v; return "fail" }
