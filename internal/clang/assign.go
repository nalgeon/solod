package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
)

// emitAssignStmt emits an assignment statement.
func (g *Generator) emitAssignStmt(w io.Writer, stmt *ast.AssignStmt) {
	switch stmt.Tok {
	case token.DEFINE:
		g.emitDefine(w, stmt)

	case token.ASSIGN:
		g.emitAssign(w, stmt)

	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.OR_ASSIGN, token.AND_ASSIGN, token.XOR_ASSIGN,
		token.SHL_ASSIGN, token.SHR_ASSIGN:
		if idx, ok := stmt.Lhs[0].(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.fail(stmt, "compound assignment on map index is not supported")
			}
		}
		// String += uses so_string_add.
		if stmt.Tok == token.ADD_ASSIGN && g.hasStringType(stmt.Lhs[0]) {
			fmt.Fprint(w, g.indent())
			g.emitExpr(w, stmt.Lhs[0])
			fmt.Fprint(w, " = so_string_add(")
			g.emitExpr(w, stmt.Lhs[0])
			fmt.Fprint(w, ", ")
			g.emitExpr(w, stmt.Rhs[0])
			fmt.Fprint(w, ");\n")
			return
		}
		// Integer /= and %= guard against a zero divisor.
		if (stmt.Tok == token.QUO_ASSIGN || stmt.Tok == token.REM_ASSIGN) &&
			g.needsIntDivGuard(stmt.Lhs[0], stmt.Rhs[0]) {
			fmt.Fprint(w, g.indent())
			g.emitExpr(w, stmt.Lhs[0])
			if stmt.Tok == token.QUO_ASSIGN {
				fmt.Fprint(w, " = so_div(")
			} else {
				fmt.Fprint(w, " = so_mod(")
			}
			g.emitExpr(w, stmt.Lhs[0])
			fmt.Fprint(w, ", ")
			g.emitExpr(w, stmt.Rhs[0])
			fmt.Fprint(w, ");\n")
			return
		}
		fmt.Fprint(w, g.indent())
		g.emitExpr(w, stmt.Lhs[0])
		fmt.Fprintf(w, " %s ", stmt.Tok)
		g.emitExpr(w, stmt.Rhs[0])
		fmt.Fprint(w, ";\n")

	default:
		g.fail(stmt, "unsupported AssignStmt token: %s", stmt.Tok)
	}
}

// emitDefine emits a short variable declaration (:=).
func (g *Generator) emitDefine(w io.Writer, stmt *ast.AssignStmt) {
	// Detect: _, ok := s.(Rect)
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if ta, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); ok {
			g.emitTypeAssertion(w, stmt, ta)
			return
		}
	}
	// Map comma-ok: v, ok := m[key]
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if idx, ok := stmt.Rhs[0].(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapCommaOk(w, stmt, idx, true)
				return
			}
		}
	}
	// Multi-return destructuring: x, y := f()
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) == 1 {
		if call, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
			g.emitMultiReturnDefine(w, stmt, call)
			return
		}
	}
	// Detect self-shadowing - a variable x is defined using a variable with
	// the same name from an outer scope, eg. `x := x + 1`. C does not support
	// this (the right-hand x refers to the new variable, not the outer one).
	rhsNames := collectIdents(stmt.Rhs...)
	for _, lhs := range stmt.Lhs {
		ident, ok := lhs.(*ast.Ident)
		if !ok || ident.Name == "_" {
			continue
		}
		if g.types.Defs[ident] == nil {
			continue
		}
		if rhsNames[ident.Name] {
			g.fail(stmt, "self-shadowing variable %q is not supported", ident.Name)
		}
	}
	// Detect LHS/RHS variable overlap in multi-assignments.
	// Eg. `a, b = x, y` is fine, but `a, b = b, a` is not.
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) > 1 {
		lhsNames := collectIdents(stmt.Lhs...)
		for name := range rhsNames {
			if lhsNames[name] {
				g.fail(stmt, "multiple assignment with LHS/RHS variable overlap is not supported")
			}
		}
	}
	// Regular define: group consecutive variables by type.
	i := 0
	for i < len(stmt.Lhs) {
		ident := stmt.Lhs[i].(*ast.Ident)
		if ident.Name == "_" {
			i++
			continue
		}

		def := g.types.Defs[ident]
		if def == nil {
			// Redeclared variable - emit plain assignment.
			typ := g.types.Uses[ident].Type()
			fmt.Fprintf(w, "%s%s = ", g.indent(), ident.Name)
			g.emitExprAsType(w, stmt, stmt.Rhs[i], typ)
			fmt.Fprint(w, ";\n")
			i++
			continue
		}

		typ := def.Type()
		ct := g.mapTypeDecl(stmt, typ)

		if ct.IsArray() {
			// Arrays can't be grouped with other variables.
			if _, isLit := stmt.Rhs[i].(*ast.CompositeLit); isLit {
				// Composite literal: so_int d[3] = {1, 2, 3};
				fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ident.Name))
				g.emitExpr(w, stmt.Rhs[i])
				fmt.Fprint(w, ";\n")
			} else {
				// Variable: declaration + memcpy.
				fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(ident.Name))
				fmt.Fprintf(w, "%smemcpy(%s, ", g.indent(), ident.Name)
				g.emitExpr(w, stmt.Rhs[i])
				fmt.Fprintf(w, ", sizeof(%s));\n", ident.Name)
			}
			i++
			continue
		}

		// Emit a variable declaration for this variable
		// (grouped with subsequent variables of the same type).
		fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ident.Name))
		g.emitExpr(w, stmt.Rhs[i])
		i++

		// Pointer types and anonymous structs can't be grouped:
		//  - `T* a, b` declares a as T* but b as T
		//  - __auto_type allows only one declarator per statement
		if ct.IsPointer() || isAnonStruct(typ) {
			fmt.Fprint(w, ";\n")
			continue
		}

		for i < len(stmt.Lhs) {
			nextIdent := stmt.Lhs[i].(*ast.Ident)
			if nextIdent.Name == "_" {
				break
			}
			nextDef := g.types.Defs[nextIdent]
			if nextDef == nil {
				break
			}
			nextCType := g.mapTypeName(stmt, nextDef.Type())
			if nextCType != ct.Base {
				break
			}
			if isArrayType(nextDef.Type()) {
				break
			}
			fmt.Fprintf(w, ", %s = ", nextIdent.Name)
			g.emitExpr(w, stmt.Rhs[i])
			i++
		}
		fmt.Fprint(w, ";\n")
	}
}

