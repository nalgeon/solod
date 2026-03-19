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
		if g.hasStringType(n.X) {
			fmt.Fprintf(w, "%s(", stringCompareFunc(n.Op))
			g.emitExpr(n.X)
			fmt.Fprintf(w, ", ")
			g.emitExpr(n.Y)
			fmt.Fprintf(w, ")")
			return
		}
	}

	// String addition: only literal concatenation is supported.
	if n.Op == token.ADD || n.Op == token.ADD_ASSIGN {
		if g.hasStringType(n.X) {
			if n.Op == token.ADD && isStringLit(n.X) && isStringLit(n.Y) {
				fmt.Fprintf(w, "so_str(")
				g.emitStringLitConcat(n)
				fmt.Fprintf(w, ")")
				return
			}
			g.fail(n, "string addition is not supported")
		}
	}

	// Interface nil comparisons: emit iface.self == NULL / != NULL.
	if n.Op == token.EQL || n.Op == token.NEQ {
		if isNamedNonEmptyInterface(g.types.TypeOf(n.X)) && isNilType(g.types.TypeOf(n.Y)) {
			g.emitExpr(n.X)
			fmt.Fprintf(w, ".self %s NULL", n.Op.String())
			return
		}
	}

	// Slice nil comparisons: emit s.ptr == NULL / != NULL.
	if n.Op == token.EQL || n.Op == token.NEQ {
		if _, ok := g.types.TypeOf(n.X).Underlying().(*types.Slice); ok && isNilType(g.types.TypeOf(n.Y)) {
			g.emitExpr(n.X)
			fmt.Fprintf(w, ".ptr %s NULL", n.Op.String())
			return
		}
	}

	// Array comparisons: emit so_array_eq/ne calls.
	if n.Op == token.EQL || n.Op == token.NEQ {
		if arr, ok := g.types.TypeOf(n.X).Underlying().(*types.Array); ok {
			if n.Op == token.EQL {
				fmt.Fprintf(w, "so_array_eq(")
			} else {
				fmt.Fprintf(w, "so_array_ne(")
			}
			g.emitArrayCmpOperand(n.X, arr)
			fmt.Fprintf(w, ", ")
			g.emitArrayCmpOperand(n.Y, arr)
			elemType := g.mapType(n, arr.Elem())
			fmt.Fprintf(w, ", %d * sizeof(%s))", arr.Len(), elemType)
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

	// Generic function call with explicit type argument (e.g. fn[T](a) or pkg.Fn[T](a)).
	if indexExpr, ok := n.Fun.(*ast.IndexExpr); ok {
		if ident := exprIdent(indexExpr.X); ident != nil {
			if inst, ok := g.types.Instances[ident]; ok && inst.TypeArgs.Len() > 0 {
				g.emitGenericCall(n, indexExpr.X, inst)
				return
			}
		}
	}

	// Generic function call with inferred type argument (e.g. fn(a) where fn is generic).
	if ident := exprIdent(n.Fun); ident != nil {
		if inst, ok := g.types.Instances[ident]; ok && inst.TypeArgs.Len() > 0 {
			g.emitGenericCall(n, n.Fun, inst)
			return
		}
	}

	if tv, ok := g.types.Types[n.Fun]; ok && tv.IsType() {
		// Convert value to an interface type (e.g. Shape(r)).
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
		// String-to-slice conversion ([]byte(s) or []rune(s)).
		if sl, ok := tv.Type.Underlying().(*types.Slice); ok {
			if g.hasStringType(n.Args[0]) {
				g.emitSliceCast(n, sl)
				return
			}
		}
		// Slice/byte/rune-to-string conversion.
		if basic, ok := tv.Type.Underlying().(*types.Basic); ok && basic.Kind() == types.String {
			argType := g.types.TypeOf(n.Args[0])
			if sl, ok := argType.Underlying().(*types.Slice); ok {
				g.emitStringCast(n, sl)
				return
			}
			if argBasic, ok := argType.Underlying().(*types.Basic); ok {
				switch argBasic.Kind() {
				case types.Byte:
					fmt.Fprintf(w, "so_byte_string(")
					g.emitExpr(n.Args[0])
					fmt.Fprintf(w, ")")
					return
				case types.Int32:
					fmt.Fprintf(w, "so_rune_string(")
					g.emitExpr(n.Args[0])
					fmt.Fprintf(w, ")")
					return
				}
			}
		}
		// Regular type conversion (e.g. int(3.14)).
		cType := g.mapType(n, tv.Type)
		fmt.Fprintf(w, "(%s)(", cType)
		g.emitExpr(n.Args[0])
		fmt.Fprintf(w, ")")
		return
	}

	// Method call (e.g. r.Area()).
	if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
		if selection, ok := g.types.Selections[sel]; ok && selection.Kind() == types.MethodVal {
			g.emitMethodCall(sel, n.Args)
			return
		}
	}

	// Regular function call.
	g.emitFuncCall(n)
}

