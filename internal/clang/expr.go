package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// emitExpr dispatches expression generation to per-type methods.
func (g *Generator) emitExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		g.emitBasicLit(e)
	case *ast.BinaryExpr:
		g.emitBinaryExpr(e)
	case *ast.CallExpr:
		g.emitCallExpr(e)
	case *ast.CompositeLit:
		g.emitCompositeLit(e)
	case *ast.Ident:
		g.emitIdent(e)
	case *ast.IndexExpr:
		g.emitIndexExpr(e)
	case *ast.ParenExpr:
		g.emitParenExpr(e)
	case *ast.SelectorExpr:
		g.emitSelectorExpr(e)
	case *ast.SliceExpr:
		g.emitSliceExpr(e)
	case *ast.StarExpr:
		g.emitStarExpr(e)
	case *ast.TypeAssertExpr:
		g.emitTypeAssertExpr(e)
	case *ast.UnaryExpr:
		g.emitUnaryExpr(e)
	default:
		g.fail(expr, "unsupported expression type: %T", expr)
	}
}

// emitBasicLit emits a literal.
func (g *Generator) emitBasicLit(n *ast.BasicLit) {
	if n.Kind == token.STRING {
		g.emitStringLit(n)
		return
	}
	if n.Kind == token.CHAR {
		if basic, ok := g.types.TypeOf(n).(*types.Basic); ok && basic.Kind() == types.Byte {
			fmt.Fprintf(g.state.writer, "%s", n.Value)
		} else {
			fmt.Fprintf(g.state.writer, "U%s", n.Value)
		}
		return
	}
	g.emitNumericLit(n)
}

// emitNumericLit emits a numeric literal, converting Go-specific formats to C.
func (g *Generator) emitNumericLit(n *ast.BasicLit) {
	val := strings.ReplaceAll(n.Value, "_", "")
	if n.Kind == token.INT && (strings.HasPrefix(val, "0o") || strings.HasPrefix(val, "0O")) {
		val = "0" + val[2:]
	}
	fmt.Fprintf(g.state.writer, "%s", val)
}

// emitBinaryExpr emits a binary expression.
func (g *Generator) emitBinaryExpr(n *ast.BinaryExpr) {
	w := g.state.writer
	// String comparisons: emit so_string_eq/ne/lt/gt/lte/gte calls.
	if isCompare(n.Op) {
		if basic, ok := g.types.TypeOf(n.X).Underlying().(*types.Basic); ok && basic.Kind() == types.String {
			fmt.Fprintf(w, "%s(", stringCompareFunc(n.Op))
			g.emitExpr(n.X)
			fmt.Fprintf(w, ", ")
			g.emitExpr(n.Y)
			fmt.Fprintf(w, ")")
			return
		}
	}
	// Go's &^ (AND NOT) has no C equivalent — emit & ~ instead.
	if n.Op == token.AND_NOT {
		g.emitExpr(n.X)
		fmt.Fprintf(w, " & ~")
		g.emitExpr(n.Y)
		return
	}
	// Regular binary expression.
	g.emitExpr(n.X)
	fmt.Fprintf(w, " %s ", n.Op.String())
	g.emitExpr(n.Y)
}

// emitCallExpr emits a function call or type conversion.
func (g *Generator) emitCallExpr(n *ast.CallExpr) {
	w := g.state.writer
	if tv, ok := g.types.Types[n.Fun]; ok && tv.IsType() {
		if isInterfaceType(tv.Type) {
			iface := tv.Type.Underlying().(*types.Interface)
			if iface.Empty() {
				g.emitAnyValue(n, n.Args[0])
				return
			}
			// Named non-empty interface conversion (e.g. Shape(r)).
			g.emitInterfaceLit(tv.Type, n.Args[0])
			return
		}
		// Type conversion (e.g. int(3.14)).
		cType := g.mapType(n, tv.Type)
		fmt.Fprintf(w, "(%s)", cType)
		g.emitExpr(n.Args[0])
		return
	}

	// Method call (e.g. r.Area()).
	if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
		if selection, ok := g.types.Selections[sel]; ok && selection.Kind() == types.MethodVal {
			g.emitMethodCall(sel, n.Args)
			return
		}
	}

	// errors.New("msg") → so_error("msg")
	// Special case, because there're no builtins for creating errors in Go.
	// The only way to create an error is stdlib call, but we want to emit it
	// as a builtin so_error call.
	if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			if pkgName, ok := g.types.Uses[ident].(*types.PkgName); ok && pkgName.Imported().Path() == "errors" && sel.Sel.Name == "New" {
				arg := n.Args[0].(*ast.BasicLit)
				fmt.Fprintf(w, "so_error(%s)", arg.Value)
				return
			}
		}
	}

	// Regular function call.
	g.emitFuncCall(n)
}

