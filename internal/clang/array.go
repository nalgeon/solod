package clang

import (
	"fmt"
	"go/ast"
	"go/types"
)

// emitArrayLit emits a fixed-size array literal as a C initializer list.
// Example: [5]int{1, 2, 3, 4, 5} → {1, 2, 3, 4, 5}
func (g *Generator) emitArrayLit(n *ast.CompositeLit) {
	w := g.state.writer
	fmt.Fprintf(w, "{")

	if hasKeyedElements(n) {
		g.emitSparseArrayValues(n)
	} else {
		for i, elt := range n.Elts {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			g.emitExpr(elt)
		}
	}

	fmt.Fprintf(w, "}")
}

// emitArrayCmpOperand emits an array comparison operand.
// Composite literals need a C compound literal prefix (e.g. (so_int[3]){...})
// wrapped in extra parentheses so commas inside braces don't split macro args.
func (g *Generator) emitArrayCmpOperand(expr ast.Expr, arr *types.Array) {
	w := g.state.writer
	if _, isLit := expr.(*ast.CompositeLit); isLit {
		elemType := g.mapType(expr, arr.Elem())
		fmt.Fprintf(w, "((%s%s)", elemType, arrayDims(arr))
		g.emitExpr(expr)
		fmt.Fprintf(w, ")")
		return
	}
	g.emitExpr(expr)
}

// emitSliceLit emits a slice literal as a so_Slice compound literal.
// Example: []int{1, 2, 3, 4} → {(so_int[4]){1, 2, 3, 4}, 4, 4}
func (g *Generator) emitSliceLit(n *ast.CompositeLit) {
	w := g.state.writer
	sl := g.types.TypeOf(n).Underlying().(*types.Slice)
	elemType := g.mapType(n, sl.Elem())
	size := len(n.Elts)
	if size == 0 {
		fmt.Fprintf(w, "(so_Slice){0}")
		return
	}
	fmt.Fprintf(w, "(so_Slice){(%s[%d]){", elemType, size)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		g.emitExpr(elt)
	}
	fmt.Fprintf(w, "}, %d, %d}", size, size)
}

// emitSparseArrayValues emits array values using C99 designated initializers
// for keyed elements. Example: [...]int{100, 3: 400, 500} → 100, [3] = 400, 500
func (g *Generator) emitSparseArrayValues(n *ast.CompositeLit) {
	w := g.state.writer
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fmt.Fprintf(w, "[")
			g.emitExpr(kv.Key)
			fmt.Fprintf(w, "] = ")
			g.emitExpr(kv.Value)
		} else {
			g.emitExpr(elt)
		}
	}
}

// emitSliceExpr emits a slice expression (e.g. nums[1:4]).
// For arrays: so_array_slice(T, arr, low, high, size).
// For slices: so_slice(T, s, low, high).
func (g *Generator) emitSliceExpr(n *ast.SliceExpr) {
	w := g.state.writer

	switch t := g.types.TypeOf(n.X).Underlying().(type) {
	case *types.Array:
		elemType := g.mapType(n, t.Elem())
		if n.Slice3 {
			fmt.Fprintf(w, "so_array_slice3(%s, ", elemType)
		} else {
			fmt.Fprintf(w, "so_array_slice(%s, ", elemType)
		}
		g.emitExpr(n.X)
		fmt.Fprintf(w, ", ")
		if n.Low != nil {
			g.emitExpr(n.Low)
		} else {
			fmt.Fprintf(w, "0")
		}
		fmt.Fprintf(w, ", ")
		if n.High != nil {
			g.emitExpr(n.High)
		} else {
			fmt.Fprintf(w, "%d", t.Len())
		}
		if n.Slice3 {
			fmt.Fprintf(w, ", ")
			g.emitExpr(n.Max)
			fmt.Fprintf(w, ")")
		} else {
			fmt.Fprintf(w, ", %d)", t.Len())
		}

	case *types.Basic:
		if t.Kind() != types.String {
			g.fail(n, "unsupported slice expression on basic type: %s", t)
			break
		}
		fmt.Fprintf(w, "so_string_slice(")
		g.emitExpr(n.X)
		fmt.Fprintf(w, ", ")
		if n.Low != nil {
			g.emitExpr(n.Low)
		} else {
			fmt.Fprintf(w, "0")
		}
		fmt.Fprintf(w, ", ")
		if n.High != nil {
			g.emitExpr(n.High)
		} else {
			g.emitExpr(n.X)
			fmt.Fprintf(w, ".len")
		}
		fmt.Fprintf(w, ")")

	case *types.Slice:
		elemType := g.mapType(n, t.Elem())
		if n.Slice3 {
			fmt.Fprintf(w, "so_slice3(%s, ", elemType)
		} else {
			fmt.Fprintf(w, "so_slice(%s, ", elemType)
		}
		g.emitExpr(n.X)
		fmt.Fprintf(w, ", ")
		if n.Low != nil {
			g.emitExpr(n.Low)
		} else {
			fmt.Fprintf(w, "0")
		}
		fmt.Fprintf(w, ", ")
		if n.High != nil {
			g.emitExpr(n.High)
		} else {
			g.emitExpr(n.X)
			fmt.Fprintf(w, ".len")
		}
		if n.Slice3 {
			fmt.Fprintf(w, ", ")
			g.emitExpr(n.Max)
		}
		fmt.Fprintf(w, ")")

	default:
		g.fail(n, "unsupported slice expression type: %T", t)
	}
}

// isArrayType reports whether a type has array dimensions.
func isArrayType(typ types.Type) bool {
	return arrayDims(typ) != ""
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
	var dims string
	for arr, ok := typ.(*types.Array); ok; arr, ok = arr.Elem().(*types.Array) {
		dims += fmt.Sprintf("[%d]", arr.Len())
	}
	return dims
}

// arraySize returns the compile-time size of an array type, or -1 if not an array.
func arraySize(typ types.Type) int64 {
	if arr, ok := typ.Underlying().(*types.Array); ok {
		return arr.Len()
	}
	return -1
}