// emitGenericCall emits a generic function call as fn(T, a, b),
// where type arguments are prepended to the regular arguments.
func (g *Generator) emitGenericCall(n *ast.CallExpr, fun ast.Expr, inst types.Instance) {
	w := g.state.writer
	g.emitExpr(fun)
	fmt.Fprintf(w, "(%s", g.mapType(n, inst.TypeArgs.At(0)))
	for i := 1; i < inst.TypeArgs.Len(); i++ {
		fmt.Fprintf(w, ", %s", g.mapType(n, inst.TypeArgs.At(i)))
	}
	sig, _ := inst.Type.(*types.Signature)
	for i, arg := range n.Args {
		fmt.Fprintf(w, ", ")
		if sig != nil && i < sig.Params().Len() {
			g.emitExprAsType(n, arg, sig.Params().At(i).Type())
		} else {
			g.emitExpr(arg)
		}
	}
	fmt.Fprintf(w, ")")
}

// emitSliceCast emits a string-to-slice conversion ([]byte(s) or []rune(s)).
func (g *Generator) emitSliceCast(call *ast.CallExpr, sl *types.Slice) {
	w := g.state.writer
	elem := sl.Elem().(*types.Basic)
	switch elem.Kind() {
	case types.Byte:
		fmt.Fprintf(w, "so_string_bytes(")
		g.emitExpr(call.Args[0])
		fmt.Fprintf(w, ")")
	case types.Int32:
		fmt.Fprintf(w, "so_string_runes(")
		g.emitExpr(call.Args[0])
		fmt.Fprintf(w, ", so_len(")
		g.emitExpr(call.Args[0])
		fmt.Fprintf(w, "))")
	}
}

