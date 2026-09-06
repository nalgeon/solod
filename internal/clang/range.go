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
	k := g.rangeKeyVar(stmt)
	fmt.Fprintf(w, "%sfor (%s%s = 0; %s < ", g.indent(), k.decl, k.name, k.name)
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, "; %s++) {\n", k.name)
	g.emitRangeBody(w, stmt, k, nil)
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

	k := g.rangeKeyVar(stmt)
	fmt.Fprintf(w, "%sfor (%s%s = 0; %s < %d; %s++) {\n",
		g.indent(), k.decl, k.name, k.name, arrType.Len(), k.name)
	g.emitRangeBody(w, stmt, k, func() {
		g.emitArrayRangeValue(w, stmt, arrType.Elem(), k.name, ptrDeref)
	})
}

// emitArrayRangeValue emits the value variable of a range loop over an array,
// e.g. `so_int v = nums[i];`. index is the key variable.
func (g *Generator) emitArrayRangeValue(w io.Writer, stmt *ast.RangeStmt, elem types.Type, index string, ptrDeref bool) {
	name, declare := rangeVar(stmt, stmt.Value)
	if name == "" {
		return
	}
	emitElem := func() { g.emitArrayRangeElem(w, stmt.X, index, ptrDeref) }

	if isUnderlyingArray(elem) {
		// Range copies an array element, so the value needs a memcpy.
		var decl string
		if declare {
			decl = g.mapTypeDecl(stmt, elem).Decl(name)
		}
		g.emitArrayCopy(w, decl, name, emitElem)
		return
	}

	typPrefix := rangeDecl(g.mapTypeName(stmt, elem), declare)
	fmt.Fprintf(w, "%s%s%s = ", g.indent(), typPrefix, name)
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

	k := g.rangeKeyVar(stmt)
	fmt.Fprintf(w, "%sfor (%s%s = 0; %s < so_len(", g.indent(), k.decl, k.name, k.name)
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, "); %s++) {\n", k.name)
	g.emitRangeBody(w, stmt, k, func() {
		name, declare := rangeVar(stmt, stmt.Value)
		if name == "" {
			return
		}
		sliceType := g.types.TypeOf(stmt.X).Underlying().(*types.Slice)
		elemType := g.mapTypeName(stmt, sliceType.Elem())
		fmt.Fprintf(w, "%s%s%s = so_at(%s, ", g.indent(), rangeDecl(elemType, declare), name, elemType)
		g.emitExpr(w, stmt.X)
		fmt.Fprintf(w, ", %s);\n", k.name)
	})
}

// emitStringRange emits a range loop over a string (rune iteration).
func (g *Generator) emitStringRange(w io.Writer, stmt *ast.RangeStmt) {
	k := g.rangeKeyVar(stmt)
	fmt.Fprintf(w, "%sfor (%s%s = 0, %s = 0; %s < so_len(", g.indent(), k.decl, k.name, k.width, k.name)
	g.emitExpr(w, stmt.X)
	fmt.Fprintf(w, "); %s += %s) {\n", k.name, k.width)
	g.emitRangeBody(w, stmt, k, func() {
		// Decode the rune and the width once per iteration.
		fmt.Fprintf(w, "%s%s = 0;\n", g.indent(), k.width)
		fmt.Fprint(w, g.indent())
		if name, declare := rangeVar(stmt, stmt.Value); name != "" {
			fmt.Fprintf(w, "%s%s = ", rangeDecl("so_rune", declare), name)
		}
		fmt.Fprint(w, "so_utf8_decode(")
		g.emitExpr(w, stmt.X)
		fmt.Fprintf(w, ", %s, &%s);\n", k.name, k.width)
	})
}

// emitRangeBody emits the body of a range loop and the closing brace. top emits
// the statements at the top of the body, before the Go statements. It is nil
// when the loop needs no such statements.
func (g *Generator) emitRangeBody(w io.Writer, stmt *ast.RangeStmt, key rangeKey, top func()) {
	g.emitRangeKeyCopy(w, key)
	if top != nil {
		g.state.depth++
		top()
		g.state.depth--
	}
	g.emitBlock(w, stmt.Body)
	fmt.Fprintf(w, "%s}\n", g.indent())
}

// rangeKey describes the C loop variable of a range statement.
type rangeKey struct {
	decl  string // type prefix for the init clause, with a trailing space
	name  string // name of the C loop variable
	copy  string // Go key variable to copy into, empty when the loop needs no copy
	width string // name of the rune width variable, used by a string range
}

// rangeKeyVar returns the C loop variable for a range statement.
//
// A define form ("for i := range x") declares the key in the init clause, so
// the loop counts on the key. An assign form ("for i = range x") must leave
// the key at the index of the last iteration, and a C loop variable stops one
// past it. So the loop counts on a variable of its own and copies it to the
// key at the top of the body.
func (g *Generator) rangeKeyVar(stmt *ast.RangeStmt) rangeKey {
	if stmt.Key == nil {
		// Basic form: "for range x".
		return rangeKey{decl: "so_int ", name: "_i", width: "_iw"}
	}
	key := stmt.Key.(*ast.Ident)
	width := "_" + key.Name + "w"
	switch {
	case key.Name == "_":
		return rangeKey{decl: "so_int ", name: "_", width: width}
	case stmt.Tok == token.ASSIGN:
		return rangeKey{decl: "so_int ", name: "_" + key.Name + "i", copy: key.Name, width: width}
	default:
		decl := g.mapTypeName(stmt, g.types.Defs[key].Type()) + " "
		return rangeKey{decl: decl, name: key.Name, width: width}
	}
}

// emitRangeKeyCopy writes the key assignment at the top of a range loop body.
func (g *Generator) emitRangeKeyCopy(w io.Writer, key rangeKey) {
	if key.copy == "" {
		return
	}
	g.state.depth++
	fmt.Fprintf(w, "%s%s = %s;\n", g.indent(), key.copy, key.name)
	g.state.depth--
}

// rangeVar returns the key or the value variable of a range statement. name is
// empty when the loop needs no such variable. declare reports whether the loop
// declares the variable (true) of assigns to an existing one (false).
func rangeVar(stmt *ast.RangeStmt, expr ast.Expr) (name string, declare bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident.Name == "_" {
		return "", false
	}
	return ident.Name, stmt.Tok != token.ASSIGN
}

// rangeDecl returns the type prefix of a range variable declaration.
// The prefix is empty for the assign form, which declares nothing.
func rangeDecl(typeName string, declare bool) string {
	if !declare {
		return ""
	}
	return typeName + " "
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
