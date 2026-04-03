package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"
)

// Visit implements the ast.Visitor interface to walk the AST and generate code.
func (g *Generator) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			// Only log once - the deepest Visit that catches the panic.
			if !g.panicked {
				g.panicked = true
				pos := g.pkg.Fset.Position(node.Pos())
				fmt.Fprintf(os.Stderr, "%s: %v\n", pos, r)
				if srcLine, err := readSourceLine(pos.Filename, pos.Line); err == nil {
					fmt.Fprintf(os.Stderr, "%s\n", srcLine)
				}
			}
			panic(r)
		}
	}()

	switch n := node.(type) {
	case *ast.AssignStmt:
		g.emitAssignStmt(n)
		return nil
	case *ast.BlockStmt:
		g.emitBlockStmt(n)
		return nil
	case *ast.BranchStmt:
		g.emitBranchStmt(n)
		return nil
	case *ast.DeclStmt:
		// Declaration inside a function body -
		// walk into the inner Decl.
		return g
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
	case *ast.GenDecl:
		g.emitGenDecl(n)
		return nil
	case *ast.Ident:
		// Package name or other identifiers
		// visited during file walk.
		return g
	case *ast.IfStmt:
		g.emitIfStmt(n)
		return nil
	case *ast.IncDecStmt:
		g.emitIncDecStmt(n)
		return nil
	case *ast.LabeledStmt:
		g.emitLabeledStmt(n)
		return nil
	case *ast.RangeStmt:
		g.emitRangeStmt(n)
		return nil
	case *ast.ReturnStmt:
		g.emitReturnStmt(n)
		return nil
	case *ast.SwitchStmt:
		g.emitSwitchStmt(n)
		return nil
	}

	// Fail on unsupported expressions, statements, and declarations.
	switch node.(type) {
	case ast.Stmt:
		g.fail(node, "unsupported statement: %T", node)
	case ast.Decl:
		g.fail(node, "unsupported declaration: %T", node)
	case ast.Expr:
		g.fail(node, "unsupported expression: %T", node)
	}

	return g
}

// emitBlockStmt emits a bare block statement (scoping block inside a function body).
func (g *Generator) emitBlockStmt(stmt *ast.BlockStmt) {
	defers := g.state.defers // stash outer defers
	g.state.defers = nil
	fmt.Fprintf(g.state.writer, "%s{\n", g.indent())
	g.emitBlock(stmt)
	g.state.indent++
	g.emitDeferredCalls()
	g.state.indent--
	g.state.defers = defers // restore outer defers
	fmt.Fprintf(g.state.writer, "%s}\n", g.indent())
}

// emitBranchStmt emits a break, continue, or goto statement.
func (g *Generator) emitBranchStmt(stmt *ast.BranchStmt) {
	w := g.state.writer
	if stmt.Label != nil && stmt.Tok == token.BREAK {
		// Labeled break is translated to goto because C has no "break label".
		// ("break label" -> "goto label_end").
		fmt.Fprintf(w, "%sgoto %s_end;\n", g.indent(), stmt.Label.Name)
	} else if stmt.Label != nil {
		// Regular labeled goto, emit as-is.
		fmt.Fprintf(w, "%s%s %s;\n", g.indent(), stmt.Tok, stmt.Label.Name)
	} else {
		// Unlabeled break/continue.
		fmt.Fprintf(w, "%s%s;\n", g.indent(), stmt.Tok)
	}
}

// emitDeferStmt emits a defer statement. Deferred calls are captured
// and emitted inline before returns, panics, and function end.
func (g *Generator) emitDeferStmt(stmt *ast.DeferStmt) {
	var buf strings.Builder
	saved := g.state.writer
	g.state.writer = &buf
	g.emitCallExpr(stmt.Call)
	g.state.writer = saved
	g.state.defers = append(g.state.defers, buf.String())
}

// emitExprStmt emits an expression statement.
// Emits deferred generic calls before panic() calls.
func (g *Generator) emitExprStmt(stmt *ast.ExprStmt) {
	w := g.state.writer
	if g.isPanicCall(stmt.X) {
		g.emitDeferredCalls()
	}
	fmt.Fprintf(w, "%s", g.indent())
	g.emitExpr(stmt.X)
	fmt.Fprintf(w, ";\n")
}

