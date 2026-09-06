package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"
)

// emitArrayLit emits a fixed-size array literal as a C initializer list.
// Example: [5]int{1, 2, 3, 4, 5} → {1, 2, 3, 4, 5}
func (g *Generator) emitArrayLit(w io.Writer, n *ast.CompositeLit) {
	elem := g.types.TypeOf(n).Underlying().(*types.Array).Elem()
	fmt.Fprint(w, "{")

	if hasKeyedElements(n) {
		g.emitSparseArrayValues(w, n, elem)
	} else {
		for i, elt := range n.Elts {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			g.checkArrayValue(elt, elem)
			g.emitExprAsType(w, n, elt, elem)
		}
	}

	fmt.Fprint(w, "}")
}

// emitArrayValue emits an array expression in a value position.
func (g *Generator) emitArrayValue(w io.Writer, node ast.Node, expr ast.Expr, arr *types.Array) {
	lit, isLit := ast.Unparen(expr).(*ast.CompositeLit)
	if !isLit {
		g.emitExpr(w, expr)
		return
	}
	// A composite literal needs a compound literal prefix (e.g. (so_int[3]){11, 22, 33}).
	fmt.Fprintf(w, "(%s%s)", g.mapTypeName(node, arr.Elem()), arrayDims(arr))
	g.emitExpr(w, lit)
}

// emitArrayCmpOperand emits an array comparison operand.
func (g *Generator) emitArrayCmpOperand(w io.Writer, expr ast.Expr, arr *types.Array) {
	if _, isLit := ast.Unparen(expr).(*ast.CompositeLit); !isLit {
		g.emitExpr(w, expr)
		return
	}
	// A compound literal needs extra parentheses to protect
	// against the preprocessor misinterpreting commas.
	fmt.Fprint(w, "(")
	g.emitArrayValue(w, expr, expr, arr)
	fmt.Fprint(w, ")")
}

// emitSliceLit emits a slice literal as a so_Slice compound literal.
// Example: []int{1, 2, 3, 4} → {(so_int[4]){1, 2, 3, 4}, 4, 4}
func (g *Generator) emitSliceLit(w io.Writer, n *ast.CompositeLit) {
	sl := g.types.TypeOf(n).Underlying().(*types.Slice)
	elemType := g.mapTypeName(n, sl.Elem())
	size := len(n.Elts)
	if size == 0 {
		fmt.Fprint(w, "(so_Slice){}")
		return
	}
	fmt.Fprintf(w, "(so_Slice){(%s[%d]){", elemType, size)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		g.emitExprAsType(w, n, elt, sl.Elem())
	}
	fmt.Fprintf(w, "}, %d, %d}", size, size)
}

// emitSparseArrayValues emits array values using C99 designated initializers
// for keyed elements. Example: [...]int{100, 3: 400, 500} → 100, [3] = 400, 500
func (g *Generator) emitSparseArrayValues(w io.Writer, n *ast.CompositeLit, elem types.Type) {
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fmt.Fprint(w, "[")
			g.emitExpr(w, kv.Key)
			fmt.Fprint(w, "] = ")
			g.checkArrayValue(kv.Value, elem)
			g.emitExprAsType(w, n, kv.Value, elem)
		} else {
			g.checkArrayValue(elt, elem)
			g.emitExprAsType(w, n, elt, elem)
		}
	}
}

// emitSliceExpr emits a slice expression (e.g. nums[1:4]).
// For arrays: so_array_slice(T, arr, low, high, size).
// For slices: so_slice(T, s, low, high).
func (g *Generator) emitSliceExpr(w io.Writer, n *ast.SliceExpr) {
	g.checkLocalScope(n)
	typ := g.types.TypeOf(n.X).Underlying()

	// Unwrap pointer-to-array: p[a:b] becomes (*p)[a:b].
	ptrDeref := false
	if ptr, ok := typ.(*types.Pointer); ok {
		if _, ok := ptr.Elem().Underlying().(*types.Array); ok {
			typ = ptr.Elem().Underlying()
			ptrDeref = true
		}
	}

	switch t := typ.(type) {
	case *types.Array:
		elemType := g.mapTypeName(n, t.Elem())
		if n.Slice3 {
			fmt.Fprintf(w, "so_array_slice3(%s, ", elemType)
		} else {
			fmt.Fprintf(w, "so_array_slice(%s, ", elemType)
		}
		if ptrDeref {
			fmt.Fprint(w, "(*")
			g.emitExpr(w, n.X)
			fmt.Fprint(w, ")")
		} else {
			g.emitExpr(w, n.X)
		}
		fmt.Fprint(w, ", ")
		if n.Low != nil {
			g.emitExpr(w, n.Low)
		} else {
			fmt.Fprint(w, "0")
		}
		fmt.Fprint(w, ", ")
		if n.High != nil {
			g.emitExpr(w, n.High)
		} else {
			fmt.Fprintf(w, "%d", t.Len())
		}
		if n.Slice3 {
			fmt.Fprint(w, ", ")
			g.emitExpr(w, n.Max)
			fmt.Fprint(w, ")")
		} else {
			fmt.Fprintf(w, ", %d)", t.Len())
		}

	case *types.Basic:
		if t.Kind() != types.String && t.Kind() != types.UntypedString {
			g.fail(n, "unsupported slice expression on basic type: %s", g.typeString(t))
			break
		}
		fmt.Fprint(w, "so_string_slice(")
		g.emitExpr(w, n.X)
		fmt.Fprint(w, ", ")
		if n.Low != nil {
			g.emitExpr(w, n.Low)
		} else {
			fmt.Fprint(w, "0")
		}
		fmt.Fprint(w, ", ")
		if n.High != nil {
			g.emitExpr(w, n.High)
		} else {
			g.emitPostfixOperand(w, n.X)
			fmt.Fprint(w, ".len")
		}
		fmt.Fprint(w, ")")

	case *types.Slice:
		elemType := g.mapTypeName(n, t.Elem())
		if n.Slice3 {
			fmt.Fprintf(w, "so_slice3(%s, ", elemType)
		} else {
			fmt.Fprintf(w, "so_slice(%s, ", elemType)
		}
		g.emitMacroArg(w, n.X)
		fmt.Fprint(w, ", ")
		if n.Low != nil {
			g.emitExpr(w, n.Low)
		} else {
			fmt.Fprint(w, "0")
		}
		fmt.Fprint(w, ", ")
		if n.High != nil {
			g.emitExpr(w, n.High)
		} else {
			g.emitPostfixOperand(w, n.X)
			fmt.Fprint(w, ".len")
		}
		if n.Slice3 {
			fmt.Fprint(w, ", ")
			g.emitExpr(w, n.Max)
		}
		fmt.Fprint(w, ")")

	default:
		g.fail(n, "unsupported slice expression type: %T", t)
	}
}

