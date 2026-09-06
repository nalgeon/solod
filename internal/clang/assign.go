package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"slices"
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
		g.checkMapIndex(stmt.Lhs[0])
		g.checkAssignCall(stmt)
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
	// Regular define: group consecutive variables by type.
	rhs := stmt.Rhs
	if g.needsRhsTemps(stmt) {
		rhs = g.hoistRhs(w, stmt)
	}
	i := 0
	for i < len(stmt.Lhs) {
		ident := stmt.Lhs[i].(*ast.Ident)
		if ident.Name == "_" {
			// Blank identifier - the value is still evaluated.
			if rhs[i] != nil {
				g.emitDiscard(w, rhs[i])
			}
			i++
			continue
		}

		def := g.types.Defs[ident]
		if def == nil {
			// Redeclared variable - emit plain assignment.
			typ := g.types.Uses[ident].Type()
			fmt.Fprintf(w, "%s%s = ", g.indent(), ident.Name)
			g.emitExprAsType(w, stmt, rhs[i], typ)
			fmt.Fprint(w, ";\n")
			i++
			continue
		}

		typ := def.Type()
		ct := g.mapVarType(stmt, typ, true)

		if ct.IsArray() {
			// Arrays can't be grouped with other variables.
			g.emitArrayVarDecl(w, ct, ident.Name, rhs[i])
			i++
			continue
		}

		// Emit a variable declaration for this variable
		// (grouped with subsequent variables of the same type).
		fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ident.Name))
		g.emitExpr(w, rhs[i])
		i++

		// Pointer types, anonymous structs and type parameters can't be grouped:
		//  - `T* a, b` declares a as T* but b as T
		//  - __auto_type allows only one declarator per statement
		//  - a macro takes a pointer type for a type parameter, which gives `T* a, b`
		if ct.IsPointer() || isAnonStruct(typ) || isTypeParam(typ) {
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
			g.emitExpr(w, rhs[i])
			i++
		}
		fmt.Fprint(w, ";\n")
	}
}

// emitAssign emits a regular assignment (=).
func (g *Generator) emitAssign(w io.Writer, stmt *ast.AssignStmt) {
	g.checkAssignTargets(stmt)
	g.checkAssignAnonStruct(stmt)

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
	// Regular assignment.
	rhs := stmt.Rhs
	if g.needsRhsTemps(stmt) {
		rhs = g.hoistRhs(w, stmt)
	}
	for i, lhs := range stmt.Lhs {
		// Blank identifier - emit a void expression.
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			if rhs[i] != nil {
				g.emitDiscard(w, rhs[i])
			}
			continue
		}

		// Map index assignment uses so_map_set.
		if idx, ok := lhs.(*ast.IndexExpr); ok {
			if _, isMap := g.types.TypeOf(idx.X).Underlying().(*types.Map); isMap {
				g.emitMapIndexAssign(w, stmt, idx, rhs[i])
				continue
			}
		}

		// Array assignment uses memcpy.
		lhsType := g.types.TypeOf(lhs)
		if arr, ok := arrayType(lhsType); ok {
			// The array type is known at compile time, so sizeof takes the
			// type instead of the left side to avoid evaluating it twice.
			arrCType := g.mapTypeName(stmt, arr.Elem()) + arrayDims(arr)
			fmt.Fprintf(w, "%smemcpy(", g.indent())
			g.emitExpr(w, lhs)
			fmt.Fprint(w, ", ")
			g.emitArrayValue(w, stmt, rhs[i], arr)
			fmt.Fprintf(w, ", sizeof(%s));\n", arrCType)
			continue
		}

		// Non-array assignment.
		fmt.Fprint(w, g.indent())
		g.emitExpr(w, lhs)
		fmt.Fprint(w, " = ")
		g.emitExprAsType(w, stmt, rhs[i], lhsType)
		fmt.Fprint(w, ";\n")
	}
}

// checkAssignCall rejects a call in the left side of a compound assignment.
// A string += and a guarded /= or %= write the left side two times, so the
// call would run two times and could return a different result if it has
// side effects.
func (g *Generator) checkAssignCall(stmt *ast.AssignStmt) {
	ast.Inspect(stmt.Lhs[0], func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		g.fail(call, "call in the left side of a compound assignment is not supported")
		return false
	})
}

// checkAssignTargets rejects a multiple assignment where a target on the left
// reads a variable the same statement assigns.
func (g *Generator) checkAssignTargets(stmt *ast.AssignStmt) {
	if len(stmt.Lhs) < 2 {
		return
	}
	assigned := g.assignedVars(stmt)
	for _, lhs := range stmt.Lhs {
		// A plain identifier has no operand to evaluate.
		if _, ok := lhs.(*ast.Ident); ok {
			continue
		}
		for name := range collectIdents(lhs) {
			if assigned[name] {
				g.fail(stmt, "multiple assignment reads and assigns %q in the same statement", name)
			}
		}
	}
}

// checkAssignAnonStruct rejects an assignment to an anonymous struct variable
// or field. Every anonymous struct declaration emits a separate C struct type,
// and C does not assign between two separate struct types.
func (g *Generator) checkAssignAnonStruct(stmt *ast.AssignStmt) {
	for _, lhs := range stmt.Lhs {
		// A blank target is not emitted, so its type is irrelevant.
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			continue
		}
		if isAnonStruct(g.types.TypeOf(lhs)) {
			g.fail(stmt, "cannot assign an anonymous struct; declare a named struct type instead")
		}
	}
}

