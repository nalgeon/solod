// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package flag implements command-line flag parsing.

# Usage

Define flags using [StringVar], [BoolVar], [IntVar], etc.

	var flagvar int
	func init() {
		flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname")
	}

You can create custom flags that satisfy the [Value] interface (with
pointer receivers) and couple them to flag parsing by

	flag.Var(&flagVal, "name", "help message for flagname")

For such flags, the default value is just the initial value of the variable.

After all flags are defined, call [Parse] to parse the command line into the defined flags.

	flag.Parse()

Flags may then be used directly.

	fmt.Println("flagvar has value ", flagvar)

After parsing, the arguments following the flags are available as the
slice [flag.Args] or individually as [flag.Arg](i).
The arguments are indexed from 0 through [flag.NArg]-1.

# Command line flag syntax

The following forms are permitted:

	-flag
	--flag   // double dashes are also permitted
	-flag=x
	-flag x  // non-boolean flags only

One or two dashes may be used; they are equivalent.
The last form is not permitted for boolean flags because the
meaning of the command

	cmd -x *

where * is a Unix shell wildcard, will change if there is a file
called 0, false, etc. You must use the -flag=false form to turn
off a boolean flag.

Flag parsing stops just before the first non-flag argument
("-" is a non-flag argument) or after the terminator "--".

Integer flags accept 1234, 0664, 0x1234 and may be negative.
Boolean flags may be:

	1, 0, t, f, T, F, true, false, TRUE, FALSE, True, False

The default set of command-line flags is controlled by
top-level functions.  The [FlagSet] type allows one to define
independent sets of flags, such as to implement subcommands
in a command-line interface. The methods of [FlagSet] are
analogous to the top-level functions for the command-line
flag set.

Based on the [flag] package.

[flag]: https://github.com/golang/go/blob/go1.26.2/src/flag
*/
package flag

import (
	"solod.dev/so/errors"
	"solod.dev/so/fmt"
	"solod.dev/so/io"
	"solod.dev/so/os"
	"solod.dev/so/strings"
)

const MaxFlags = 64

var (
	// ErrHelp is the error returned if the -help or -h flag is invoked
	// but no such flag is defined.
	ErrHelp = errors.New("flag: help requested")

	// ErrNotFound is returned by Set and Parse if the flag does not exist.
	ErrNotFound = errors.New("flag: not found")

	// ErrParse is returned by Set if a flag's value fails to parse,
	// such as with an invalid integer for Int.
	ErrParse = errors.New("flag: parse error")

	// ErrRange is returned by Set if a flag's value is out of range.
	ErrRange = errors.New("flag: value out of range")

	// ErrSyntax is returned by Parse if the syntax of a flag is invalid.
	ErrSyntax = errors.New("flag: invalid syntax")
)

// ErrorHandling defines how [FlagSet.Parse] behaves if the parse fails.
type ErrorHandling int

// These constants cause [FlagSet.Parse] to behave as described if the parse fails.
const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2) or for -h/-help Exit(0).
	PanicOnError                         // Call panic with a descriptive error.
)

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
type Value interface {
	Get() any
	Set(string) error
	Type() string
}

// A Flag represents the state of a flag.
type Flag struct {
	Name  string // name as it appears on command line
	Usage string // help message
	Value Value  // value as set
}

// Type is an educated guess of the type of the flag's value,
// or the empty string if the flag is boolean.
func (flag *Flag) Type() string {
	name := flag.Value.Type()
	if name == "bool" {
		return ""
	}
	if name == "" {
		name = "value"
	}
	return name
}

// A FlagSet represents a set of defined flags. The zero value of a FlagSet
// has no name and has [ContinueOnError] error handling.
//
// [Flag] names must be unique within a FlagSet. An attempt to define a flag whose
// name is already in use will cause a panic.
type FlagSet struct {
	name          string
	parsed        bool
	nflag         int
	flags         [MaxFlags]Flag
	args          []string // arguments after flags
	errorHandling ErrorHandling
	output        io.Writer // nil means stderr; use Output() accessor
}

// NewFlagSet returns a new, empty flag set with the specified name and
// error handling property. If the name is not empty, it will be printed
// in the default usage message and in error messages.
func NewFlagSet(name string, errorHandling ErrorHandling) FlagSet {
	return FlagSet{
		name:          name,
		errorHandling: errorHandling,
	}
}

// Init sets the name and error handling property for a flag set.
// By default, the zero [FlagSet] uses an empty name and the
// [ContinueOnError] error handling policy.
func (f *FlagSet) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}

// Name returns the name of the flag set.
func (f *FlagSet) Name() string {
	return f.name
}

// ErrorHandling returns the error handling behavior of the flag set.
func (f *FlagSet) ErrorHandling() ErrorHandling {
	return f.errorHandling
}