// emitCompositeLit emits a composite literal (struct or array initialization).
// Fields can be positional (Point{1, 2}) or named (Point{x: 1, x: 2}).
func (g *Generator) emitCompositeLit(n *ast.CompositeLit) {
	if st, ok := n.Type.(*ast.StructType); ok {
		g.emitAnonStructLit(n, st)
		return
	}

	switch g.types.TypeOf(n).Underlying().(type) {
	case *types.Array:
		g.emitArrayLit(n)
		return
	case *types.Slice:
		g.emitSliceLit(n)
		return
	}

	// Regular composite literal.
	g.emitStructLit(n)
}

// emitIdent emits an identifier.
func (g *Generator) emitIdent(n *ast.Ident) {
	name := n.Name
	if name == "nil" {
		fmt.Fprintf(g.state.writer, "NULL")
		return
	}
	if obj := g.types.Uses[n]; obj != nil {
		if ast.IsExported(name) && obj.Parent() == g.pkg.Types.Scope() {
			// Exported package-level declarations are prefixed
			// with the package name (e.g. RectArea → geom_RectArea).
			name = g.symbolName(name)
		}
	}
	fmt.Fprintf(g.state.writer, "%s", name)
}

// emitParenExpr emits a parenthesized expression.
func (g *Generator) emitParenExpr(n *ast.ParenExpr) {
	fmt.Fprintf(g.state.writer, "(")
	g.emitExpr(n.X)
	fmt.Fprintf(g.state.writer, ")")
}

// emitSelectorExpr emits a selector expression (e.g. geom.RectArea → geom_RectArea, or p.name).
func (g *Generator) emitSelectorExpr(n *ast.SelectorExpr) {
	if ident, ok := n.X.(*ast.Ident); ok {
		if pkgName, ok := g.types.Uses[ident].(*types.PkgName); ok {
			// Imported symbols are prefixed with the
			// package name (e.g. fmt.Println → fmt_Println).
			fmt.Fprintf(g.state.writer, "%s_%s", pkgName.Name(), n.Sel.Name)
			return
		}
	}

	// Struct/interface field access.
	w := g.state.writer
	xType := g.types.TypeOf(n.X)
	g.emitExpr(n.X)

	// Value receivers (T x) are passed as (void* self) and coverted to (T* x),
	// so need to use "->" instead of "." for field access.
	_, isPtr := xType.Underlying().(*types.Pointer)
	isValueRecv := false
	if ident, ok := n.X.(*ast.Ident); ok && ident.Name == g.state.recvName {
		isValueRecv = g.state.recvName != ""
	}
	// Pointers and value receivers use "->", regular values use ".".
	if isPtr || isValueRecv {
		fmt.Fprintf(w, "->%s", n.Sel.Name)
	} else {
		fmt.Fprintf(w, ".%s", n.Sel.Name)
	}
}

// emitStarExpr emits a dereference expression (e.g. *p).
func (g *Generator) emitStarExpr(n *ast.StarExpr) {
	fmt.Fprintf(g.state.writer, "*")
	g.emitExpr(n.X)
}

// emitIndexExpr emits an index expression (e.g. a[4]) as so_index(a, int, 4).
func (g *Generator) emitIndexExpr(n *ast.IndexExpr) {
	w := g.state.writer

	// Determine the element type of the array/slice.
	var elemType string
	switch t := g.types.TypeOf(n.X).Underlying().(type) {
	case *types.Array:
		elemType = g.mapType(n, t.Elem())
	case *types.Slice:
		elemType = g.mapType(n, t.Elem())
	case *types.Basic:
		if t.Kind() == types.String {
			elemType = "uint8_t"
		} else {
			g.fail(n, "unsupported index expression type: %T", t)
		}
	default:
		g.fail(n, "unsupported index expression type: %T", t)
	}

	// Emit the index expression as so_index(x, elemType, index).
	fmt.Fprintf(w, "so_index(")
	g.emitExpr(n.X)
	fmt.Fprintf(w, ", %s, ", elemType)
	g.emitExpr(n.Index)
	fmt.Fprintf(w, ")")
}

// emitUnaryExpr emits a unary expression.
func (g *Generator) emitUnaryExpr(n *ast.UnaryExpr) {
	w := g.state.writer
	if n.Op == token.AND {
		if cl, ok := n.X.(*ast.CompositeLit); ok {
			// &Person{...} → &(Person){...}
			typ := g.types.TypeOf(cl.Type)
			cType := g.mapType(n, typ)
			fmt.Fprintf(w, "&(%s)", cType)
			g.emitCompositeLit(cl)
			return
		}
	}
	fmt.Fprintf(w, "%s", n.Op.String())
	g.emitExpr(n.X)
}

func isCompare(op token.Token) bool {
	switch op {
	case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
		return true
	}
	return false
}