// emitForStmt emits a for statement.
func (g *Generator) emitForStmt(stmt *ast.ForStmt) {
	w := g.state.writer
	fmt.Fprintf(w, "%sfor (", g.indent())

	if stmt.Init != nil {
		g.emitForClause(stmt.Init)
	}
	fmt.Fprintf(w, ";")

	if stmt.Cond != nil {
		fmt.Fprintf(w, " ")
		g.emitExpr(stmt.Cond)
	}
	fmt.Fprintf(w, ";")

	if stmt.Post != nil {
		fmt.Fprintf(w, " ")
		g.emitForClause(stmt.Post)
	}

	fmt.Fprintf(w, ") {\n")
	g.emitBlock(stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitForClause emits a simple statement inline (no indent, no semicolon)
// for use in for-loop Init and Post positions.
func (g *Generator) emitForClause(stmt ast.Stmt) {
	w := g.state.writer
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if s.Tok == token.DEFINE {
			ident := s.Lhs[0].(*ast.Ident)
			cType := g.mapType(s, g.types.Defs[ident].Type())
			fmt.Fprintf(w, "%s %s = ", cType, ident.Name)
			g.emitExpr(s.Rhs[0])
		} else {
			g.emitExpr(s.Lhs[0])
			fmt.Fprintf(w, " %s ", s.Tok)
			g.emitExpr(s.Rhs[0])
		}
	case *ast.IncDecStmt:
		g.emitExpr(s.X)
		fmt.Fprintf(w, "%s", s.Tok)
	case *ast.ExprStmt:
		g.emitExpr(s.X)
	default:
		g.fail(stmt, "unsupported for-loop clause: %T", stmt)
	}
}

// emitGenDecl emits a general declaration (var, import, etc.).
func (g *Generator) emitGenDecl(decl *ast.GenDecl) {
	if found, _ := parseExternDirective(decl.Doc); found {
		return
	}
	switch decl.Tok {
	case token.IMPORT:
		// Imports are handled separately at [Generator.emitImpl].
		return
	case token.CONST:
		if g.state.indent == 0 {
			// Package-level consts are hoisted by emitPackageVars.
			return
		}
		for _, spec := range decl.Specs {
			g.emitConstSpec(spec.(*ast.ValueSpec))
		}
	case token.VAR:
		if g.state.indent == 0 {
			// Package-level vars are hoisted by emitPackageVars.
			return
		}
		for _, spec := range decl.Specs {
			vs := spec.(*ast.ValueSpec)
			if len(vs.Names) > 0 && g.embeds.vars[vs.Names[0].Name] {
				// Do not emit variables that are used as markers for embedded files.
				continue
			}
			g.emitVarSpec(vs)
		}
	case token.TYPE:
		// Package-level types are emitted by emitUnexportedTypes (unexported)
		// or emitHeaderDecls (exported). Only emit inside function bodies.
		if g.state.indent == 0 {
			return
		}
		for _, spec := range decl.Specs {
			ts := spec.(*ast.TypeSpec)
			g.emitComments(g.state.writer, decl, ts)
			g.emitTypeSpec(g.state.writer, ts)
		}
	default:
		g.fail(decl, "unsupported GenDecl token: %s", decl.Tok)
	}
}

// emitConstSpec emits a single constant specification.
func (g *Generator) emitConstSpec(spec *ast.ValueSpec) {
	w := g.state.writer
	for i, name := range spec.Names {
		typ := g.types.Defs[name].Type()
		cType := g.mapType(spec, typ)

		// Check if this is an iota-based constant (implicit value or explicit iota usage).
		isIota := i >= len(spec.Values) || containsIota(spec.Values[i])

		// Determine constant specifier and name.
		specifier, constName := "", name.Name
		if g.state.indent == 0 {
			// Package-level constant.
			if !ast.IsExported(constName) {
				specifier = "static "
			}
			constName = g.symbolName(constName)
		}

		// Emit the constant declaration.
		fmt.Fprintf(w, "%s%sconst %s %s = ", g.indent(), specifier, cType, constName)
		if isIota {
			g.emitConstVal(spec, name)
		} else {
			g.emitExpr(spec.Values[i])
		}
		fmt.Fprintf(w, ";\n")
	}
}

// emitConstVal emits the type-checker-resolved value of a constant.
func (g *Generator) emitConstVal(node ast.Node, name *ast.Ident) {
	obj := g.types.Defs[name].(*types.Const)
	val := obj.Val()
	switch val.Kind() {
	case constant.Int:
		v, ok := constant.Int64Val(val)
		if !ok {
			g.fail(node, "iota value overflows int64")
		}
		fmt.Fprintf(g.state.writer, "%d", v)
	default:
		g.fail(node, "unsupported iota constant kind: %s", val.Kind())
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
			if name.Name == "_" {
				i++
				continue
			}
			typ := g.types.Defs[name].Type()
			cType := g.mapType(spec, typ)
			fmt.Fprintf(w, "%s%s %s = ", g.indent(), cType, name.Name)
			if len(spec.Values) > i {
				g.emitExprAsType(spec, spec.Values[i], typ)
			} else {
				fmt.Fprintf(w, "%s", g.zeroValue(spec, typ))
			}
			i++
			for i < len(spec.Names) {
				nextName := spec.Names[i]
				if nextName.Name == "_" {
					break
				}
				nextTyp := g.types.Defs[nextName].Type()
				nextCType := g.mapType(spec, nextTyp)
				if nextCType != cType {
					break
				}
				fmt.Fprintf(w, ", %s = ", nextName.Name)
				if len(spec.Values) > i {
					g.emitExprAsType(spec, spec.Values[i], nextTyp)
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
		if name.Name == "_" {
			continue
		}
		typ := g.types.Defs[name].Type()
		ct := g.mapCType(spec, typ)
		specifier := ""
		if g.state.indent == 0 {
			// Package-level variable.
			if !ast.IsExported(name.Name) {
				specifier = "static "
			}
		}
		cName := g.declSymbolName(name.Name)
		if len(spec.Values) > i {
			// Has explicit initializer.
			fmt.Fprintf(w, "%s%s%s = ", g.indent(), specifier, ct.Decl(cName))
			g.emitExprAsType(spec, spec.Values[i], typ)
			fmt.Fprintf(w, ";\n")
		} else {
			// No initializer, emit zero value.
			zeroVal := g.zeroValue(spec, typ)
			fmt.Fprintf(w, "%s%s%s = %s;\n", g.indent(), specifier, ct.Decl(cName), zeroVal)
		}
	}
}

// emitTypeSpec dispatches type declaration emission based on the spec type.
func (g *Generator) emitTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	switch spec.Type.(type) {
	case *ast.FuncType:
		g.emitFuncTypeSpec(w, spec)

	case *ast.Ident, *ast.ArrayType, *ast.StarExpr, *ast.MapType:
		typ := g.types.Defs[spec.Name].Type()
		resolved := typ.Underlying()
		// When the underlying type is a struct and the spec references
		// a named type, preserve the name instead of emitting "so_auto".
		if _, isStruct := resolved.(*types.Struct); isStruct {
			if ident, ok := spec.Type.(*ast.Ident); ok {
				if obj := g.types.Uses[ident]; obj != nil {
					resolved = types.Unalias(obj.Type())
				}
			}
		}
		ct := g.mapCType(spec, resolved)
		cName := g.declSymbolName(spec.Name.Name)
		fmt.Fprintf(w, "%stypedef %s;\n", g.indent(), ct.Decl(cName))

	case *ast.InterfaceType:
		iface := g.types.Defs[spec.Name].Type().Underlying().(*types.Interface)
		if iface.Empty() {
			cType := g.mapType(spec, iface)
			cName := g.declSymbolName(spec.Name.Name)
			fmt.Fprintf(w, "%stypedef %s %s;\n", g.indent(), cType, cName)
		} else {
			g.emitInterfaceTypeSpec(w, spec)
		}

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

// emitLabeledStmt emits a label followed by its statement.
func (g *Generator) emitLabeledStmt(stmt *ast.LabeledStmt) {
	w := g.state.writer
	switch stmt.Stmt.(type) {
	case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt:
		// For loops and switches, only emit the end label
		// (for "break label" -> "goto label_end").
		ast.Walk(g, stmt.Stmt)
		fmt.Fprintf(w, "%s%s_end:;\n", g.indent(), stmt.Label.Name)
	default:
		// For other labels (regular goto targets),
		// emit the label before the statement.
		fmt.Fprintf(w, "%s%s:;\n", g.indent(), stmt.Label.Name)
		ast.Walk(g, stmt.Stmt)
	}
}

// emitRangeStmt emits a range-based for statement.
func (g *Generator) emitRangeStmt(stmt *ast.RangeStmt) {
	typ := g.types.TypeOf(stmt.X).Underlying()
	// Unwrap pointer-to-array so `for range p` dispatches to emitArrayRange.
	if ptr, ok := typ.(*types.Pointer); ok {
		if _, ok := ptr.Elem().Underlying().(*types.Array); ok {
			typ = ptr.Elem().Underlying()
		}
	}
	switch t := typ.(type) {
	case *types.Array:
		g.emitArrayRange(stmt)
	case *types.Slice:
		g.emitSliceRange(stmt)
	case *types.Map:
		g.emitMapRange(stmt)
	case *types.Basic:
		if t.Kind() == types.String || t.Kind() == types.UntypedString {
			g.emitStringRange(stmt)
		} else {
			g.emitIntRange(stmt)
		}
	default:
		g.emitIntRange(stmt)
	}
}

// emitReturnStmt emits a return statement, preceded by any deferred generic calls.
func (g *Generator) emitReturnStmt(stmt *ast.ReturnStmt) {
	w := g.state.writer
	if g.state.inMacro {
		// In macro mode: "return X" becomes just "X;", void return is a no-op.
		if len(stmt.Results) > 0 {
			fmt.Fprintf(w, "%s", g.indent())
			g.emitReturnExpr(stmt)
			fmt.Fprintf(w, ";\n")
		}
		return
	}

	g.emitDeferredCalls()

	if len(stmt.Results) == 0 {
		fmt.Fprintf(w, "%sreturn;\n", g.indent())
		return
	}

	fmt.Fprintf(w, "%sreturn ", g.indent())
	g.emitReturnExpr(stmt)
	fmt.Fprintf(w, ";\n")
}

// emitReturnExpr emits the return value expression (without "return" keyword or ";").
// Handles single-return and multi-return compound literals.
func (g *Generator) emitReturnExpr(stmt *ast.ReturnStmt) {
	w := g.state.writer
	if len(stmt.Results) > 1 {
		info := g.multiReturnFields(stmt, g.state.funcSig)
		if info.resultType != "" {
			fmt.Fprintf(w, "(%s){.val = ", info.resultType)
			g.emitExpr(stmt.Results[0])
			fmt.Fprintf(w, ", .err = ")
			g.emitExpr(stmt.Results[1])
			fmt.Fprintf(w, "}")
			return
		}
		fmt.Fprintf(w, "(%s){.val = ", info.typeName())
		g.emitExpr(stmt.Results[0])
		if info.hasError {
			fmt.Fprintf(w, ", .err = ")
		} else {
			fmt.Fprintf(w, ", .val2 = ")
		}
		g.emitExpr(stmt.Results[1])
		fmt.Fprintf(w, "}")
		return
	}
	retType := g.state.funcSig.Results().At(0).Type()
	g.emitExprAsType(stmt, stmt.Results[0], retType)
}

// emitComments looks up comments for the given nodes from the CommentMap,
// filters out directives, and emits them. Returns true if any were emitted.
func (g *Generator) emitComments(w io.Writer, nodes ...ast.Node) bool {
	var lines []string
	for _, node := range nodes {
		for _, cg := range g.comments[node] {
			for _, c := range cg.List {
				text := strings.TrimSpace(c.Text)
				if strings.HasPrefix(text, "//so:") {
					continue
				}
				lines = append(lines, text)
			}
		}
	}
	if len(lines) == 0 {
		return false
	}
	fmt.Fprintln(w)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return true
}

// emitDeferredCalls emits saved generic deferred calls in LIFO order.
func (g *Generator) emitDeferredCalls() {
	for i := len(g.state.defers) - 1; i >= 0; i-- {
		fmt.Fprintf(g.state.writer, "%s%s;\n", g.indent(), g.state.defers[i])
	}
}

// emitBlock emits the statements within a block, adjusting indentation.
func (g *Generator) emitBlock(block *ast.BlockStmt) {
	g.state.indent++
	g.walkStmts(block.List)
	g.state.indent--
}

// walkStmts walks statements, emitting any associated comments before each.
func (g *Generator) walkStmts(stmts []ast.Stmt) {
	w := g.state.writer
	for _, stmt := range stmts {
		for _, cg := range g.comments[stmt] {
			for _, c := range cg.List {
				fmt.Fprintf(w, "%s%s\n", g.indent(), strings.TrimSpace(c.Text))
			}
		}
		ast.Walk(g, stmt)
	}
}

// isBlockTypeSpec returns true for type specs that emit multi-line blocks
// (structs, non-empty interfaces, func types) and need a blank line separator.
func isBlockTypeSpec(spec *ast.TypeSpec) bool {
	switch spec.Type.(type) {
	case *ast.StructType, *ast.FuncType:
		return true
	case *ast.InterfaceType:
		// Non-empty interfaces are block types; empty ones are single-line typedefs.
		iface := spec.Type.(*ast.InterfaceType)
		return len(iface.Methods.List) > 0
	}
	return false
}
