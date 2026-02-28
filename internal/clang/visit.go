package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

// Visit implements the ast.Visitor interface to walk the AST and generate code.
func (g *Generator) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.AssignStmt:
		g.emitAssignStmt(n)
		return nil
	case *ast.BranchStmt:
		g.emitBranchStmt(n)
		return nil
	case *ast.DeferStmt:
		g.emitDeferStmt(n)
		return nil
	case *ast.ExprStmt:
		g.emitExprStmt(n)
		return nil
	case *ast.ForStmt:
		g.emitForStmt(n)
		return nil
	case *ast.FuncDecl:
		g.emitFuncDecl(n)
		return nil
	case *ast.IfStmt:
		g.emitIfStmt(n)
		return nil
	case *ast.IncDecStmt:
		g.emitIncDecStmt(n)
		return nil
	case *ast.GenDecl:
		g.emitGenDecl(n)
		return nil
	case *ast.RangeStmt:
		g.emitRangeStmt(n)
		return nil
	case *ast.ReturnStmt:
		g.emitReturnStmt(n)
		return nil
	case *ast.BlockStmt:
		// Bare block (scoping block inside a function body).
		fmt.Fprintf(g.state.writer, "%s{\n", g.indent())
		g.emitBlock(n)
		fmt.Fprintf(g.state.writer, "%s}\n", g.indent())
		return nil
	}

	return g
}

// emitBranchStmt emits a break or continue statement.
func (g *Generator) emitBranchStmt(stmt *ast.BranchStmt) {
	fmt.Fprintf(g.state.writer, "%s%s;\n", g.indent(), stmt.Tok)
}

// emitDeferStmt emits a defer statement as so_defer(fn, arg).
func (g *Generator) emitDeferStmt(stmt *ast.DeferStmt) {
	w := g.state.writer
	call := stmt.Call
	if len(call.Args) != 1 {
		g.fail(stmt, "defer call must have exactly 1 argument")
	}
	fmt.Fprintf(w, "%sso_defer(", g.indent())
	g.emitExpr(call.Fun)
	fmt.Fprintf(w, ", ")
	g.emitExpr(call.Args[0])
	fmt.Fprintf(w, ");\n")
}

// emitExprStmt emits an expression statement.
func (g *Generator) emitExprStmt(stmt *ast.ExprStmt) {
	w := g.state.writer
	fmt.Fprintf(w, "%s", g.indent())
	g.emitExpr(stmt.X)
	fmt.Fprintf(w, ";\n")
}

