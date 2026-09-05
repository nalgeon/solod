package clang

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// EmitOptions holds the options for code generation.
type EmitOptions struct {
	Pkg         *packages.Package
	OutDir      string
	InitArgs    bool // main takes argc/argv and initializes os.Args
	TrackSource bool // track source locations for panics
}

// EmitResult holds information produced by Emit for later pipeline steps.
type EmitResult struct {
	// Header and Impl hold the generated .h and .c file contents.
	Header []byte
	Impl   []byte
	// Libs lists the C libraries the package must link against, taken from
	// its so:link directives. Names are given without the -l prefix.
	Libs []string
}

// Emit generates C code for the given Go package and all its subpackages,
// and writes it to the specified output directory. Creates a single header
// file with typedefs (.h) and a single implementation file (.c) for each package.
func Emit(opts EmitOptions) (EmitResult, error) {
	res, err := Generate(opts)
	if err != nil {
		return res, err
	}
	err = Write(opts.OutDir, opts.Pkg.Name, res.Header, res.Impl)
	return res, err
}

// Generate generates C code for the given Go package and returns the contents
// of the header file with typedefs (.h) and the implementation file (.c).
func Generate(opts EmitOptions) (res EmitResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			f, ok := r.(*failure)
			if !ok {
				panic(r) // not a diagnostic: a real bug, keep crashing
			}
			err = f
		}
	}()

	// Initialize the generator with package information.
	g := newGenerator(opts)
	g.collect()

	var header, impl bytes.Buffer
	g.emitHeader(&header)
	g.emitImpl(&impl)

	res.Header = header.Bytes()
	res.Impl = impl.Bytes()
	res.Libs = g.links
	return res, nil
}

// Write writes the generated header and implementation
// of the package named pkgName to outDir.
func Write(outDir, pkgName string, header, impl []byte) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", outDir, err)
	}
	hPath := filepath.Join(outDir, pkgName+".h")
	if err := os.WriteFile(hPath, header, 0o644); err != nil {
		return fmt.Errorf("write header file %s: %w", hPath, err)
	}
	cPath := filepath.Join(outDir, pkgName+".c")
	if err := os.WriteFile(cPath, impl, 0o644); err != nil {
		return fmt.Errorf("write C file %s: %w", cPath, err)
	}
	return nil
}

// Includes holds the C headers to be included in the emitted .h and .c files.
type Includes struct {
	header []string // so:include -> emitted in .h
	impl   []string // so:include.c -> emitted in .c
}

// Generator is responsible for generating C code from Go ASTs.
type Generator struct {
	opts        EmitOptions
	pkg         *packages.Package
	types       *types.Info
	state       State
	modulePkgs  map[string]bool              // import paths of the packages inside a module
	externs     map[types.Object]externInfo  // symbols provided by C headers
	promoted    map[types.Object]bool        // unexported symbols forced into the header
	implObjs    map[types.Object]bool        // symbols only the .c file declares
	renames     map[types.Object]string      // C names changed to avoid name conflicts
	fieldNames  map[*types.Var]string        // C name overrides from `c:"..."` field tags
	includes    Includes                     // included headers from so:include
	links       []string                     // link libraries from so:link
	embeds      Embeds                       // embedded C files from so:embed
	symbols     []symbol                     // pre-collected top-level declarations
	funcDirs    map[*ast.FuncDecl]directives // parsed directives per function decl
	resultTypes []resultTypeInfo             // auto-generated result types for (T, error)
	comments    ast.CommentMap               // all comments across all files
	initFunc    *ast.FuncDecl                // package init() function, if any
	panicked    bool                         // true after first panic caught in Visit
}

// newGenerator creates a new Generator instance.
func newGenerator(opts EmitOptions) *Generator {
	return &Generator{
		opts:       opts,
		pkg:        opts.Pkg,
		types:      opts.Pkg.TypesInfo,
		modulePkgs: make(map[string]bool),
		externs:    make(map[types.Object]externInfo),
		renames:    make(map[types.Object]string),
		fieldNames: make(map[*types.Var]string),
		funcDirs:   make(map[*ast.FuncDecl]directives),
	}
}

// emitHeader creates the .h file with typedefs, includes, and extern declarations.
func (g *Generator) emitHeader(w io.Writer) {
	fmt.Fprint(w, "#pragma once\n")
	fmt.Fprintf(w, "#include \"so/builtin/builtin.h\"\n")
	for _, inc := range g.includes.header {
		fmt.Fprintf(w, "#include %s\n", inc)
	}
	g.emitImports(w)
	g.emitEmbeds(w, g.embeds.header)
	g.emitHeaderDecls(w)
}

// emitImpl creates the .c implementation file by walking the AST.
func (g *Generator) emitImpl(w io.Writer) {
	fmt.Fprintf(w, "#include \"%s.h\"\n", g.pkg.Name)
	for _, inc := range g.includes.impl {
		fmt.Fprintf(w, "#include %s\n", inc)
	}

	g.emitEmbeds(w, g.embeds.impl)
	g.emitUnexportedTypes(w)
	g.emitResultTypes(w, false)
	// Forward func decls before package vars: a package-var initializer may
	// reference a function (e.g. a func-typed struct field), and a function
	// prototype never references a package var, so this order satisfies both.
	g.emitForwardFuncDecls(w)
	g.emitPackageVars(w)

	multiFile := len(g.pkg.Syntax) > 1
	if !multiFile {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "// -- Implementation --")
	}
	for _, file := range g.pkg.Syntax {
		if multiFile {
			pos := g.pkg.Fset.Position(file.Pos())
			fmt.Fprintf(w, "\n// -- %s --\n", filepath.Base(pos.Filename))
		}
		g.walkAST(w, file)
	}
	g.emitInitFunc(w)
}

// emitEmbeds writes the content of embedded files, separated by blank lines.
func (g *Generator) emitEmbeds(w io.Writer, files []embedFile) {
	if len(files) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// -- Embeds --")
	for _, ef := range files {
		fmt.Fprintf(w, "\n%s\n", strings.TrimRight(ef.content, "\n"))
	}
}

// emitInitFunc emits the package init() function as a GCC constructor
// that runs automatically before main().
func (g *Generator) emitInitFunc(w io.Writer) {
	if g.initFunc == nil {
		return
	}
	decl := g.initFunc
	g.state.enterFunc(decl, g.funcSig(decl))
	defer g.state.leaveFunc()

	fmt.Fprintf(w, "\nstatic void __attribute__((constructor)) %s_init() {\n", g.pkg.Name)
	g.state.depth++
	g.walkStmts(w, decl.Body.List)
	g.emitDeferredCalls(w)
	g.state.depth--
	fmt.Fprint(w, "}\n")
}

// indent is a shorthand for the current scope's indentation.
func (g *Generator) indent() string {
	return g.state.indent()
}
