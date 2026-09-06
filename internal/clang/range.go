package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
)

// emitIntRange emits a range loop over an integer.
func (g *Generator) emitIntRange(w io.Writer, stmt *ast.RangeStmt) {
	if stmt.Key == nil {
		// Basic form: `for range n { ... }`
		fmt.Fprintf(w, "%sfor (so_int _i = 0; _i < ", g.indent())
		g.emitExpr(w, stmt.X)
		fmt.Fprint(w, "; _i++) {\n")
		g.emitBlock(w, stmt.Body)
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	key := stmt.Key.(*ast.Ident)
	k := g.rangeKeyVar(stmt, key)
	fmt.Fprintf(w, "%sfor (%s%s = 0; %s < ", g.indent(), k.decl, k.name, k.name)
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, "; %s++) {\n", k.name)
	g.emitRangeKeyCopy(w, k)
	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitArrayRange emits a range loop over a fixed-size array.
func (g *Generator) emitArrayRange(w io.Writer, stmt *ast.RangeStmt) {
	if _, ok := ast.Unparen(stmt.X).(*ast.CompositeLit); ok {
		g.fail(stmt.X, "for-range over literal not supported")
	}

	// Unwrap pointer-to-array to get the array type.
	typ := g.types.TypeOf(stmt.X).Underlying()
	ptrDeref := false
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem().Underlying()
		ptrDeref = true
	}
	arrType := typ.(*types.Array)

	if stmt.Key == nil {
		// Basic form: `for range arr { ... }`
		fmt.Fprintf(w, "%sfor (so_int _i = 0; _i < %d; _i++) {\n", g.indent(), arrType.Len())
		g.emitBlock(w, stmt.Body)
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	key := stmt.Key.(*ast.Ident)
	k := g.rangeKeyVar(stmt, key)

	fmt.Fprintf(w, "%sfor (%s%s = 0; %s < %d; %s++) {\n",
		g.indent(), k.decl, k.name, k.name, arrType.Len(), k.name)
	g.emitRangeKeyCopy(w, k)

	// Emit value variable if present (e.g. `for i, v := range nums`).
	if stmt.Value != nil {
		if valIdent, ok := stmt.Value.(*ast.Ident); ok && valIdent.Name != "_" {
			g.state.depth++
			g.emitArrayRangeValue(w, stmt, arrType.Elem(), valIdent.Name, k.name, ptrDeref)
			g.state.depth--
		}
	}

	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitArrayRangeValue emits the value variable of a range loop over an array,
// e.g. `so_int v = nums[i];`. name is the value variable and index is the key
// variable.
func (g *Generator) emitArrayRangeValue(w io.Writer, stmt *ast.RangeStmt, elem types.Type, name, index string, ptrDeref bool) {
	emitElem := func() { g.emitArrayRangeElem(w, stmt.X, index, ptrDeref) }

	if isUnderlyingArray(elem) {
		// Range copies an array element, so the value needs a memcpy.
		decl := g.mapTypeDecl(stmt, elem).Decl(name)
		if stmt.Tok == token.ASSIGN {
			decl = ""
		}
		g.emitArrayCopy(w, decl, name, emitElem)
		return
	}

	valDecl := g.mapTypeName(stmt, elem) + " "
	if stmt.Tok == token.ASSIGN {
		valDecl = ""
	}
	fmt.Fprintf(w, "%s%s%s = ", g.indent(), valDecl, name)
	emitElem()
	fmt.Fprint(w, ";\n")
}

// emitArrayRangeElem emits the element access of a range loop over an array,
// e.g. `nums[i]`. ptrDeref reports whether x is a pointer to the array.
func (g *Generator) emitArrayRangeElem(w io.Writer, x ast.Expr, index string, ptrDeref bool) {
	// A dereference needs parens, because [] binds tighter than * in C.
	_, srcDeref := x.(*ast.StarExpr)
	switch {
	case ptrDeref:
		fmt.Fprint(w, "(*")
		g.emitExpr(w, x)
		fmt.Fprint(w, ")")
	case srcDeref:
		fmt.Fprint(w, "(")
		g.emitExpr(w, x)
		fmt.Fprint(w, ")")
	default:
		g.emitExpr(w, x)
	}
	fmt.Fprintf(w, "[%s]", index)
}

// emitSliceRange emits a range loop over a slice.
func (g *Generator) emitSliceRange(w io.Writer, stmt *ast.RangeStmt) {
	if _, ok := ast.Unparen(stmt.X).(*ast.CompositeLit); ok {
		g.fail(stmt.X, "for-range over literal not supported")
	}
	if stmt.Key == nil {
		// Basic form: `for range slice { ... }`
		fmt.Fprintf(w, "%sfor (so_int _i = 0; _i < so_len(", g.indent())
		g.emitExpr(w, stmt.X)
		fmt.Fprint(w, "); _i++) {\n")
		g.emitBlock(w, stmt.Body)
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	key := stmt.Key.(*ast.Ident)
	sliceType := g.types.TypeOf(stmt.X).Underlying().(*types.Slice)
	elemType := g.mapTypeName(stmt, sliceType.Elem())
	k := g.rangeKeyVar(stmt, key)

	fmt.Fprintf(w, "%sfor (%s%s = 0; %s < so_len(", g.indent(), k.decl, k.name, k.name)
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, "); %s++) {\n", k.name)
	g.emitRangeKeyCopy(w, k)

	// Emit value variable if present (e.g. `for i, v := range nums`).
	if stmt.Value != nil {
		if valIdent, ok := stmt.Value.(*ast.Ident); ok && valIdent.Name != "_" {
			g.state.depth++
			valDecl := elemType + " "
			if stmt.Tok == token.ASSIGN {
				valDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = so_at(%s, ", g.indent(), valDecl, valIdent.Name, elemType)
			g.emitExpr(w, stmt.X)
			fmt.Fprintf(w, ", %s);\n", k.name)
			g.state.depth--
		}
	}

	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// emitStringRange emits a range loop over a string (rune iteration).
func (g *Generator) emitStringRange(w io.Writer, stmt *ast.RangeStmt) {
	if stmt.Key == nil {
		// Basic form: `for range str { ... }`
		fmt.Fprintf(w, "%sfor (so_int _i = 0, _iw = 0; _i < so_len(", g.indent())
		g.emitExpr(w, stmt.X)
		fmt.Fprint(w, "); _i += _iw) {\n")
		g.state.depth++
		fmt.Fprintf(w, "%s_iw = 0;\n", g.indent())
		fmt.Fprintf(w, "%sso_utf8_decode(", g.indent())
		g.emitExpr(w, stmt.X)
		fmt.Fprint(w, ", _i, &_iw);\n")
		g.state.depth--
		g.emitBlock(w, stmt.Body)
		fmt.Fprintf(w, "%s}\n", g.indent())
		return
	}

	key := stmt.Key.(*ast.Ident)
	k := g.rangeKeyVar(stmt, key)
	widthVar := "_" + key.Name + "w"

	fmt.Fprintf(w, "%sfor (%s%s = 0, %s = 0; %s < so_len(", g.indent(), k.decl, k.name, widthVar, k.name)
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, "); %s += %s) {\n", k.name, widthVar)
	g.emitRangeKeyCopy(w, k)

	// Decode rune and width once per iteration.
	g.state.depth++
	fmt.Fprintf(w, "%s%s = 0;\n", g.indent(), widthVar)
	if stmt.Value != nil {
		if valIdent, ok := stmt.Value.(*ast.Ident); ok && valIdent.Name != "_" {
			valDecl := "so_rune "
			if stmt.Tok == token.ASSIGN {
				valDecl = ""
			}
			fmt.Fprintf(w, "%s%s%s = so_utf8_decode(", g.indent(), valDecl, valIdent.Name)
		} else {
			fmt.Fprintf(w, "%sso_utf8_decode(", g.indent())
		}
	} else {
		fmt.Fprintf(w, "%sso_utf8_decode(", g.indent())
	}
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, ", %s, &%s);\n", k.name, widthVar)
	g.state.depth--

	g.emitBlock(w, stmt.Body)

	fmt.Fprintf(w, "%s}\n", g.indent())
}

// rangeKey describes the C loop variable of a range statement.
type rangeKey struct {
	decl string // type prefix for the init clause, with a trailing space
	name string // name of the C loop variable
	copy string // Go key variable to copy into, empty when the loop needs no copy
}

// rangeKeyVar returns the C loop variable for a range statement.
//
// A define form ("for i := range x") declares the key in the init clause, so
// the loop counts on the key. An assign form ("for i = range x") must leave
// the key at the index of the last iteration, and a C loop variable stops one
// past it. So the loop counts on a variable of its own and copies it to the
// key at the top of the body.
func (g *Generator) rangeKeyVar(stmt *ast.RangeStmt, key *ast.Ident) rangeKey {
	if key.Name == "_" {
		return rangeKey{decl: "so_int ", name: "_"}
	}
	if stmt.Tok == token.ASSIGN {
		// The name follows the key, like the width variable of a string range,
		// so it cannot conflict with the local names of the builtin macros.
		return rangeKey{decl: "so_int ", name: "_" + key.Name + "i", copy: key.Name}
	}
	return rangeKey{decl: g.mapTypeName(stmt, g.types.Defs[key].Type()) + " ", name: key.Name}
}

// emitRangeKeyCopy writes the key assignment at the top of a range loop body.
func (g *Generator) emitRangeKeyCopy(w io.Writer, k rangeKey) {
	if k.copy == "" {
		return
	}
	g.state.depth++
	fmt.Fprintf(w, "%s%s = %s;\n", g.indent(), k.copy, k.name)
	g.state.depth--
}

// unparenRange removes the parentheses around the key and the value of a range loop.
func unparenRange(stmt *ast.RangeStmt) *ast.RangeStmt {
	key, value := stmt.Key, stmt.Value
	if key != nil {
		key = ast.Unparen(key)
	}
	if value != nil {
		value = ast.Unparen(value)
	}
	if key == stmt.Key && value == stmt.Value {
		return stmt
	}
	clone := *stmt
	clone.Key, clone.Value = key, value
	return &clone
}
