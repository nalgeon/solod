package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"os"
	"slices"
	"strings"
)

// walkAST traverses the AST rooted at root, dispatching to emit methods.
// The io.Writer is captured by the closure, eliminating the need for g.state.writer.
func (g *Generator) walkAST(w io.Writer, root ast.Node) {
	ast.Inspect(root, func(node ast.Node) bool {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(*failure); ok {
					panic(r) // already a diagnostic; don't reformat as an unexpected panic
				}
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
			g.emitAssignStmt(w, n)
			return false
		case *ast.BlockStmt:
			g.emitBlockStmt(w, n)
			return false
		case *ast.BranchStmt:
			g.emitBranchStmt(w, n)
			return false
		case *ast.DeclStmt:
			return true // recurse into inner Decl
		case *ast.DeferStmt:
			g.emitDeferStmt(w, n)
			return false
		case *ast.ExprStmt:
			g.emitExprStmt(w, n)
			return false
		case *ast.ForStmt:
			g.emitForStmt(w, n)
			return false
		case *ast.FuncDecl:
			g.emitFuncDecl(w, n)
			return false
		case *ast.GenDecl:
			g.emitGenDecl(w, n)
			return false
		case *ast.Ident:
			return true // package name etc
		case *ast.IfStmt:
			g.emitIfStmt(w, n)
			return false
		case *ast.IncDecStmt:
			g.emitIncDecStmt(w, n)
			return false
		case *ast.LabeledStmt:
			g.emitLabeledStmt(w, n)
			return false
		case *ast.RangeStmt:
			g.emitRangeStmt(w, n)
			return false
		case *ast.ReturnStmt:
			g.emitReturnStmt(w, n)
			return false
		case *ast.SwitchStmt:
			g.emitSwitchStmt(w, n)
			return false
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

		return true
	})
}