// emitAssign emits a regular assignment (=).
func (g *Generator) emitAssign(w io.Writer, stmt *ast.AssignStmt) {
	// Detect: _, ok = s.(Rect)
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if ta, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); ok {
			g.emitTypeAssertion(w, stmt, ta)
			return
		}
	}
	// Map comma-ok: v, ok = m[key]
	if len(stmt.Lhs) == 2 && len(stmt.Rhs) == 1 {
		if idx, ok := stmt.Rhs[0].(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapCommaOk(w, stmt, idx, false)
				return
			}
		}
	}
	// Multi-return destructuring: x, y = f()
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) == 1 {
		if call, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
			g.emitMultiReturnAssign(w, stmt, call)
			return
		}
	}
	// Detect LHS/RHS variable overlap in multi-assignments.
	if len(stmt.Lhs) > 1 && len(stmt.Rhs) > 1 {
		lhsNames := collectIdents(stmt.Lhs...)
		rhsNames := collectIdents(stmt.Rhs...)
		for name := range rhsNames {
			if lhsNames[name] {
				g.fail(stmt, "multiple assignment with LHS/RHS variable overlap is not supported")
			}
		}
	}
	// Regular assignment.
	for i, lhs := range stmt.Lhs {
		// Blank identifier - emit a void expression.
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			g.emitDiscard(w, stmt.Rhs[i])
			continue
		}

		// Map index assignment uses so_map_set.
		if idx, ok := lhs.(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapIndexAssign(w, stmt, idx, stmt.Rhs[i])
				continue
			}
		}

		// Array assignment uses memcpy.
		lhsType := g.types.TypeOf(lhs)
		if arr, ok := lhsType.Underlying().(*types.Array); ok {
			fmt.Fprintf(w, "%smemcpy(", g.indent())
			g.emitExpr(w, lhs)
			fmt.Fprint(w, ", ")
			if _, isLit := stmt.Rhs[i].(*ast.CompositeLit); isLit {
				// Compound literal: (int[3]){1, 2, 3}
				elemType := g.mapTypeName(stmt, arr.Elem())
				fmt.Fprintf(w, "(%s%s)", elemType, arrayDims(arr))
			}
			g.emitExpr(w, stmt.Rhs[i])
			fmt.Fprint(w, ", sizeof(")
			g.emitExpr(w, lhs)
			fmt.Fprint(w, "));\n")
			continue
		}

		// Non-array assignment.
		fmt.Fprint(w, g.indent())
		g.emitExpr(w, lhs)
		fmt.Fprint(w, " = ")
		g.emitExprAsType(w, stmt, stmt.Rhs[i], lhsType)
		fmt.Fprint(w, ";\n")
	}
}

// collectIdents returns the set of identifier names in the given expressions.
// The blank identifier is excluded.
func collectIdents(exprs ...ast.Expr) map[string]bool {
	names := map[string]bool{}
	var visit func(ast.Node) bool
	visit = func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Ident:
			if n.Name != "_" {
				names[n.Name] = true
			}
		case *ast.KeyValueExpr:
			// Only recurse into Value, skip Key (struct field names
			// are not variable references; map key variables are
			// also skipped but self-shadowing there is unlikely).
			ast.Inspect(n.Value, visit)
			return false
		case *ast.SelectorExpr:
			// Only recurse into X, skip Sel (field/method names
			// are not variable references).
			ast.Inspect(n.X, visit)
			return false
		}
		return true
	}
	for _, expr := range exprs {
		ast.Inspect(expr, visit)
	}
	return names
}