// NFlag returns the number of defined flags.
func (f *FlagSet) NFlag() int { return f.nflag }

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// NArg is the number of arguments remaining after flags have been processed.
func (f *FlagSet) NArg() int { return len(f.args) }

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string { return f.args }

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// Output returns the destination for usage and error messages. [os.Stderr] is returned if
// output was not set or was set to nil.
func (f *FlagSet) Output() io.Writer {
	if f.output == nil {
		return os.Stderr
	}
	return f.output
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, [os.Stderr] is used.
func (f *FlagSet) SetOutput(output io.Writer) {
	f.output = output
}

// Lookup returns the [Flag] structure of the named flag, returning nil if none exists.
func (f *FlagSet) Lookup(name string) *Flag {
	idx := f.find(name)
	if idx == -1 {
		return nil
	}
	return &f.flags[idx]
}

// Set sets the value of the named flag.
func (f *FlagSet) Set(name, value string) error {
	idx := f.find(name)
	if idx == -1 {
		return ErrNotFound
	}
	err := f.flags[idx].Value.Set(value)
	return err
}

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the [FlagSet]
// are defined and before flags are accessed by the program.
// The return value will be [ErrHelp] if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = arguments
	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			if err == ErrHelp {
				os.Exit(0)
			}
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

// Usage is called when an error occurs while parsing flags.
// What happens after Usage is called depends on the ErrorHandling setting;
// for the command line, this defaults to ExitOnError, which exits the program
// after calling Usage.
func (f *FlagSet) Usage() {
	if f.name == "" {
		fmt.Fprintf(f.Output(), "Usage:\n")
	} else {
		fmt.Fprintf(f.Output(), "Usage of %s:\n", f.name)
	}
	f.PrintDefaults()
}

// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined command-line flags in the set. See the
// documentation for the global function PrintDefaults for more information.
func (f *FlagSet) PrintDefaults() {
	out := f.Output()
	for i := range f.nflag {
		flag := f.flags[i]
		// n counts the bytes before the usage message.
		n := 3 + len(flag.Name) // Two spaces before -; see next two comments.
		io.WriteString(out, "  -")
		io.WriteString(out, flag.Name)
		name := flag.Type()
		if len(name) > 0 {
			n += 1 + len(name)
			io.WriteString(out, " ")
			io.WriteString(out, name)
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if n <= 4 { // space, space, '-', 'x'.
			io.WriteString(out, "\t")
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			io.WriteString(out, "\n    \t")
		}
		io.WriteString(out, flag.Usage)
		io.WriteString(out, "\n")
	}
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	*p = value
	f.Var((*boolValue)(p), name, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	*p = value
	f.Var((*intValue)(p), name, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (f *FlagSet) Int64Var(p *int64, name string, value int64, usage string) {
	*p = value
	f.Var((*int64Value)(p), name, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (f *FlagSet) UintVar(p *uint, name string, value uint, usage string) {
	*p = value
	f.Var((*uintValue)(p), name, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (f *FlagSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
	*p = value
	f.Var((*uint64Value)(p), name, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (f *FlagSet) Float64Var(p *float64, name string, value float64, usage string) {
	*p = value
	f.Var((*float64Value)(p), name, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {
	*p = value
	f.Var((*stringValue)(p), name, usage)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type [Value], which
// typically holds a user-defined implementation of [Value]. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of [Value]; in particular, [Set] would
// decompose the comma-separated string into the slice.
func (f *FlagSet) Var(value Value, name string, usage string) {
	// Flag must not begin "-" or contain "=".
	if strings.HasPrefix(name, "-") {
		panic("flag '" + name + "' begins with -")
	} else if strings.Contains(name, "=") {
		panic("flag '" + name + "' contains =")
	}

	flag := Flag{name, usage, value}
	idx := f.find(name)
	if idx != -1 {
		panic("flag '" + name + "' redefined")
	}
	if f.nflag >= MaxFlags {
		panic("too many flags defined")
	}
	f.flags[f.nflag] = flag
	f.nflag++
}

// parseOne parses one flag. It reports whether a flag was seen.
func (f *FlagSet) parseOne() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}
	s := f.args[0]
	if len(s) < 2 || s[0] != '-' {
		return false, nil
	}
	numMinuses := 1
	if s[1] == '-' {
		numMinuses++
		if len(s) == 2 { // "--" terminates the flags
			f.args = f.args[1:]
			return false, nil
		}
	}
	name := s[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		f.failf("bad flag syntax: %s\n", s, nil)
		return false, ErrSyntax
	}

	// it's a flag. does it have an argument?
	f.args = f.args[1:]
	hasValue := false
	value := ""
	for i := 1; i < len(name); i++ { // equals cannot be first
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}

	idx := f.find(name)
	if idx == -1 {
		if name == "help" || name == "h" { // special case for nice help message.
			f.Usage()
			return false, ErrHelp
		}
		f.failf("flag -%s provided but not defined\n", name, nil)
		return false, ErrNotFound
	}

	flag := &f.flags[idx]
	if _, ok := flag.Value.(*boolValue); ok { // special case: doesn't need an arg
		if hasValue {
			if err := flag.Value.Set(value); err != nil {
				f.failf("invalid boolean value for flag -%s: %s\n", name, &value)
				return false, err
			}
		} else {
			if err := flag.Value.Set("true"); err != nil {
				f.failf("invalid boolean flag -%s\n", name, nil)
				return false, err
			}
		}
	} else {
		// It must have a value, which might be the next argument.
		if !hasValue && len(f.args) > 0 {
			// value is the next arg
			hasValue = true
			value = f.args[0]
			f.args = f.args[1:]
		}
		if !hasValue {
			f.failf("flag -%s needs an argument\n", name, nil)
			return false, ErrSyntax
		}
		if err := flag.Value.Set(value); err != nil {
			f.failf("invalid value for flag -%s: %s\n", name, &value)
			return false, err
		}
	}
	return true, nil
}

// failf prints to standard error a formatted error and usage message.
func (f *FlagSet) failf(format string, name string, value *string) {
	if value != nil {
		fmt.Fprintf(f.Output(), format, name, *value)
	} else {
		fmt.Fprintf(f.Output(), format, name)
	}
	f.Usage()
}

// find returns the index of the flag named name, or -1 if none exists.
func (f *FlagSet) find(name string) int {
	for i := range f.nflag {
		if f.flags[i].Name == name {
			return i
		}
	}
	return -1
}