// emitBlockStmt emits a bare block statement (scoping block inside a function body).
func (g *Generator) emitBlockStmt(w io.Writer, stmt *ast.BlockStmt) {
	fmt.Fprintf(w, "%s{\n", g.indent())
	g.emitBlock(w, stmt)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitBranchStmt emits a break, continue, or goto statement.
func (g *Generator) emitBranchStmt(w io.Writer, stmt *ast.BranchStmt) {
	if stmt.Tok == token.FALLTHROUGH {
		// A switch becomes an if/else chain, which has no case to fall into.
		g.fail(stmt, "fallthrough is not supported")
	}
	if stmt.Label != nil && stmt.Tok == token.BREAK {
		// Labeled break is translated to goto because C has no "break label".
		// ("break label" -> "goto label_end").
		fmt.Fprintf(w, "%sgoto %s_end;\n", g.indent(), stmt.Label.Name)
	} else if stmt.Label != nil && stmt.Tok == token.CONTINUE {
		g.fail(stmt, "labeled continue is not supported")
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
func (g *Generator) emitDeferStmt(w io.Writer, stmt *ast.DeferStmt) {
	_ = w // defer statement does not emit anything
	var buf strings.Builder
	g.emitCallExpr(&buf, stmt.Call)
	g.state.defers = append(g.state.defers, buf.String())
}

// emitExprStmt emits an expression statement.
// Emits deferred generic calls before panic() calls.
func (g *Generator) emitExprStmt(w io.Writer, stmt *ast.ExprStmt) {
	if g.isPanicCall(stmt.X) {
		g.emitDeferredCalls(w)
	}
	// c.Raw intrinsic: emit the string literal as a raw C block.
	if raw, ok := g.cIntrinsic(stmt.X); ok {
		for line := range strings.SplitSeq(raw, "\n") {
			fmt.Fprintf(w, "%s%s\n", g.indent(), line)
		}
		return
	}
	fmt.Fprint(w, g.indent())
	g.emitExpr(w, stmt.X)
	fmt.Fprint(w, ";\n")
}

// emitGenDecl emits a general declaration (var, import, etc.).
func (g *Generator) emitGenDecl(w io.Writer, decl *ast.GenDecl) {
	if found, _ := parseExtern(decl.Doc); found {
		return
	}
	switch decl.Tok {
	case token.IMPORT:
		// Imports are handled separately at [Generator.emitImpl].
		return
	case token.CONST:
		if g.state.atTopLevel() {
			// Package-level consts are hoisted by emitPackageVars.
			return
		}
		for _, spec := range decl.Specs {
			g.emitConstSpec(w, spec.(*ast.ValueSpec))
		}
	case token.VAR:
		if g.state.atTopLevel() {
			// Package-level vars are hoisted by emitPackageVars.
			return
		}
		for _, spec := range decl.Specs {
			vs := spec.(*ast.ValueSpec)
			if len(vs.Names) > 0 && g.embeds.vars[vs.Names[0].Name] {
				// Do not emit variables that are used as markers for embedded files.
				continue
			}
			g.emitVarSpec(w, vs, directives{})
		}
	case token.TYPE:
		// Package-level types are emitted by emitUnexportedTypes (unexported)
		// or emitHeaderDecls (exported). Only emit inside function bodies.
		if g.state.atTopLevel() {
			return
		}
		for _, spec := range decl.Specs {
			ts := spec.(*ast.TypeSpec)
			if isConstraintInterface(g.types.Defs[ts.Name].Type()) {
				// A constraint has nothing to represent in C.
				continue
			}
			g.emitComments(w, decl, ts)
			g.emitTypeSpec(w, ts, directives{})
		}
	default:
		g.fail(decl, "unsupported GenDecl token: %s", decl.Tok)
	}
}

// emitConstSpec emits a single constant specification.
func (g *Generator) emitConstSpec(w io.Writer, spec *ast.ValueSpec) {
	for i, name := range spec.Names {
		if name.Name == "_" {
			continue
		}
		typ := g.constType(spec, g.types.Defs[name])
		cType := g.mapTypeName(spec, typ)

		// Check if this is an iota-based constant (implicit value or explicit iota usage).
		isIota := isIotaValue(spec, i)
		if !isIota {
			g.checkConstShadow(spec, i)
		}

		// Determine constant specifier and name.
		specifier, constName := "", name.Name
		if g.state.atTopLevel() {
			// Exported package-level constants are emitted
			// in the header with static linkage.
			if ast.IsExported(constName) {
				continue
			}
			specifier = "static "
			constName = g.mapObjName(g.types.Defs[name])
		}

		// Emit the constant declaration. Go allows unused constants, so mark
		// them so_unused to avoid unused-variable warnings from the C compiler.
		fmt.Fprintf(w, "%s%sconst so_unused %s %s = ", g.indent(), specifier, cType, constName)
		if isIota {
			g.emitConstVal(w, spec, name)
		} else {
			g.emitExpr(w, spec.Values[i])
		}
		fmt.Fprint(w, ";\n")
	}
}

// checkConstShadow rejects a local constant with a value that reads its own name.
func (g *Generator) checkConstShadow(spec *ast.ValueSpec, i int) {
	if g.state.atTopLevel() {
		return
	}
	valueNames := collectIdents(spec.Values[i])
	for _, name := range spec.Names {
		if name.Name == "_" {
			continue
		}
		if valueNames[name.Name] {
			g.fail(spec, "self-shadowing constant %q is not supported", name.Name)
		}
	}
}

// emitConstVal emits the type-checker-resolved value of a constant.
func (g *Generator) emitConstVal(w io.Writer, node ast.Node, name *ast.Ident) {
	obj := g.types.Defs[name].(*types.Const)
	val := obj.Val()
	switch {
	case isFloatType(obj.Type()):
		fmt.Fprint(w, g.floatLit(node, val, obj.Type()))
	case val.Kind() == constant.Int:
		fmt.Fprint(w, intLit(val))
	default:
		g.fail(node, "unsupported iota constant kind: %s", val.Kind())
	}
}

// emitVarSpec emits a single var specification (e.g. `var a int = 1`).
// dirs provides parsed so: directives for package-level declarations.
func (g *Generator) emitVarSpec(w io.Writer, spec *ast.ValueSpec, dirs directives) {
	// Detect self-shadowing in local variable declarations.
	if !g.state.atTopLevel() && len(spec.Values) > 0 {
		rhsNames := collectIdents(spec.Values...)
		for _, name := range spec.Names {
			if name.Name == "_" {
				continue
			}
			if rhsNames[name.Name] {
				g.fail(spec, "self-shadowing variable %q is not supported", name.Name)
			}
		}
	}

	// Local multi-variable declaration: group consecutive same-type variables,
	// but emit separate declarations for different types
	// (e.g. `int a = 1, b = 2; float c = 3.14;`).
	if !g.state.atTopLevel() && len(spec.Names) > 1 {
		// emitInit emits the i-th initializer, or the zero value if absent.
		emitInit := func(i int, typ types.Type) {
			if len(spec.Values) > i {
				g.emitExprAsType(w, spec, spec.Values[i], typ)
			} else {
				fmt.Fprint(w, g.zeroValue(spec, typ))
			}
		}
		i := 0
		for i < len(spec.Names) {
			name := spec.Names[i]
			if name.Name == "_" {
				// Blank identifier - the value is still evaluated.
				g.emitDiscardVar(w, spec, i)
				i++
				continue
			}
			typ := g.types.Defs[name].Type()
			ct := g.mapVarType(spec, typ, len(spec.Values) > i)

			// An array with an initializer needs a declaration of its own.
			if ct.IsArray() && len(spec.Values) > i {
				g.emitArrayVarDecl(w, ct, name.Name, spec.Values[i])
				i++
				continue
			}

			// Emit the leading declarator: "T name = init".
			fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(name.Name))
			emitInit(i, typ)
			i++

			// Arrays, pointers and anonymous structs can't be grouped:
			//  - an array carries its dimension after the name (so_byte a[8])
			//  - `T* a, b` declares a as T* but b as T
			//  - __auto_type allows only one declarator per statement
			_, isPtr := ptrElem(typ)
			if ct.IsArray() || isPtr || ct.FuncPtr || isAnonStruct(typ) {
				fmt.Fprint(w, ";\n")
				continue
			}

			// Group following variables of the same scalar type.
			for i < len(spec.Names) {
				nextName := spec.Names[i]
				if nextName.Name == "_" {
					break
				}
				nextTyp := g.types.Defs[nextName].Type()
				nextCt := g.mapTypeDecl(spec, nextTyp)
				if nextCt.IsArray() || nextCt.Base != ct.Base {
					break
				}
				fmt.Fprintf(w, ", %s = ", nextName.Name)
				emitInit(i, nextTyp)
				i++
			}
			fmt.Fprint(w, ";\n")
		}
		return
	}

	// Single variable or package-level declaration.
	for i, name := range spec.Names {
		if name.Name == "_" {
			// Blank identifier - the value is still evaluated.
			g.emitDiscardVar(w, spec, i)
			continue
		}
		typ := g.types.Defs[name].Type()
		// Anonymous structs are only supported as local variables.
		if g.state.atTopLevel() && isAnonStruct(typ) {
			g.fail(spec, "use a named struct type instead of an anonymous struct")
		}
		ct := g.mapVarType(spec, typ, len(spec.Values) > i)
		specifier := ""
		if g.state.atTopLevel() {
			// Package-level variable: build specifier with qualifiers.
			if !ast.IsExported(name.Name) && !dirs.promote {
				// Go allows unused package-level variables, so mark them
				// so_unused to avoid unused-variable warnings from the C compiler.
				specifier = "static so_unused "
			}
			if dirs.threadLocal {
				specifier += "_Thread_local "
			}
			if dirs.volatile {
				specifier += "volatile "
			}
			if attr := dirs.attrString(); attr != "" {
				specifier += attr + " "
			}
		}
		cName := g.mapObjName(g.types.Defs[name])
		if len(spec.Values) > i {
			// Variable has an explicit initializer.
			if ct.IsArray() && !g.state.atTopLevel() {
				g.emitArrayVarDecl(w, ct, cName, spec.Values[i])
				continue
			}
			fmt.Fprintf(w, "%s%s%s = ", g.indent(), specifier, ct.Decl(cName))
			g.emitExprAsType(w, spec, spec.Values[i], typ)
			fmt.Fprint(w, ";\n")
		} else {
			// No initializer, emit zero value.
			zeroVal := g.zeroValue(spec, typ)
			fmt.Fprintf(w, "%s%s%s = %s;\n", g.indent(), specifier, ct.Decl(cName), zeroVal)
		}
	}
}

// emitTypeSpec dispatches type declaration emission based on the spec type.
// dirs provides parsed so: directives for package-level declarations.
func (g *Generator) emitTypeSpec(w io.Writer, spec *ast.TypeSpec, dirs directives) {
	switch spec.Type.(type) {
	case *ast.FuncType:
		g.emitFuncTypeSpec(w, spec)

	case *ast.Ident, *ast.SelectorExpr, *ast.ArrayType, *ast.StarExpr, *ast.MapType:
		typ := g.types.Defs[spec.Name].Type()
		resolved := typ.Underlying()
		if alias, ok := typ.(*types.Alias); ok {
			resolved = types.Unalias(alias)
		}
		// The underlying type of a named struct is an anonymous struct, which
		// mapTypeDecl rejects. When the spec references a named type, use that
		// name instead.
		if _, isStruct := resolved.(*types.Struct); isStruct {
			var refIdent *ast.Ident
			switch t := spec.Type.(type) {
			case *ast.Ident:
				refIdent = t
			case *ast.SelectorExpr:
				refIdent = t.Sel
			}
			if refIdent != nil {
				if obj := g.types.Uses[refIdent]; obj != nil {
					resolved = types.Unalias(obj.Type())
				}
			}
		}
		ct := g.mapTypeDecl(spec, resolved)
		cName := g.mapObjName(g.types.Defs[spec.Name])
		attr := dirs.attrString()
		if attr != "" {
			fmt.Fprintf(w, "%stypedef %s %s;\n", g.indent(), attr, ct.Decl(cName))
		} else {
			fmt.Fprintf(w, "%stypedef %s;\n", g.indent(), ct.Decl(cName))
		}

	case *ast.InterfaceType:
		iface := g.types.Defs[spec.Name].Type().Underlying().(*types.Interface)
		if iface.Empty() {
			cType := g.mapTypeName(spec, iface)
			cName := g.mapObjName(g.types.Defs[spec.Name])
			fmt.Fprintf(w, "%stypedef %s %s;\n", g.indent(), cType, cName)
		} else {
			g.emitInterfaceTypeSpec(w, spec)
		}

	case *ast.StructType:
		g.emitStructTypeSpec(w, spec, dirs)

	default:
		g.fail(spec, "unsupported type: %T", spec.Type)
	}
}

// emitIfStmt emits an if statement, wrapping in a scope block if there's an init statement.
func (g *Generator) emitIfStmt(w io.Writer, stmt *ast.IfStmt) {
	if stmt.Init == nil {
		g.emitIfInner(w, stmt, g.indent())
		return
	}
	fmt.Fprintf(w, "%s{\n", g.indent())
	g.emitScopedIf(w, stmt)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitScopedIf emits the init statement and the if chain one level deeper.
// The caller emits the enclosing braces, which scope the init statement.
func (g *Generator) emitScopedIf(w io.Writer, stmt *ast.IfStmt) {
	g.state.depth++
	g.walkAST(w, stmt.Init)
	g.emitIfInner(w, stmt, g.indent())
	g.state.depth--
}

// emitIfInner emits the if/else-if/else chain. The prefix controls leading
// indentation: top-level calls pass g.indent(), recursive else-if calls pass "".
func (g *Generator) emitIfInner(w io.Writer, stmt *ast.IfStmt, prefix string) {
	// Emit the if condition and body.
	fmt.Fprintf(w, "%sif (", prefix)
	g.emitExpr(w, stmt.Cond)
	fmt.Fprint(w, ") {\n")
	g.emitBlock(w, stmt.Body)
	if stmt.Else == nil {
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	// Handle else-if and else clauses.
	switch els := stmt.Else.(type) {
	case *ast.IfStmt:
		if els.Init != nil {
			// C has no init statement in if, so the else-if becomes a nested
			// block. The block scopes the init statement like Go does.
			fmt.Fprintf(w, "%s} else {\n", g.indent())
			g.emitScopedIf(w, els)
			fmt.Fprintf(w, "%s}\n", g.indent())
			return
		}
		fmt.Fprintf(w, "%s} else ", g.indent())
		g.emitIfInner(w, els, "")
	case *ast.BlockStmt:
		fmt.Fprintf(w, "%s} else {\n", g.indent())
		g.emitBlock(w, els)
		fmt.Fprintf(w, "%s}\n", g.indent())
	default:
		g.fail(stmt.Else, "unsupported else clause: %T", stmt.Else)
	}
}

// emitIncDecStmt emits an increment or decrement statement.
func (g *Generator) emitIncDecStmt(w io.Writer, stmt *ast.IncDecStmt) {
	g.checkMapIndex(stmt.X)
	fmt.Fprint(w, g.indent())
	g.emitPostfixOperand(w, stmt.X)
	fmt.Fprintf(w, "%s;\n", stmt.Tok)
}

// emitLabeledStmt emits a label followed by its statement.
func (g *Generator) emitLabeledStmt(w io.Writer, stmt *ast.LabeledStmt) {
	name := stmt.Label.Name
	switch stmt.Stmt.(type) {
	case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt:
		// A label on a loop/switch may be a goto target (jump before)
		// and/or a break target (jump after). Emit both labels and rely
		// on -Wno-unused-label in CFLAGS.
		fmt.Fprintf(w, "%s%s:;\n", g.indent(), name)
		g.walkAST(w, stmt.Stmt)
		fmt.Fprintf(w, "%s%s_end:;\n", g.indent(), name)
	default:
		// For other labels (regular goto targets),
		// emit the label before the statement.
		fmt.Fprintf(w, "%s%s:;\n", g.indent(), name)
		g.walkAST(w, stmt.Stmt)
	}
}

// emitRangeStmt emits a range-based for statement.
func (g *Generator) emitRangeStmt(w io.Writer, stmt *ast.RangeStmt) {
	stmt = unparenRange(stmt)
	typ := g.types.TypeOf(stmt.X).Underlying()
	// Unwrap pointer-to-array so `for range p` dispatches to emitArrayRange.
	if ptr, ok := typ.(*types.Pointer); ok {
		if _, ok := ptr.Elem().Underlying().(*types.Array); ok {
			typ = ptr.Elem().Underlying()
		}
	}
	switch t := typ.(type) {
	case *types.Array:
		g.emitArrayRange(w, stmt)
	case *types.Slice:
		g.emitSliceRange(w, stmt)
	case *types.Map:
		g.emitMapRange(w, stmt)
	case *types.Basic:
		if t.Kind() == types.String || t.Kind() == types.UntypedString {
			g.emitStringRange(w, stmt)
		} else {
			g.emitIntRange(w, stmt)
		}
	default:
		g.emitIntRange(w, stmt)
	}
}

// emitComments looks up comments for the given nodes from the CommentMap,
// filters out directives, and emits them. Returns true if any were emitted.
func (g *Generator) emitComments(w io.Writer, nodes ...ast.Node) bool {
	var lines []string
	for _, node := range nodes {
		for _, cg := range g.comments[node] {
			for _, c := range cg.List {
				text, ok := g.commentText(c)
				if !ok {
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
func (g *Generator) emitDeferredCalls(w io.Writer) {
	for _, call := range slices.Backward(g.state.defers) {
		fmt.Fprintf(w, "%s%s;\n", g.indent(), call)
	}
}

// emitBlock emits the statements within a block, adjusting indentation.
func (g *Generator) emitBlock(w io.Writer, block *ast.BlockStmt) {
	g.state.depth++
	g.walkStmts(w, block.List)
	g.state.depth--
}

// walkStmts walks statements, emitting any associated comments before each.
func (g *Generator) walkStmts(w io.Writer, stmts []ast.Stmt) {
	for _, stmt := range stmts {
		for _, cg := range g.comments[stmt] {
			for _, c := range cg.List {
				if text, ok := g.commentText(c); ok {
					fmt.Fprintf(w, "%s%s\n", g.indent(), text)
				}
			}
		}
		if g.opts.TrackSource && g.state.scope != macroScope {
			pos := g.pkg.Fset.Position(stmt.Pos())
			fmt.Fprintf(w, "#line %d \"%s\"\n", pos.Line, pos.Filename)
		}
		g.walkAST(w, stmt)
	}
}

// commentText returns the text of a comment for output.
// It reports false for a directive, which emits nothing.
func (g *Generator) commentText(c *ast.Comment) (string, bool) {
	text := strings.TrimSpace(c.Text)
	if strings.HasPrefix(text, "//so:") {
		return "", false
	}
	if g.state.scope != macroScope || !strings.HasPrefix(text, "//") {
		return text, true
	}
	// A macro body ends every line with a backslash-newline, and C removes
	// the backslash-newline before it removes the comments. A line comment
	// inside a macro deletes the rest of the macro, so the generator writes
	// a block comment. A */ in the text closes the block comment too early.
	body := strings.ReplaceAll(strings.TrimPrefix(text, "//"), "*/", "* /")
	return "/*" + body + " */", true
}
