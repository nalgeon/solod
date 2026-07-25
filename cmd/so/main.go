package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"solod.dev/internal/compiler"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "translate":
		err = translate(args)
	case "build":
		err = build(args)
	case "run":
		err = run(args)
	case "test":
		err = test(args)
	case "bench":
		err = bench(args)
	case "version":
		fmt.Printf("so version %s\n", compiler.Version())
		return
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		// A non-zero exit from the compiled program (e.g. a failing `so test`
		// run, or a program that calls os.Exit) is not a tool error: the
		// program already wrote its own output. Propagate the code silently.
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "so %s: %s\n", cmd, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `So is a tool for managing Solod source code.

Usage: so <command> [arguments]

Commands:
    build        compile package to executable
    bench        run benchmarks in a package's bench subdirectory
    run          compile and run a package
    test         run tests in a package's test subdirectory
    translate    translate package to C
    version      print compiler version

Run 'so <command> -h' for details.
`)
}

const (
	assertUsage      = "assertions: on (default) or off"
	panicModeUsage   = "panic termination mode: trace (default), exit, or abort"
	sanitizeUsage    = "comma-separated list of C sanitizers"
	trackSourceUsage = "track source locations for panics"
)

func translate(args []string) error {
	flags := flag.NewFlagSet("translate", flag.ContinueOnError)
	outDir := flags.String("o", "", "output directory (default: current directory)")
	trackSource := flags.Bool("track-source", false, trackSourceUsage)
	if err := flags.Parse(args); err != nil {
		return err
	}

	pkg := "."
	if flags.NArg() > 0 {
		pkg = flags.Arg(0)
	}

	out := *outDir
	if out == "" {
		out = "."
	}

	opts := compiler.Options{
		TrackSource: *trackSource,
	}
	_, err := compiler.Translate(pkg, out, opts)
	return err
}

func build(args []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	outFile := flags.String("o", "", "output file (default: basename of package directory)")
	assert := flags.String("assert", "on", assertUsage)
	panicMode := flags.String("panic", "trace", panicModeUsage)
	sanitize := sanitizeFlag(flags, "sanitize", sanitizeUsage)
	trackSource := flags.Bool("track-source", false, trackSourceUsage)
	if err := flags.Parse(args); err != nil {
		return err
	}

	pkg := "."
	if flags.NArg() > 0 {
		pkg = flags.Arg(0)
	}

	out := *outFile
	if out == "" {
		absDir, err := filepath.Abs(pkg)
		if err != nil {
			return err
		}
		out = filepath.Base(absDir)
	}

	opts := compiler.Options{
		Assert:      *assert,
		PanicMode:   *panicMode,
		Sanitize:    sanitize.list,
		TrackSource: *trackSource,
	}
	return compiler.Build(pkg, out, opts)
}

func test(args []string) error {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	run := flags.String("run", "", "run only tests whose names start with this prefix")
	assert := flags.String("assert", "on", assertUsage)
	panicMode := flags.String("panic", "trace", panicModeUsage)
	sanitize := sanitizeFlag(flags, "sanitize", sanitizeUsage)
	trackSource := flags.Bool("track-source", false, trackSourceUsage)
	if err := flags.Parse(args); err != nil {
		return err
	}

	pkg := "."
	if flags.NArg() > 0 {
		pkg = flags.Arg(0)
	}

	// Forward the test-related options to the compiled runner.
	var runArgs []string
	if *run != "" {
		runArgs = []string{"-run=" + *run}
	}

	opts := compiler.Options{
		Assert:      *assert,
		PanicMode:   *panicMode,
		Sanitize:    sanitize.list,
		TrackSource: *trackSource,
	}
	return compiler.Test(pkg, runArgs, opts)
}

func bench(args []string) error {
	flags := flag.NewFlagSet("bench", flag.ContinueOnError)
	run := flags.String("run", "", "run only benchmarks whose names start with this prefix")
	assert := flags.String("assert", "on", assertUsage)
	panicMode := flags.String("panic", "trace", panicModeUsage)
	sanitize := sanitizeFlag(flags, "sanitize", sanitizeUsage)
	trackSource := flags.Bool("track-source", false, trackSourceUsage)
	if err := flags.Parse(args); err != nil {
		return err
	}

	pkg := "."
	if flags.NArg() > 0 {
		pkg = flags.Arg(0)
	}

	// Forward the bench-related options to the compiled runner.
	var runArgs []string
	if *run != "" {
		runArgs = []string{"-run=" + *run}
	}

	opts := compiler.Options{
		Assert:      *assert,
		PanicMode:   *panicMode,
		Sanitize:    sanitize.list,
		TrackSource: *trackSource,
	}
	return compiler.Bench(pkg, runArgs, opts)
}

func run(args []string) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	assert := flags.String("assert", "on", assertUsage)
	panicMode := flags.String("panic", "trace", panicModeUsage)
	sanitize := sanitizeFlag(flags, "sanitize", sanitizeUsage)
	trackSource := flags.Bool("track-source", false, trackSourceUsage)
	if err := flags.Parse(args); err != nil {
		return err
	}

	pkg := "."
	var runArgs []string
	if flags.NArg() > 0 {
		pkg = flags.Arg(0)
		runArgs = flags.Args()[1:]
	}

	opts := compiler.Options{
		Assert:      *assert,
		PanicMode:   *panicMode,
		Sanitize:    sanitize.list,
		TrackSource: *trackSource,
	}
	return compiler.Run(pkg, runArgs, opts)
}

// sanitizeValue is the flag.Value for -sanitize. Bare -sanitize enables the
// default set; -sanitize=address,undefined enables a specific list. The zero
// value (flag absent) enables no sanitizers.
type sanitizeValue struct{ list string }

func sanitizeFlag(flags *flag.FlagSet, name, usage string) sanitizeValue {
	var s sanitizeValue
	flags.Var(&s, name, usage)
	return s
}

func (s *sanitizeValue) String() string { return s.list }

func (s *sanitizeValue) Set(v string) error {
	if v == "" || v == "true" {
		s.list = "address,undefined"
	} else {
		s.list = v
	}
	return nil
}

// IsBoolFlag lets -sanitize be given without a value, defaulting to the
// standard set, while still accepting -sanitize=list.
func (s *sanitizeValue) IsBoolFlag() bool { return true }
