package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Build translates the Go package in srcDir to C and compiles it into outFile.
// Creates the directory of outFile when the directory does not exist.
// Uses CC (default "cc"), CFLAGS, and LDFLAGS environment variables.
func Build(srcDir, outFile string, opts Options) error {
	return build(dirSource(srcDir), outFile, opts)
}

// build is Build for an arbitrary source.
func build(src source, outFile string, opts Options) error {
	copts, err := newCompileOptions(opts)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "solod_build")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	libs, err := translate(src, tmpDir, opts)
	if err != nil {
		return err
	}

	cFiles, err := findCFiles(tmpDir)
	if err != nil {
		return err
	}

	// Create the the output directory to match "go build" behavior.
	if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	copts.libs = libs
	return compileC(tmpDir, cFiles, outFile, copts)
}

// Run translates and compiles the Go package in srcDir, then executes it.
// Returns an *exec.ExitError if the program exits with a non-zero status.
func Run(srcDir string, args []string, opts Options) error {
	return run(dirSource(srcDir), args, opts)
}

// run is Run for an arbitrary source.
func run(src source, args []string, opts Options) error {
	tmpFile, err := os.CreateTemp("", "solod_run")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := build(src, tmpFile.Name(), opts); err != nil {
		return err
	}

	cmd := exec.Command(tmpFile.Name(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Version returns the compiler version string to embed into compiled
// programs via -Dso_version. It uses the module version from
// runtime/debug.BuildInfo when available (e.g. go install ...@vx.y.z),
// falling back to "(devel)" (e.g. go run during development).
func Version() string {
	const devel = "(devel)"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return devel
	}
	if v := info.Main.Version; v != "" {
		return v
	}
	return devel
}

// findCFiles returns all .c files under dir, recursively.
func findCFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".c") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find C files: %w", err)
	}
	return files, nil
}

// compileOptions holds the extra C preprocessor defines and
// compiler flags derived from the compiler Options.
type compileOptions struct {
	defines []string // preprocessor -D flags
	flags   []string // additional C compiler flags
	libs    []string // libraries to link (without -l)
}

// newCompileOptions derives the C defines and flags from opts.
func newCompileOptions(opts Options) (compileOptions, error) {
	panicDef, panicFlags, err := panicMode(opts.PanicMode, ccName(), opts.Freestanding)
	if err != nil {
		return compileOptions{}, err
	}
	assertDefs, err := assertMode(opts.Assert)
	if err != nil {
		return compileOptions{}, err
	}
	flags := append(panicFlags, sanitizeFlags(opts.Sanitize)...)
	if opts.Freestanding {
		// builtin.h derives so_build_hosted from __STDC_HOSTED__, which
		// -ffreestanding sets to 0. The flag comes before CFLAGS, so the
		// caller can still override it.
		flags = append(flags, "-ffreestanding")
	}
	defines := append([]string{panicDef}, assertDefs...)
	return compileOptions{
		defines: defines,
		flags:   flags,
	}, nil
}

// compileC invokes the C compiler to produce an executable.
func compileC(includeDir string, cFiles []string, outFile string, copts compileOptions) error {
	cc := ccName()

	args := []string{"-I" + includeDir}
	args = append(args, fmt.Sprintf(`-Dso_version="%s"`, Version()))
	args = append(args, copts.defines...)
	args = append(args, copts.flags...)
	args = append(args, splitFlags(os.Getenv("CFLAGS"))...)
	args = append(args, cFiles...)
	args = append(args, "-o", outFile)
	// Link libraries the packages declared, then any user LDFLAGS. Both come
	// after the object files so the linker resolves their symbols.
	for _, lib := range copts.libs {
		args = append(args, "-l"+lib)
	}
	args = append(args, splitFlags(os.Getenv("LDFLAGS"))...)

	cmd := exec.Command(cc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler: %w", err)
	}
	return nil
}

// panicMode maps a panic mode name to the -DSO_PANIC_MODE define and any
// extra C compiler flags the mode needs, given the C compiler cc.
// An empty mode defaults to "trace".
func panicMode(mode, cc string, freestanding bool) (define string, flags []string, err error) {
	switch mode {
	case "", "trace":
		// A freestanding build traps whatever the mode is, so it needs no flags.
		if freestanding {
			return "-DSO_PANIC_MODE=SO_PANIC_TRACE", nil, nil
		}
		// Needs frame pointers to unwind and -rdynamic for symbol names.
		// MinGW rejects -rdynamic, so we skip it there.
		flags := []string{"-fno-omit-frame-pointer"}
		if !isMinGW(cc) {
			flags = append(flags, "-rdynamic")
		}
		return "-DSO_PANIC_MODE=SO_PANIC_TRACE", flags, nil
	case "exit":
		return "-DSO_PANIC_MODE=SO_PANIC_EXIT", nil, nil
	case "abort":
		return "-DSO_PANIC_MODE=SO_PANIC_ABORT", nil, nil
	default:
		return "", nil, fmt.Errorf("invalid panic mode %q (want exit, abort, or trace)", mode)
	}
}

// assertMode maps an assert mode name to the C defines the mode needs.
// An empty mode defaults to "on", which needs no defines.
func assertMode(mode string) (defines []string, err error) {
	switch mode {
	case "", "on":
		return nil, nil
	case "off":
		return []string{"-DSO_NO_ASSERT"}, nil
	default:
		return nil, fmt.Errorf("invalid assert mode %q (want on or off)", mode)
	}
}

// sanitizeFlags returns the C compiler flags for the requested sanitizers.
// list is a comma-separated set (e.g. "address,undefined"); empty enables none.
func sanitizeFlags(list string) []string {
	names := splitList(list)
	if len(names) == 0 {
		return nil
	}
	// Add these so reports carry readable file:line stack traces.
	flags := []string{"-g", "-fno-omit-frame-pointer"}
	for _, name := range names {
		flags = append(flags, "-fsanitize="+name)
	}
	return flags
}

// splitFlags splits a space-separated flags string into individual args.
func splitFlags(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}

// splitList splits a comma-separated string into trimmed, non-empty items.
func splitList(s string) []string {
	var items []string
	for item := range strings.SplitSeq(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// ccName returns the C compiler to invoke, from CC (default "cc").
func ccName() string {
	if cc := os.Getenv("CC"); cc != "" {
		return cc
	}
	return "cc"
}

// isMinGW reports whether cc names a MinGW compiler, which targets Windows.
func isMinGW(cc string) bool {
	return strings.Contains(strings.ToLower(cc), "mingw")
}