// needsRhsTemps reports whether a multiple assignment must evaluate its right
// side into temporaries to match Go's semantics (evaluate the whole right side
// before assigning any variable on the left).
func (g *Generator) needsRhsTemps(stmt *ast.AssignStmt) bool {
	if len(stmt.Lhs) < 2 || len(stmt.Rhs) < 2 {
		return false
	}
	assigned := g.changedNames(stmt)
	if len(assigned) == 0 {
		return false
	}
	if slices.ContainsFunc(stmt.Rhs, g.callsFunc) {
		return true
	}
	for name := range collectIdents(stmt.Rhs...) {
		if assigned[name] {
			return true
		}
	}
	return false
}

// changedNames returns the names of the variables a multiple assignment
// changes. It adds the root of every target that writes into memory: arr in
// arr[i], p in *p and p.f.
func (g *Generator) changedNames(stmt *ast.AssignStmt) map[string]bool {
	names := g.assignedVars(stmt)
	for _, lhs := range stmt.Lhs {
		if _, ok := lhs.(*ast.Ident); ok {
			continue
		}
		if ident := rootIdent(lhs); ident != nil {
			names[ident.Name] = true
		}
	}
	return names
}

// assignedVars returns the names of the variables a multiple assignment sets.
func (g *Generator) assignedVars(stmt *ast.AssignStmt) map[string]bool {
	names := map[string]bool{}
	for _, lhs := range stmt.Lhs {
		// Only a plain identifier target sets a variable. Any other target
		// writes into memory, and the variable itself keeps its value.
		ident, ok := lhs.(*ast.Ident)
		if !ok || ident.Name == "_" {
			continue
		}
		// Redeclaration does not introduce a new variable.
		if stmt.Tok == token.DEFINE && g.types.Defs[ident] != nil {
			continue
		}
		names[ident.Name] = true
	}
	return names
}

// rootIdent returns the identifier at the root of an assignment target:
// arr in arr[i], p in *p and p.f. It returns nil for any other target.
func rootIdent(expr ast.Expr) *ast.Ident {
	for {
		switch e := expr.(type) {
		case *ast.Ident:
			return e
		case *ast.IndexExpr:
			expr = e.X
		case *ast.StarExpr:
			expr = e.X
		case *ast.SelectorExpr:
			expr = e.X
		case *ast.ParenExpr:
			expr = e.X
		default:
			return nil
		}
	}
}

// callsFunc reports whether an expression calls a function with side effects.
func (g *Generator) callsFunc(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// A pure call still has operands, so the walk goes on.
		if g.isPureCall(call) {
			return true
		}
		found = true
		return false
	})
	return found
}

// isPureCall reports whether a call has no side effects. A type conversion and
// the len and cap builtins are pure. Every other call is not.
func (g *Generator) isPureCall(call *ast.CallExpr) bool {
	if tv, ok := g.types.Types[call.Fun]; ok && tv.IsType() {
		return true
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	b, ok := g.types.Uses[ident].(*types.Builtin)
	return ok && (b.Name() == "len" || b.Name() == "cap")
}

// hoistRhs evaluates the right side of a multiple assignment into temporaries
// and returns a reference for each value.
//
// A blank identifier gets no temporary and no reference. Its value is discarded
// in place, which keeps the calls on the right in order.
func (g *Generator) hoistRhs(w io.Writer, stmt *ast.AssignStmt) []ast.Expr {
	refs := make([]ast.Expr, len(stmt.Rhs))
	for i, rhs := range stmt.Rhs {
		// A constant needs no temporary. No assignment can change it.
		if tv, ok := g.types.Types[rhs]; ok && tv.Value != nil {
			refs[i] = rhs
			continue
		}
		typ := g.assignTargetType(stmt, i)
		if typ == nil {
			g.emitDiscard(w, rhs)
			continue
		}
		ref := &ast.Ident{NamePos: rhs.Pos(), Name: g.newTemp(stmt, tempAssign)}
		// The reference has no declaration to look up, so record its type directly.
		g.types.Types[ref] = types.TypeAndValue{Type: typ}
		refs[i] = ref

		ct := g.mapVarType(stmt, typ, true)
		if !ct.IsArray() {
			fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ref.Name))
			g.emitExprAsType(w, stmt, rhs, typ)
			fmt.Fprint(w, ";\n")
			continue
		}
		// C cannot assign an array, so a composite literal initializes the
		// temporary and any other value is copied into it.
		if lit, isLit := ast.Unparen(rhs).(*ast.CompositeLit); isLit {
			fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(ref.Name))
			g.emitExpr(w, lit)
			fmt.Fprint(w, ";\n")
			continue
		}
		fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(ref.Name))
		fmt.Fprintf(w, "%smemcpy(%s, ", g.indent(), ref.Name)
		g.emitExpr(w, rhs)
		fmt.Fprintf(w, ", sizeof(%s));\n", ref.Name)
	}
	return refs
}

// assignTargetType returns the type of the i-th target of an assignment.
// It returns nil for a blank identifier, which has no type.
func (g *Generator) assignTargetType(stmt *ast.AssignStmt, i int) types.Type {
	lhs := stmt.Lhs[i]
	ident, isIdent := lhs.(*ast.Ident)
	if isIdent && ident.Name == "_" {
		return nil
	}
	if stmt.Tok != token.DEFINE {
		return g.types.TypeOf(lhs)
	}
	if def := g.types.Defs[ident]; def != nil {
		return def.Type()
	}
	return g.types.Uses[ident].Type()
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

// isTypeParam reports whether typ is a type parameter of a generic function.
func isTypeParam(typ types.Type) bool {
	_, ok := types.Unalias(typ).(*types.TypeParam)
	return ok
}