// emitForStmt emits a for statement.
func (g *Generator) emitForStmt(stmt *ast.ForStmt) {
	w := g.state.writer
	fmt.Fprintf(w, "%sfor (", g.indent())

	if stmt.Init != nil {
		assign := stmt.Init.(*ast.AssignStmt)
		ident := assign.Lhs[0].(*ast.Ident)
		cType := g.mapType(stmt, g.types.Defs[ident].Type())
		fmt.Fprintf(w, "%s %s = ", cType, ident.Name)
		g.emitExpr(assign.Rhs[0])
	}
	fmt.Fprintf(w, ";")

	if stmt.Cond != nil {
		fmt.Fprintf(w, " ")
		g.emitExpr(stmt.Cond)
	}
	fmt.Fprintf(w, ";")

	if stmt.Post != nil {
		fmt.Fprintf(w, " ")
		inc := stmt.Post.(*ast.IncDecStmt)
		g.emitExpr(inc.X)
		fmt.Fprintf(w, "%s", inc.Tok.String())
	}

	fmt.Fprintf(w, ") {\n")
	g.emitBlock(stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitGenDecl emits a general declaration (var, import, etc.).
func (g *Generator) emitGenDecl(decl *ast.GenDecl) {
	if hasExternDirective(decl.Doc) {
		return
	}
	switch decl.Tok {
	case token.IMPORT:
		for _, spec := range decl.Specs {
			g.emitImportSpec(spec.(*ast.ImportSpec))
		}
		return
	case token.CONST:
		for _, spec := range decl.Specs {
			g.emitConstSpec(spec.(*ast.ValueSpec))
		}
	case token.VAR:
		for _, spec := range decl.Specs {
			g.emitVarSpec(spec.(*ast.ValueSpec))
		}
	case token.TYPE:
		// Exported types are emitted in the header.
		// Unexported types are emitted here in the .c file.
		for _, spec := range decl.Specs {
			ts := spec.(*ast.TypeSpec)
			if !ast.IsExported(ts.Name.Name) {
				g.emitTypeSpec(g.state.writer, ts)
			}
		}
	default:
		g.fail(decl, "unsupported GenDecl token: %s", decl.Tok)
	}
}

// emitImportSpec emits a #include directive for an import.
func (g *Generator) emitImportSpec(spec *ast.ImportSpec) {
	path := strings.Trim(spec.Path.Value, `"`)
	if path == "errors" {
		// Errors are handled as builtins, so we have to skip the stdlib import.
		return
	}
	// Strip the imported package's own module prefix.
	if imp, ok := g.pkg.Imports[path]; ok && imp.Module != nil {
		path = strings.TrimPrefix(path, imp.Module.Path+"/")
	}
	// Add the package.h file (e.g. package -> package/package.h).
	parts := strings.Split(path, "/")
	parts = append(parts, parts[len(parts)-1]+".h")
	cPath := strings.Join(parts, "/")
	// Emit the #include directive (e.g. #include "package/package.h").
	fmt.Fprintf(g.state.writer, "#include \"%s\"\n", cPath)
}

// emitConstSpec emits a single constant specification.
func (g *Generator) emitConstSpec(spec *ast.ValueSpec) {
	w := g.state.writer
	for i, name := range spec.Names {
		typ := g.types.Defs[name].Type()
		cType := g.mapType(spec, typ)
		if g.state.indent == 0 {
			// Package-level constant.
			specifier := "static "
			if ast.IsExported(name.Name) {
				specifier = ""
			}
			fmt.Fprintf(w, "%sconst %s %s = ", specifier, cType, g.symbolName(name.Name))
			g.emitExpr(spec.Values[i])
			fmt.Fprintf(w, ";\n")
		} else {
			// Local constant (e.g. inside a function).
			fmt.Fprintf(w, "%sconst %s %s = ", g.indent(), cType, name.Name)
			g.emitExpr(spec.Values[i])
			fmt.Fprintf(w, ";\n")
		}
	}
}

// emitVarSpec emits a single var specification (e.g. `var a int = 1`).
func (g *Generator) emitVarSpec(spec *ast.ValueSpec) {
	w := g.state.writer

	// Local multi-variable declaration: group consecutive same-type variables,
	// but emit separate declarations for different types
	// (e.g. `int a = 1, b = 2; float c = 3.14;`).
	if g.state.indent > 0 && len(spec.Names) > 1 {
		i := 0
		for i < len(spec.Names) {
			name := spec.Names[i]
			typ := g.types.Defs[name].Type()
			cType := g.mapType(spec, typ)
			fmt.Fprintf(w, "%s%s %s = ", g.indent(), cType, name.Name)
			if len(spec.Values) > i {
				g.emitExpr(spec.Values[i])
			} else {
				fmt.Fprintf(w, "%s", g.zeroValue(spec, typ))
			}
			i++
			for i < len(spec.Names) {
				nextName := spec.Names[i]
				nextTyp := g.types.Defs[nextName].Type()
				nextCType := g.mapType(spec, nextTyp)
				if nextCType != cType {
					break
				}
				fmt.Fprintf(w, ", %s = ", nextName.Name)
				if len(spec.Values) > i {
					g.emitExpr(spec.Values[i])
				} else {
					fmt.Fprintf(w, "%s", g.zeroValue(spec, nextTyp))
				}
				i++
			}
			fmt.Fprintf(w, ";\n")
		}
		return
	}

	// Single variable or package-level declaration.
	for i, name := range spec.Names {
		typ := g.types.Defs[name].Type()
		cType := g.mapType(spec, typ)
		specifier := ""
		if g.state.indent == 0 {
			// Package-level variable.
			if !ast.IsExported(name.Name) {
				specifier = "static "
			}
		}
		cName := g.symbolName(name.Name)
		if len(spec.Values) > i {
			// Has explicit initializer.
			fmt.Fprintf(w, "%s%s%s %s = ", g.indent(), specifier, cType, cName)
			if iface, ok := typ.Underlying().(*types.Interface); ok && iface.Empty() {
				g.emitAnyValue(spec, spec.Values[i])
			} else {
				g.emitExpr(spec.Values[i])
			}
			fmt.Fprintf(w, ";\n")
		} else {
			// No initializer, emit zero value.
			zeroVal := g.zeroValue(spec, typ)
			fmt.Fprintf(w, "%s%s%s %s = %s;\n", g.indent(), specifier, cType, cName, zeroVal)
		}
	}
}

// emitTypeSpec dispatches type declaration emission based on the spec type.
func (g *Generator) emitTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	switch spec.Type.(type) {
	case *ast.FuncType:
		g.emitFuncTypeSpec(w, spec)
	case *ast.Ident:
		typ := g.types.Defs[spec.Name].Type()
		cType := g.mapType(spec, typ.Underlying())
		cName := g.symbolName(spec.Name.Name)
		fmt.Fprintf(w, "typedef %s %s;\n", cType, cName)
	case *ast.InterfaceType:
		g.emitInterfaceTypeSpec(w, spec)
	case *ast.StructType:
		g.emitStructTypeSpec(w, spec)
	default:
		g.fail(spec, "unsupported type: %T", spec.Type)
	}
}

// emitIfStmt emits an if statement, wrapping in a scope block if there's an init statement.
func (g *Generator) emitIfStmt(stmt *ast.IfStmt) {
	w := g.state.writer
	if stmt.Init != nil {
		fmt.Fprintf(w, "%s{\n", g.indent())
		g.state.indent++
		ast.Walk(g, stmt.Init)
		g.emitIfInner(w, stmt, g.indent())
		g.state.indent--
		fmt.Fprintf(w, "%s}\n", g.indent())
	} else {
		g.emitIfInner(w, stmt, g.indent())
	}
}

// emitIfInner emits the if/else-if/else chain. The prefix controls leading
// indentation: top-level calls pass g.indent(), recursive else-if calls pass "".
func (g *Generator) emitIfInner(w io.Writer, stmt *ast.IfStmt, prefix string) {
	// Emit the if condition and body.
	fmt.Fprintf(w, "%sif (", prefix)
	g.emitExpr(stmt.Cond)
	fmt.Fprintf(w, ") {\n")
	g.emitBlock(stmt.Body)
	if stmt.Else == nil {
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	// Handle else-if and else clauses.
	switch els := stmt.Else.(type) {
	case *ast.IfStmt:
		fmt.Fprintf(w, "%s} else ", g.indent())
		g.emitIfInner(w, els, "")
	case *ast.BlockStmt:
		fmt.Fprintf(w, "%s} else {\n", g.indent())
		g.emitBlock(els)
		fmt.Fprintf(w, "%s}\n", g.indent())
	default:
		g.fail(stmt.Else, "unsupported else clause: %T", stmt.Else)
	}
}

// emitIncDecStmt emits an increment or decrement statement.
func (g *Generator) emitIncDecStmt(stmt *ast.IncDecStmt) {
	w := g.state.writer
	fmt.Fprintf(w, "%s", g.indent())
	g.emitExpr(stmt.X)
	fmt.Fprintf(w, "%s;\n", stmt.Tok)
}

// emitRangeStmt emits a range-based for statement.
func (g *Generator) emitRangeStmt(stmt *ast.RangeStmt) {
	switch t := g.types.TypeOf(stmt.X).Underlying().(type) {
	case *types.Array, *types.Slice:
		g.emitSliceRange(stmt)
	case *types.Basic:
		if t.Kind() == types.String {
			g.emitStringRange(stmt)
		} else {
			g.emitIntRange(stmt)
		}
	default:
		g.emitIntRange(stmt)
	}
}

// emitReturnStmt emits a return statement.
func (g *Generator) emitReturnStmt(stmt *ast.ReturnStmt) {
	w := g.state.writer
	if len(stmt.Results) == 0 {
		fmt.Fprintf(w, "%sreturn;\n", g.indent())
		return
	}
	// Assign extra return values to out-parameters.
	if len(stmt.Results) > 1 && len(g.state.outParams) > 0 {
		for i := 1; i < len(stmt.Results); i++ {
			fmt.Fprintf(w, "%s*%s = ", g.indent(), g.state.outParams[i-1])
			g.emitExpr(stmt.Results[i])
			fmt.Fprintf(w, ";\n")
		}
	}
	fmt.Fprintf(w, "%sreturn ", g.indent())
	g.emitExpr(stmt.Results[0])
	fmt.Fprintf(w, ";\n")
}

// emitBlock emits the statements within a block, adjusting indentation.
func (g *Generator) emitBlock(block *ast.BlockStmt) {
	g.state.indent++
	for _, stmt := range block.List {
		ast.Walk(g, stmt)
	}
	g.state.indent--
}