// emitArrayVarDecl emits a local array declaration with an initializer.
func (g *Generator) emitArrayVarDecl(w io.Writer, ct CType, name string, value ast.Expr) {
	if lit, ok := ast.Unparen(value).(*ast.CompositeLit); ok {
		// An array literal emits as a C brace initializer.
		fmt.Fprintf(w, "%s%s = ", g.indent(), ct.Decl(name))
		g.emitExpr(w, lit)
		fmt.Fprint(w, ";\n")
		return
	}
	g.emitArrayCopy(w, ct.Decl(name), name, func() { g.emitExpr(w, value) })
}

// emitArrayCopy declares an array variable and copies the value into it, e.g.:
// var a [3]int = b -> so_int a[3]; memcpy(a, b, sizeof(a)).
// An empty decl skips the declaration and copies into an existing variable.
func (g *Generator) emitArrayCopy(w io.Writer, decl, name string, emitValue func()) {
	if decl != "" {
		fmt.Fprintf(w, "%s%s;\n", g.indent(), decl)
	}
	fmt.Fprintf(w, "%smemcpy(%s, ", g.indent(), name)
	emitValue()
	fmt.Fprintf(w, ", sizeof(%s));\n", name)
}

// checkArrayConv rejects a conversion between array types.
func (g *Generator) checkArrayConv(n *ast.CallExpr, target types.Type, arg ast.Expr) {
	if _, ok := target.Underlying().(*types.Array); !ok {
		return
	}
	g.fail(n, "cannot convert %s to %s; copy the elements instead",
		g.typeString(g.types.TypeOf(arg)), g.typeString(target))
}

// checkArrayValue fails if expr initializes an array-typed target
// with anything other than an array literal. C only allows a brace
// initializer there, so an array value has no valid translation.
func (g *Generator) checkArrayValue(expr ast.Expr, targetType types.Type) {
	if _, ok := targetType.Underlying().(*types.Array); !ok {
		return
	}
	if _, ok := ast.Unparen(expr).(*ast.CompositeLit); ok {
		return
	}
	g.fail(expr, "cannot use an array value in a composite literal: use an array literal or assign the field separately")
}

// checkSliceElemType fails if the slice element type is not supported in C.
// Unwraps nested slices, since [][][3]int is as unsupported as [][3]int.
func (g *Generator) checkSliceElemType(node ast.Node, elem types.Type) {
	seen := make(map[*types.TypeName]bool)
	for {
		// A named element type can close a cycle, as in `type S []S`.
		if named, ok := types.Unalias(elem).(*types.Named); ok {
			if seen[named.Obj()] {
				return
			}
			seen[named.Obj()] = true
		}
		sl, ok := elem.Underlying().(*types.Slice)
		if !ok {
			break
		}
		elem = sl.Elem()
	}
	if _, ok := elem.Underlying().(*types.Array); ok {
		g.fail(node, "slice of arrays is not supported")
	}
}

// arrayType returns the array type of t, named or not.
func arrayType(t types.Type) (*types.Array, bool) {
	if t == nil {
		return nil, false
	}
	arr, ok := t.Underlying().(*types.Array)
	return arr, ok
}

// hasKeyedElements returns true if any element
// in the composite literal uses key:value syntax.
func hasKeyedElements(n *ast.CompositeLit) bool {
	for _, elt := range n.Elts {
		if _, ok := elt.(*ast.KeyValueExpr); ok {
			return true
		}
	}
	return false
}

// arrayDims returns the C dimension suffix for an array type.
// [3]int -> "[3]", [2][3]int -> "[2][3]", non-array -> "".
// Named types return "" because their typedef already includes the dimensions.
func arrayDims(typ types.Type) string {
	typ = types.Unalias(typ)
	if _, ok := typ.(*types.Named); ok {
		return ""
	}
	var dims strings.Builder
	for arr, ok := typ.(*types.Array); ok; arr, ok = arr.Elem().(*types.Array) {
		fmt.Fprintf(&dims, "[%d]", arr.Len())
	}
	return dims.String()
}

// arraySize returns the compile-time size of an array type, or -1 if not an array.
// For a pointer to an array type, returns the size of the pointed-to array.
func arraySize(typ types.Type) int64 {
	t := typ.Underlying()
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem().Underlying()
	}
	if arr, ok := t.(*types.Array); ok {
		return arr.Len()
	}
	return -1
}
