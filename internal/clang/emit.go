package clang

import (
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
	Pkg    *packages.Package
	OutDir string
}

// Emit generates C code for the given Go package and all its subpackages,
// and writes it to the specified output directory. Creates a single header
// file with typedefs (.h) and a single implementation file (.c) for each package.
func Emit(opts EmitOptions) error {
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", opts.OutDir, err)
	}
	g := newGenerator(opts.Pkg)
	g.collectExterns()
	g.collectSymbols()
	if err := g.emitHeader(opts.OutDir); err != nil {
		return err
	}
	return g.emitImpl(opts.OutDir)
}

// State holds the code generation state for the current scope.
type State struct {
	writer io.Writer

	// Current indentation level (number of tabs).
	indent int
	// Current receiver name (for -> access in methods).
	recvName string
	// Current function's signature (for multi-return).
	funcSig *types.Signature
	// Counter for unique temp variable names.
	tempCount int
}

// Generator is responsible for generating C code from Go ASTs.
type Generator struct {
	pkg      *packages.Package
	types    *types.Info
	state    State
	externs  map[string]bool // symbols provided by C headers
	includes []string        // #include directives from comments
	symbols  []symbol        // pre-collected top-level declarations
	panicked bool            // true after first panic caught in Visit
}

// newGenerator creates a new Generator instance.
func newGenerator(pkg *packages.Package) *Generator {
	return &Generator{
		pkg:     pkg,
		types:   pkg.TypesInfo,
		externs: make(map[string]bool),
	}
}

// emitHeader creates the .h file with typedefs, includes, and extern declarations.
func (g *Generator) emitHeader(dir string) error {
	hName := g.pkg.Name + ".h"
	hPath := filepath.Join(dir, hName)
	hFile, err := os.Create(hPath)
	if err != nil {
		return fmt.Errorf("create header file %s: %w", hPath, err)
	}
	defer hFile.Close()
	fmt.Fprintf(hFile, "#pragma once\n")
	fmt.Fprintf(hFile, "#include \"solod.h\"\n")
	g.emitImports(hFile)
	g.emitHeaderDecls(hFile)
	return nil
}

// emitImpl creates the .c implementation file by walking the AST.
func (g *Generator) emitImpl(dir string) error {
	cName := g.pkg.Name + ".c"
	cPath := filepath.Join(dir, cName)
	cFile, err := os.Create(cPath)
	if err != nil {
		return fmt.Errorf("create C file %s: %w", cPath, err)
	}
	defer cFile.Close()
	fmt.Fprintf(cFile, "#include \"%s.h\"\n", g.pkg.Name)
	// Emit additional #include directives collected from comments.
	for _, inc := range g.includes {
		fmt.Fprintf(cFile, "%s\n", inc)
	}
	g.state.writer = cFile
	g.emitForwardDecls(cFile)
	for _, file := range g.pkg.Syntax {
		ast.Walk(g, file)
	}
	return nil
}

// indent returns the current indentation string based on the indent level.
func (g *Generator) indent() string {
	return strings.Repeat("    ", g.state.indent)
}
