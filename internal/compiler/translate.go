package compiler

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/nalgeon/soan/internal/clang"
)

// Translate loads all Go packages from srcDir (including Soan stdlib dependencies),
// translates them to C, and writes the output to outDir.
func Translate(srcDir string, outDir string) error {
	pkgs, err := loadPackages(srcDir)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found")
	}

	entry := pkgs[0]

	var entryModulePath string
	if entry.Module != nil {
		entryModulePath = entry.Module.Path
	}

	var soanModulePath string
	if info, ok := debug.ReadBuildInfo(); ok {
		soanModulePath = info.Main.Path
	}

	// Walk import graph and collect transpilable packages in topological order
	ordered := topoSort(entry, entryModulePath, soanModulePath)

	// Translate each package
	for _, pkg := range ordered {
		pkgOutDir := packageOutDir(pkg, entry, outDir)
		if err := clang.Emit(clang.EmitOptions{
			Pkg:    pkg,
			OutDir: pkgOutDir,
		}); err != nil {
			return err
		}
	}

	// Write embedded builtin files (so.h, so.c) into the output directory
	return writeBuiltin(outDir)
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
// (module-internal + Soan stdlib) in topological order (dependencies before dependents).
func topoSort(entry *packages.Package, entryModulePath, soanModulePath string) []*packages.Package {
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
			if shouldTranspile(dep, entryModulePath, soanModulePath) {
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
// Other packages strip their module prefix (e.g. github.com/nalgeon/soan/math -> math).
func packageOutDir(pkg, entry *packages.Package, outDir string) string {
	if pkg.PkgPath == entry.PkgPath {
		return outDir
	}
	relPath := strings.TrimPrefix(pkg.PkgPath, pkg.Module.Path+"/")
	return filepath.Join(outDir, relPath)
}

// shouldTranspile returns true if a package should be transpiled to C.
// This includes packages from the entry module and Soan stdlib packages.
func shouldTranspile(pkg *packages.Package, entryModulePath, soanModulePath string) bool {
	if pkg.Module == nil {
		return false
	}
	return pkg.Module.Path == entryModulePath || pkg.Module.Path == soanModulePath
}
