package compiler

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"

	"solod.dev/internal/clang"
)

// Options holds the options for the compiler pipeline.
type Options struct {
	Assert      string // assertions: "on" (default) or "off"
	PanicMode   string // panic termination mode: "trace" (default), "exit", or "abort"
	Sanitize    string // comma-separated sanitizers to enable; empty disables
	TrackSource bool   // track source locations for panics
}

// Translate loads all Go packages from srcDir (including So stdlib dependencies),
// translates them to C, and writes the output to outDir. It returns the C
// libraries the transpiled packages must link against, deduplicated and sorted,
// without the -l prefix.
func Translate(srcDir, outDir string, opts Options) ([]string, error) {
	pkgs, err := loadPackages(srcDir)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	// Walk import graph and collect transpilable packages in topological order
	entry := pkgs[0]
	ordered := topoSort(entry)

	// Translate each package, collecting the union of their link libraries.
	libSet := make(map[string]bool)
	for _, pkg := range ordered {
		pkgOutDir := packageOutDir(pkg, entry, outDir)
		res, err := clang.Emit(clang.EmitOptions{
			Pkg:         pkg,
			OutDir:      pkgOutDir,
			TrackSource: opts.TrackSource,
		})
		if err != nil {
			return nil, err
		}
		for _, lib := range res.Libs {
			libSet[lib] = true
		}
	}

	// Write embedded builtin files into the output directory
	builtinDir := filepath.Join(outDir, "so", "builtin")
	if err := os.MkdirAll(builtinDir, 0o755); err != nil {
		return nil, fmt.Errorf("create builtin output directory %s: %w", builtinDir, err)
	}
	if err := writeBuiltin(builtinDir); err != nil {
		return nil, err
	}

	return slices.Sorted(maps.Keys(libSet)), nil
}

// loadPackages uses go/packages to load the entry package and all dependencies.
func loadPackages(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedImports | packages.NeedDeps |
			packages.NeedModule | packages.NeedTypesInfo,
		Dir: dir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}
	return pkgs, nil
}

// topoSort walks the import graph from entry and returns transpilable packages
// in topological order (dependencies before dependents).
func topoSort(entry *packages.Package) []*packages.Package {
	var ordered []*packages.Package
	visited := make(map[string]bool)

	var walk func(pkg *packages.Package)
	walk = func(pkg *packages.Package) {
		if visited[pkg.PkgPath] {
			return
		}
		visited[pkg.PkgPath] = true

		// Visit dependencies first (post-order)
		for _, dep := range pkg.Imports {
			if shouldTranspile(dep) {
				walk(dep)
			}
		}
		ordered = append(ordered, pkg)
	}
	walk(entry)
	return ordered
}

// packageOutDir returns the output directory for a package.
// Entry package goes to outDir directly.
// Other packages strip their module prefix (e.g. solod.dev/math -> math).
func packageOutDir(pkg, entry *packages.Package, outDir string) string {
	if pkg.PkgPath == entry.PkgPath {
		return outDir
	}
	relPath := strings.TrimPrefix(pkg.PkgPath, pkg.Module.Path+"/")
	return filepath.Join(outDir, relPath)
}

// shouldTranspile returns true if a package should be transpiled to C.
// Go standard library packages (Module == nil) are skipped;
// everything else (user code, So stdlib, third-party So packages) is transpiled.
func shouldTranspile(pkg *packages.Package) bool {
	return pkg.Module != nil
}