// emitStringCast emits a slice-to-string conversion (string(bs) or string(rs)).
func (g *Generator) emitStringCast(call *ast.CallExpr, sl *types.Slice) {
	w := g.state.writer
	elem := sl.Elem().(*types.Basic)
	switch elem.Kind() {
	case types.Byte:
		fmt.Fprintf(w, "so_bytes_string(")
		g.emitExpr(call.Args[0])
		fmt.Fprintf(w, ")")
	case types.Int32:
		fmt.Fprintf(w, "so_runes_string(")
		g.emitExpr(call.Args[0])
		fmt.Fprintf(w, ")")
	default:
		g.fail(call, "unsupported slice-to-string conversion: %s", elem)
	}
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

	_, isPtr := xType.Underlying().(*types.Pointer)
	if isPtr {
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

// emitIndexExpr emits an index expression.
// For arrays: a[i] directly. For slices/strings: so_at(T, s, i).
func (g *Generator) emitIndexExpr(n *ast.IndexExpr) {
	w := g.state.writer

	// Arrays use direct C indexing.
	if _, ok := g.types.TypeOf(n.X).Underlying().(*types.Array); ok {
		g.emitExpr(n.X)
		fmt.Fprintf(w, "[")
		g.emitExpr(n.Index)
		fmt.Fprintf(w, "]")
		return
	}

	// Pointer-to-array: p[i] becomes (*p)[i].
	if ptr, ok := g.types.TypeOf(n.X).Underlying().(*types.Pointer); ok {
		if _, ok := ptr.Elem().Underlying().(*types.Array); ok {
			fmt.Fprintf(w, "(*")
			g.emitExpr(n.X)
			fmt.Fprintf(w, ")[")
			g.emitExpr(n.Index)
			fmt.Fprintf(w, "]")
			return
		}
	}

	// Slices and strings use so_at.
	var elemType string
	switch t := g.types.TypeOf(n.X).Underlying().(type) {
	case *types.Slice:
		elemType = g.mapType(n, t.Elem())
	case *types.Basic:
		if t.Kind() == types.String || t.Kind() == types.UntypedString {
			elemType = "so_byte"
		} else {
			g.fail(n, "unsupported index expression type: %T", t)
		}
	default:
		g.fail(n, "unsupported index expression type: %T", t)
	}

	fmt.Fprintf(w, "so_at(%s, ", elemType)
	g.emitExpr(n.X)
	fmt.Fprintf(w, ", ")
	g.emitExpr(n.Index)
	fmt.Fprintf(w, ")")
}

// emitUnaryExpr emits a unary expression.
func (g *Generator) emitUnaryExpr(n *ast.UnaryExpr) {
	w := g.state.writer
	if n.Op == token.AND {
		// &arrayParam: C array params decay to pointers, so &param
		// gives T** instead of T(*)[N]. Emit a cast instead.
		if ident, ok := n.X.(*ast.Ident); ok {
			if _, ok := g.types.TypeOf(n.X).Underlying().(*types.Array); ok {
				if g.isArrayParam(ident) {
					ct := g.mapCType(n, g.types.TypeOf(n))
					fmt.Fprintf(w, "(%s)", ct.Decl(""))
					g.emitExpr(n.X)
					return
				}
			}
		}
		if _, ok := n.X.(*ast.CompositeLit); ok {
			// &Person{...} → &(Person){...}
			fmt.Fprintf(w, "&")
			g.emitExpr(n.X)
			return
		}
	}
	if n.Op == token.XOR {
		fmt.Fprintf(w, "~")
		g.emitExpr(n.X)
		return
	}
	fmt.Fprintf(w, "%s", n.Op.String())
	g.emitExpr(n.X)
}

// emitExprAsType emits an expression as a specific type, handling special cases
// like interface conversions and nil assignments.
func (g *Generator) emitExprAsType(node ast.Node, expr ast.Expr, targetType types.Type) {
	// Empty interface: emit as void*.
	if iface, ok := targetType.Underlying().(*types.Interface); ok && iface.Empty() {
		g.emitAnyValue(node, expr)
		return
	}
	// Named interface conversion: wrap concrete types as interface literals.
	if isNamedNonEmptyInterface(targetType) {
		valType := g.types.TypeOf(expr)
		if isNilType(valType) {
			cType := g.mapType(node, targetType)
			fmt.Fprintf(g.state.writer, "(%s){0}", cType)
			return
		}
		if isConcreteNamedType(valType) {
			g.emitInterfaceLit(targetType, expr)
			return
		}
	}
	// Slice nil assignment: emit zero-initialized struct instead of NULL.
	if _, ok := targetType.Underlying().(*types.Slice); ok && isNilType(g.types.TypeOf(expr)) {
		fmt.Fprintf(g.state.writer, "(so_Slice){0}")
		return
	}
	g.emitExpr(expr)
}

// needsVoidParens reports whether expr needs parentheses in a (void) cast.
func (g *Generator) needsVoidParens(expr ast.Expr) bool {
	// Binary expressions need wrapping in case we want to cast them,
	// because the cast has higher precedence than binary operators.
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		// Not a binary expression — no need for parentheses.
		return false
	}
	if isCompare(bin.Op) {
		// String comparisons are emitted as function calls and don't need wrapping.
		if g.hasStringType(bin.X) {
			return false
		}
		// Array comparisons are emitted as function calls and don't need wrapping.
		if _, ok := g.types.TypeOf(bin.X).Underlying().(*types.Array); ok {
			return false
		}
	}
	// Binary expression - needs parentheses.
	return true
}

// isArrayParam reports whether ident refers to a function parameter.
func (g *Generator) isArrayParam(ident *ast.Ident) bool {
	if g.state.funcSig == nil {
		return false
	}
	obj := g.types.ObjectOf(ident)
	params := g.state.funcSig.Params()
	for param := range params.Variables() {
		if param == obj {
			return true
		}
	}
	return false
}

// isCompare reports whether a token is a comparison operator.
func isCompare(op token.Token) bool {
	switch op {
	case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
		return true
	}
	return false
}

// exprIdent returns the leaf *ast.Ident of an expression
// (the ident itself, or the Sel of a SelectorExpr).
func exprIdent(expr ast.Expr) *ast.Ident {
	switch e := expr.(type) {
	case *ast.Ident:
		return e
	case *ast.SelectorExpr:
		return e.Sel
	}
	return nil
}

// containsIota reports whether an expression contains the iota identifier.
func containsIota(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "iota" {
			found = true
			return false
		}
		return !found
	})
	return found
}
