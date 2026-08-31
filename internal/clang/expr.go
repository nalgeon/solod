package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"strings"
)

// emitExpr dispatches expression generation to per-type methods.
func (g *Generator) emitExpr(w io.Writer, expr ast.Expr) {
	// A computed constant expression is emitted as its value.
	if g.emitFloatConst(w, expr) || g.emitIntConst(w, expr) {
		return
	}
	switch e := expr.(type) {
	case *ast.BasicLit:
		g.emitBasicLit(w, e)
	case *ast.BinaryExpr:
		g.emitBinaryExpr(w, e)
	case *ast.CallExpr:
		g.emitCallExpr(w, e)
	case *ast.CompositeLit:
		g.emitCompositeLit(w, e)
	case *ast.Ident:
		g.emitIdent(w, e)
	case *ast.IndexExpr:
		g.emitIndexExpr(w, e)
	case *ast.ParenExpr:
		g.emitParenExpr(w, e.X)
	case *ast.SelectorExpr:
		g.emitSelectorExpr(w, e)
	case *ast.SliceExpr:
		g.emitSliceExpr(w, e)
	case *ast.StarExpr:
		g.emitStarExpr(w, e)
	case *ast.TypeAssertExpr:
		g.emitTypeAssertExpr(w, e)
	case *ast.UnaryExpr:
		g.emitUnaryExpr(w, e)
	default:
		g.fail(expr, "unsupported expression type: %T", expr)
	}
}

// emitBasicLit emits a literal.
func (g *Generator) emitBasicLit(w io.Writer, n *ast.BasicLit) {
	if n.Kind == token.STRING {
		g.emitStringLit(w, n)
		return
	}
	if n.Kind == token.CHAR {
		g.emitCharLit(w, n)
		return
	}
	g.emitNumericLit(w, n)
}

// emitNumericLit emits a numeric literal, converting Go-specific formats to C.
func (g *Generator) emitNumericLit(w io.Writer, n *ast.BasicLit) {
	tv := g.types.Types[n]
	val := strings.ReplaceAll(n.Value, "_", "")
	if isFloatType(tv.Type) {
		g.emitFloatLit(w, n, val, tv)
		return
	}
	if n.Kind == token.FLOAT {
		// The literal is a float in an integer context, for example 1e9 in
		// sec * 1e9 with an int64 sec. C reads 1e9 as a double and promotes
		// the other operand to match, which loses the low bits of an int64.
		g.emitIntLitOfFloat(w, n, tv)
		return
	}
	if n.Kind == token.INT && (strings.HasPrefix(val, "0o") || strings.HasPrefix(val, "0O")) {
		val = "0" + val[2:]
	}
	fmt.Fprint(w, val)
	if n.Kind == token.INT && exceedsInt64(tv.Value) {
		fmt.Fprint(w, "u") // unsigned suffix for C
	}
}

// emitBinaryExpr emits a binary expression.
func (g *Generator) emitBinaryExpr(w io.Writer, n *ast.BinaryExpr) {
	// Arithmetic on a type narrower than int: wrap the result
	// at the width of the Go type.
	if typ, ok := g.narrowCast(n); ok {
		fmt.Fprintf(w, "(%s)(", typ)
		defer fmt.Fprint(w, ")")
	}

	// Equality comparison.
	if n.Op == token.EQL || n.Op == token.NEQ {
		g.emitEqual(w, n)
		return
	}

	// String ordering comparison: emit so_string_lt/gt/lte/gte calls.
	if isOrderCompare(n.Op) && g.hasStringType(n.X) {
		fmt.Fprintf(w, "%s(", stringCompareFunc(n.Op))
		g.emitExpr(w, n.X)
		fmt.Fprint(w, ", ")
		g.emitExpr(w, n.Y)
		fmt.Fprint(w, ")")
		return
	}

	// String addition.
	if n.Op == token.ADD && g.hasStringType(n.X) {
		if isStringLit(g.types, n.X) && isStringLit(g.types, n.Y) {
			fmt.Fprint(w, "so_str(")
			g.emitStringLitConcat(w, n)
			fmt.Fprint(w, ")")
			return
		}
		fmt.Fprint(w, "so_string_add(")
		g.emitExpr(w, n.X)
		fmt.Fprint(w, ", ")
		g.emitExpr(w, n.Y)
		fmt.Fprint(w, ")")
		return
	}

	// Shift expression: parenthesize because Go's << >> have multiplicative
	// precedence, but C's << >> are below additive (+/-).
	if isShift(n.Op) {
		fmt.Fprint(w, "(")
		g.emitShiftOperand(w, n)
		fmt.Fprintf(w, " %s ", n.Op.String())
		g.emitExpr(w, n.Y)
		fmt.Fprint(w, ")")
		return
	}

	// Go's &^ (AND NOT) has no C equivalent — emit & ~ instead.
	if n.Op == token.AND_NOT {
		fmt.Fprint(w, "(")
		g.emitExpr(w, n.X)
		fmt.Fprint(w, " & ~")
		g.emitExpr(w, n.Y)
		fmt.Fprint(w, ")")
		return
	}

	// Bitwise operators: parenthesize because Go's & has multiplicative
	// precedence (same as * and <<), but C's & is below additive (+/-).
	// Similarly, Go's | and ^ have additive precedence, but in C they
	// are below & and +. Without parentheses, expressions like
	// a & b + c would mean (a & b) + c in Go but a & (b + c) in C.
	if n.Op == token.AND || n.Op == token.OR || n.Op == token.XOR {
		fmt.Fprint(w, "(")
		g.emitExpr(w, n.X)
		fmt.Fprintf(w, " %s ", n.Op.String())
		g.emitExpr(w, n.Y)
		fmt.Fprint(w, ")")
		return
	}

	// Integer division/modulo: guard against a zero divisor.
	if (n.Op == token.QUO || n.Op == token.REM) && g.needsIntDivGuard(n.X, n.Y) {
		if n.Op == token.QUO {
			fmt.Fprint(w, "so_div(")
		} else {
			fmt.Fprint(w, "so_mod(")
		}
		g.emitExpr(w, n.X)
		fmt.Fprint(w, ", ")
		g.emitExpr(w, n.Y)
		fmt.Fprint(w, ")")
		return
	}

	// Regular binary expression.
	g.emitExpr(w, n.X)
	fmt.Fprintf(w, " %s ", n.Op.String())
	g.emitExpr(w, n.Y)
}

// emitShiftOperand emits the left operand of a shift expression,
// with a cast where the operand needs one.
func (g *Generator) emitShiftOperand(w io.Writer, n *ast.BinaryExpr) {
	typ, ok := g.shiftCast(n)
	if !ok {
		g.emitExpr(w, n.X)
		return
	}
	fmt.Fprintf(w, "(%s)", typ)
	if _, ok := n.X.(*ast.BinaryExpr); ok {
		// A cast binds tighter than an arithmetic operator.
		g.emitParenExpr(w, n.X)
		return
	}
	g.emitExpr(w, n.X)
}

// emitEqual emits an == or != comparison. Strings, interfaces,
// slices, maps and arrays each compare in their own way.
func (g *Generator) emitEqual(w io.Writer, n *ast.BinaryExpr) {
	// Every rule below picks the comparison by the left operand's type, so
	// keep nil on the right: it says nothing about how to compare. Swapping
	// is safe, since a nil literal has no side effects.
	left, right := n.X, n.Y
	if isNilType(g.types.TypeOf(left)) {
		left, right = right, left
	}
	op := n.Op.String()

	leftType := g.types.TypeOf(left)
	rightType := g.types.TypeOf(right)
	rightIsNil := isNilType(rightType)

	if !rightIsNil && !comparableInC(leftType, rightType) {
		bad := right
		if !isInterfaceType(leftType) {
			// Report the error at the concrete operand, which is the one to fix.
			bad = left
		}
		g.fail(bad, "cannot compare %s with %s", g.typeString(leftType), g.typeString(rightType))
	}

	// String comparison: emit so_string_eq/ne calls.
	if g.hasStringType(left) {
		fmt.Fprintf(w, "%s(", stringCompareFunc(n.Op))
		g.emitExpr(w, left)
		fmt.Fprint(w, ", ")
		g.emitExpr(w, right)
		fmt.Fprint(w, ")")
		return
	}

	// Interface comparison: emit iface.self == NULL for nil,
	// or iface.self == other.self for non-nil.
	if isNamedNonEmptyInterface(leftType) {
		g.emitExpr(w, left)
		fmt.Fprintf(w, ".self %s ", op)
		if rightIsNil {
			fmt.Fprint(w, "NULL")
			return
		}
		g.emitExpr(w, right)
		fmt.Fprint(w, ".self")
		return
	}

	// Slice nil comparison: emit s.cap == 0 / != 0.
	if _, ok := leftType.Underlying().(*types.Slice); ok && rightIsNil {
		g.emitExpr(w, left)
		fmt.Fprintf(w, ".cap %s 0", op)
		return
	}

	// Map nil comparison: emit m == NULL / != NULL.
	if _, ok := leftType.Underlying().(*types.Map); ok && rightIsNil {
		g.emitExpr(w, left)
		fmt.Fprintf(w, " %s NULL", op)
		return
	}

	// Struct comparison.
	if _, ok := leftType.Underlying().(*types.Struct); ok {
		g.fail(left, "struct comparison is not supported")
		return
	}

	// Array comparison: emit so_mem_eq/ne calls.
	if arr, ok := leftType.Underlying().(*types.Array); ok {
		if n.Op == token.EQL {
			fmt.Fprint(w, "so_mem_eq(")
		} else {
			fmt.Fprint(w, "so_mem_ne(")
		}
		g.emitArrayCmpOperand(w, left, arr)
		fmt.Fprint(w, ", ")
		g.emitArrayCmpOperand(w, right, arr)
		elemType := g.mapTypeName(left, arr.Elem())
		fmt.Fprintf(w, ", %d * sizeof(%s))", arr.Len(), elemType)
		return
	}

	// Plain comparison.
	g.emitExpr(w, left)
	fmt.Fprintf(w, " %s ", op)
	g.emitExpr(w, right)
}

// emitCallExpr emits a function call or type conversion.
func (g *Generator) emitCallExpr(w io.Writer, n *ast.CallExpr) {
	g.checkMultiForward(n)

	// c.Val intrinsic: emit the string literal as a raw C expression.
	if raw, ok := g.cIntrinsic(n); ok {
		fmt.Fprint(w, raw)
		return
	}

	// unsafe.Offsetof intrinsic.
	if g.emitOffsetof(w, n) {
		return
	}

	// Generic function call with explicit type argument (e.g. fn[T](a) or pkg.Fn[T](a)).
	if indexExpr, ok := n.Fun.(*ast.IndexExpr); ok {
		if ident := exprIdent(indexExpr.X); ident != nil {
			if inst, ok := g.types.Instances[ident]; ok && inst.TypeArgs.Len() > 0 {
				g.emitGenericCall(w, n, indexExpr.X, inst)
				return
			}
		}
	}

	// Generic function call with multiple explicit type arguments (e.g. fn[K, V](a)).
	if indexListExpr, ok := n.Fun.(*ast.IndexListExpr); ok {
		if ident := exprIdent(indexListExpr.X); ident != nil {
			if inst, ok := g.types.Instances[ident]; ok && inst.TypeArgs.Len() > 0 {
				g.emitGenericCall(w, n, indexListExpr.X, inst)
				return
			}
		}
	}

	// Generic function call with inferred type argument (e.g. fn(a) where fn is generic).
	if ident := exprIdent(n.Fun); ident != nil {
		if inst, ok := g.types.Instances[ident]; ok && inst.TypeArgs.Len() > 0 {
			g.emitGenericCall(w, n, n.Fun, inst)
			return
		}
	}

	if tv, ok := g.types.Types[n.Fun]; ok && tv.IsType() {
		// Convert nil to the target type (e.g. ([]byte)(nil)).
		if isNilType(g.types.TypeOf(n.Args[0])) {
			g.emitExprAsType(w, n, n.Args[0], tv.Type)
			return
		}
		// Convert value to an interface type (e.g. Shape(r)).
		if isInterfaceType(tv.Type) {
			iface := tv.Type.Underlying().(*types.Interface)
			if iface.Empty() {
				g.emitAnyValue(w, n, n.Args[0])
				return
			}
			// Named non-empty interface conversion (e.g. Shape(r)).
			g.emitInterfaceLit(w, tv.Type, n.Args[0])
			return
		}
		// String-to-slice conversion ([]byte(s) or []rune(s)).
		if sl, ok := tv.Type.Underlying().(*types.Slice); ok {
			if g.hasStringType(n.Args[0]) {
				g.emitSliceCast(w, n, sl)
				return
			}
		}
		// Slice/integer-to-string conversion.
		if basic, ok := tv.Type.Underlying().(*types.Basic); ok && basic.Kind() == types.String {
			// A constant conversion becomes a string literal.
			if val := g.types.Types[n].Value; val != nil && val.Kind() == constant.String {
				fmt.Fprintf(w, "so_str(%s)", cStringVal(constant.StringVal(val)))
				return
			}
			argType := g.types.TypeOf(n.Args[0])
			if sl, ok := argType.Underlying().(*types.Slice); ok {
				g.emitStringCast(w, n, sl)
				return
			}
			if isIntegerType(argType) {
				fmt.Fprint(w, "so_int_string(")
				g.emitExpr(w, n.Args[0])
				fmt.Fprint(w, ")")
				return
			}
		}
		// Slice-to-array conversion (e.g. [3]int(s)). Unlike in Go, it doesn't
		// allocate a new array. Instead, it returns a pointer to the slice data.
		if arrType, ok := tv.Type.Underlying().(*types.Array); ok {
			argType := g.types.TypeOf(n.Args[0])
			if _, ok := argType.Underlying().(*types.Slice); ok {
				fmt.Fprint(w, "so_slice_array(")
				g.emitMacroArg(w, n.Args[0])
				fmt.Fprintf(w, ", %d)", arrType.Len())
				return
			}
		}
		// Slice-to-array-pointer conversion (e.g. (*[3]int)(s)).
		if ptrType, ok := tv.Type.Underlying().(*types.Pointer); ok {
			arrType, isArray := ptrType.Elem().Underlying().(*types.Array)
			_, isSlice := g.types.TypeOf(n.Args[0]).Underlying().(*types.Slice)
			if isArray && isSlice {
				fmt.Fprintf(w, "(%s)so_slice_array(", g.mapTypeName(n, tv.Type))
				g.emitMacroArg(w, n.Args[0])
				fmt.Fprintf(w, ", %d)", arrType.Len())
				return
			}
		}
		// Struct, slice and array conversions.
		g.checkStructConv(n, tv.Type, n.Args[0])
		g.checkArrayConv(n, tv.Type, n.Args[0])
		// The cast isn't needed because the target and the argument have the
		// same C type, as guaranteed by checkStructConv.
		if isCStructType(tv.Type) {
			g.emitPostfixOperand(w, n.Args[0])
			return
		}
		// Regular type conversion (e.g. int(3.14)).
		cType := g.mapTypeName(n, tv.Type)
		fmt.Fprintf(w, "(%s)", cType)
		g.emitParenExpr(w, n.Args[0])
		return
	}

	// Method call (e.g. r.Area()).
	if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
		if selection, ok := g.types.Selections[sel]; ok && selection.Kind() == types.MethodVal {
			g.emitMethodCall(w, sel, n)
			return
		}
	}

	// Regular function call.
	g.emitFuncCall(w, n)
}

// emitGenericCall emits a generic function call as fn(T, a, b),
// where type arguments are prepended to the regular arguments.
func (g *Generator) emitGenericCall(w io.Writer, n *ast.CallExpr, fun ast.Expr, inst types.Instance) {
	if n.Ellipsis.IsValid() {
		if ident := exprIdent(fun); ident != nil {
			if ext, ok := g.getExtern(g.types.Uses[ident]); ok && !ext.nodecay {
				g.fail(n, "spreading variadic arguments to an extern function is not supported")
			}
		}
	}
	g.emitExpr(w, fun)
	fmt.Fprintf(w, "(%s", g.mapTypeName(n, inst.TypeArgs.At(0)))
	for i := 1; i < inst.TypeArgs.Len(); i++ {
		fmt.Fprintf(w, ", %s", g.mapTypeName(n, inst.TypeArgs.At(i)))
	}
	sig, _ := inst.Type.(*types.Signature)
	for i, arg := range n.Args {
		// Wrap args in parens to protect against
		// the preprocessor misinterpreting commas.
		fmt.Fprint(w, ", (")
		if sig != nil && i < sig.Params().Len() {
			g.emitExprAsType(w, n, arg, sig.Params().At(i).Type())
		} else {
			g.emitExpr(w, arg)
		}
		fmt.Fprint(w, ")")
	}
	fmt.Fprint(w, ")")
}

// emitSliceCast emits a string-to-slice conversion ([]byte(s) or []rune(s)).
func (g *Generator) emitSliceCast(w io.Writer, call *ast.CallExpr, sl *types.Slice) {
	switch elemKind(sl) {
	case types.Byte:
		fmt.Fprint(w, "so_string_bytes(")
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprint(w, ")")
	case types.Int32:
		fmt.Fprint(w, "so_string_runes(")
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprint(w, ")")
	default:
		g.fail(call, "unsupported string-to-slice conversion: %s", g.typeString(sl))
	}
}

// emitStringCast emits a slice-to-string conversion (string(bs) or string(rs)).
func (g *Generator) emitStringCast(w io.Writer, call *ast.CallExpr, sl *types.Slice) {
	switch elemKind(sl) {
	case types.Byte:
		fmt.Fprint(w, "so_bytes_string(")
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprint(w, ")")
	case types.Int32:
		fmt.Fprint(w, "so_runes_string(")
		g.emitMacroArg(w, call.Args[0])
		fmt.Fprint(w, ")")
	default:
		g.fail(call, "unsupported slice-to-string conversion: %s", g.typeString(sl))
	}
}

// emitCompositeLit emits a composite literal (struct or array initialization).
// Fields can be positional (Point{1, 2}) or named (Point{x: 1, x: 2}).
func (g *Generator) emitCompositeLit(w io.Writer, n *ast.CompositeLit) {
	if st, ok := n.Type.(*ast.StructType); ok {
		g.emitAnonStructLit(w, n, st)
		return
	}

	switch g.types.TypeOf(n).Underlying().(type) {
	case *types.Array:
		g.emitArrayLit(w, n)
		return
	case *types.Slice:
		g.emitSliceLit(w, n)
		return
	case *types.Map:
		g.emitMapLit(w, n)
		return
	}

	// Regular composite literal.
	g.emitStructLit(w, n)
}

// emitIdent emits an identifier.
func (g *Generator) emitIdent(w io.Writer, n *ast.Ident) {
	name := n.Name
	if name == "nil" {
		fmt.Fprint(w, "NULL")
		return
	}
	if obj := g.types.Uses[n]; obj != nil {
		if obj.Parent() == g.pkg.Types.Scope() {
			// Package-level declarations: exported names are prefixed
			// with the package name (e.g. RectArea -> geom_RectArea),
			// and extern overrides are applied (e.g. maxInt64 -> INT64_MAX).
			name = g.symbolName(obj)
		}
		if cast, ok := g.constCast(n, obj); ok {
			fmt.Fprintf(w, "(%s)", cast)
		}
	}
	if g.state.macroParams[name] {
		fmt.Fprintf(w, "%s_", name)
		return
	}
	fmt.Fprint(w, name)
}

// emitParenExpr emits a parenthesized expression.
func (g *Generator) emitParenExpr(w io.Writer, expr ast.Expr) {
	if isSelfParenthesized(expr) || g.isBraceInit(expr) {
		g.emitExpr(w, expr)
		return
	}
	fmt.Fprint(w, "(")
	g.emitExpr(w, expr)
	fmt.Fprint(w, ")")
}

// emitSelectorExpr emits a selector expression (e.g. geom.RectArea → geom_RectArea, or p.name).
func (g *Generator) emitSelectorExpr(w io.Writer, n *ast.SelectorExpr) {
	if ident, ok := n.X.(*ast.Ident); ok {
		if pkgName, ok := g.types.Uses[ident].(*types.PkgName); ok {
			// Use the extern C name if the symbol has one
			// (e.g. math.MaxInt64 → INT64_MAX).
			if info, ok := g.getExtern(g.types.Uses[n.Sel]); ok && info.name != "" {
				fmt.Fprint(w, info.name)
				return
			}
			// Imported symbols are prefixed with the
			// package name (e.g. fmt.Println → fmt_Println).
			g.checkPackage(n, g.types.Uses[n.Sel])
			if cast, ok := g.constCast(n, g.types.Uses[n.Sel]); ok {
				fmt.Fprintf(w, "(%s)", cast)
			}
			// The prefix is the original package name, not the import alias.
			fmt.Fprintf(w, "%s_%s", pkgName.Imported().Name(), n.Sel.Name)
			return
		}
	}

	// Method expression: T.method or (*T).method -> function name.
	if selection, ok := g.types.Selections[n]; ok && selection.Kind() == types.MethodExpr {
		// Get the named type (strip pointer if present).
		recv := selection.Recv()
		var named *types.Named
		if ptr, ok := recv.(*types.Pointer); ok {
			named = types.Unalias(ptr.Elem()).(*types.Named)
		} else {
			named = types.Unalias(recv).(*types.Named)
		}
		cName := g.mapTypeName(n, named) + "_" + n.Sel.Name

		// Pointer receiver methods use void* in C, but the function type expects T*.
		// Cast to match the function pointer type.
		declSig := selection.Obj().Type().(*types.Signature)
		if _, isPtrRecv := declSig.Recv().Type().(*types.Pointer); isPtrRecv {
			ct := g.mapTypeDecl(n, g.types.TypeOf(n))
			fmt.Fprintf(w, "(%s)%s", ct.Cast(), cName)
		} else {
			fmt.Fprint(w, cName)
		}
		return
	}

	// Method values are not supported, because they are essentially closures.
	if selection, ok := g.types.Selections[n]; ok && selection.Kind() == types.MethodVal {
		g.fail(n, "method values are not supported")
	}

	// Struct/interface field access.
	xType := g.types.TypeOf(n.X)
	_, isPtr := xType.Underlying().(*types.Pointer)
	g.emitPostfixOperand(w, n.X)
	if isPtr {
		fmt.Fprintf(w, "->%s", g.fieldNameOf(n.Sel))
	} else {
		fmt.Fprintf(w, ".%s", g.fieldNameOf(n.Sel))
	}
}

// emitStarExpr emits a dereference expression (e.g. *p).
func (g *Generator) emitStarExpr(w io.Writer, n *ast.StarExpr) {
	fmt.Fprint(w, "*")
	g.emitExpr(w, n.X)
}

// emitPostfixOperand emits an expression that a postfix operator (., ->, [], ++)
// will be appended to. C postfix operators bind tighter than a cast and than
// unary *, so both forms must be parenthesized, e.g. ((T*)p)->f or (*p)++.
func (g *Generator) emitPostfixOperand(w io.Writer, expr ast.Expr) {
	if !g.emitsCastOrUnary(expr) {
		g.emitExpr(w, expr)
		return
	}
	fmt.Fprint(w, "(")
	g.emitExpr(w, expr)
	fmt.Fprint(w, ")")
}

// emitIndexExpr emits an index expression.
// For arrays: a[i] directly. For slices/strings: so_at(T, s, i).
func (g *Generator) emitIndexExpr(w io.Writer, n *ast.IndexExpr) {
	// Maps use so_map_get.
	if _, ok := g.types.TypeOf(n.X).Underlying().(*types.Map); ok {
		g.emitMapIndexExpr(w, n)
		return
	}

	// Arrays use direct C indexing.
	if _, ok := g.types.TypeOf(n.X).Underlying().(*types.Array); ok {
		if _, ok := ast.Unparen(n.X).(*ast.CompositeLit); ok {
			g.fail(n, "cannot index an array literal; assign it to a variable first")
		}
		g.emitExpr(w, n.X)
		fmt.Fprint(w, "[")
		g.emitExpr(w, n.Index)
		fmt.Fprint(w, "]")
		return
	}

	// Pointer-to-array: p[i] becomes (*p)[i].
	if ptr, ok := g.types.TypeOf(n.X).Underlying().(*types.Pointer); ok {
		if _, ok := ptr.Elem().Underlying().(*types.Array); ok {
			fmt.Fprint(w, "(*")
			g.emitExpr(w, n.X)
			fmt.Fprint(w, ")[")
			g.emitExpr(w, n.Index)
			fmt.Fprint(w, "]")
			return
		}
	}

	// Slices and strings use so_at.
	var elemType string
	switch t := g.types.TypeOf(n.X).Underlying().(type) {
	case *types.Slice:
		elemType = g.mapTypeName(n, t.Elem())
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
	g.emitMacroArg(w, n.X)
	fmt.Fprint(w, ", ")
	g.emitExpr(w, n.Index)
	fmt.Fprint(w, ")")
}

// emitUnaryExpr emits a unary expression.
func (g *Generator) emitUnaryExpr(w io.Writer, n *ast.UnaryExpr) {
	if n.Op == token.AND {
		// &arrayParam: C array params decay to pointers, so &param
		// gives T** instead of T(*)[N]. Emit a cast instead.
		if ident, ok := n.X.(*ast.Ident); ok {
			if _, ok := g.types.TypeOf(n.X).Underlying().(*types.Array); ok {
				if g.isFuncParam(ident) {
					ct := g.mapTypeDecl(n, g.types.TypeOf(n))
					fmt.Fprintf(w, "(%s)", ct.Cast())
					g.emitExpr(w, n.X)
					return
				}
			}
		}
		if lit, ok := ast.Unparen(n.X).(*ast.CompositeLit); ok {
			// &Person{...} → &(Person){...}
			g.checkLitAddress(n, lit)
			fmt.Fprint(w, "&")
			g.emitExpr(w, lit)
			return
		}
	}
	// Arithmetic on a type narrower than int: wrap the result
	// at the width of the Go type.
	if typ, ok := g.narrowCast(n); ok {
		fmt.Fprintf(w, "(%s)(", typ)
		defer fmt.Fprint(w, ")")
	}
	if n.Op == token.XOR {
		fmt.Fprint(w, "~")
		g.emitExpr(w, n.X)
		return
	}
	fmt.Fprint(w, n.Op.String())
	g.emitExpr(w, n.X)
}

// emitExprAsType emits an expression as a specific type, handling special cases
// like interface conversions and nil assignments.
func (g *Generator) emitExprAsType(w io.Writer, node ast.Node, expr ast.Expr, targetType types.Type) {
	// Empty interface: emit as void*.
	if iface, ok := targetType.Underlying().(*types.Interface); ok && iface.Empty() {
		g.emitAnyValue(w, node, expr)
		return
	}
	// Named interface conversion: wrap concrete types as interface literals.
	if isNamedNonEmptyInterface(targetType) {
		valType := g.types.TypeOf(expr)
		if isNilType(valType) {
			cType := g.mapTypeName(node, targetType)
			fmt.Fprintf(w, "(%s){}", cType)
			return
		}
		if isConcreteNamedType(valType) {
			g.emitInterfaceLit(w, targetType, expr)
			return
		}
	}
	// Slice nil assignment: emit zero-initialized struct instead of NULL.
	if _, ok := targetType.Underlying().(*types.Slice); ok && isNilType(g.types.TypeOf(expr)) {
		fmt.Fprint(w, "(so_Slice){}")
		return
	}
	// Map nil assignment: emit NULL.
	if _, ok := targetType.Underlying().(*types.Map); ok && isNilType(g.types.TypeOf(expr)) {
		fmt.Fprint(w, "NULL")
		return
	}
	g.emitExpr(w, expr)
}

// emitMacroArg emits an argument to a function-like macro.
//
// Composite literals reference stack-backed temporaries, so passing them to a
// macro that indexes, slices, or stores them would create a use-after-scope
// bug. Only a struct value literal is safe (it gets copied into the slot); it
// is wrapped in parentheses so the preprocessor does not split it on the commas
// of its braced initializer. Everything else fails.
//
// All emitted macro calls must use emitMacroArg for their arguments.
func (g *Generator) emitMacroArg(w io.Writer, arg ast.Expr) {
	// &T{...}: a pointer to a block-scoped temporary that dangles once it escapes.
	if u, ok := arg.(*ast.UnaryExpr); ok && u.Op == token.AND {
		if _, ok := u.X.(*ast.CompositeLit); ok {
			g.fail(arg, "cannot use composite literal here; assign it to a variable first")
		}
	}
	// Handle composite literals either by rejecting them (arrays, slices, maps)
	// or emitting them with extra parens (structs).
	if lit, ok := arg.(*ast.CompositeLit); ok {
		// Only struct value literals are safe; they are copied into the slot.
		if _, ok := g.types.TypeOf(lit).Underlying().(*types.Struct); ok {
			fmt.Fprint(w, "(")
			g.emitExpr(w, lit)
			fmt.Fprint(w, ")")
			return
		}
		// Array, slice, or map literal: references stack-backed storage.
		g.fail(arg, "cannot use composite literal here; assign it to a variable first")
		return
	}
	// Not a composite literal, emit directly.
	g.emitExpr(w, arg)
}

// emitDiscard emits an expression statement that evaluates expr
// and throws the value away: (void)expr;
func (g *Generator) emitDiscard(w io.Writer, expr ast.Expr) {
	fmt.Fprintf(w, "%s(void)", g.indent())
	if g.needsVoidParens(expr) {
		fmt.Fprint(w, "(")
		g.emitExpr(w, expr)
		fmt.Fprint(w, ")")
	} else {
		g.emitExpr(w, expr)
	}
	fmt.Fprint(w, ";\n")
}

// checkLitAddress rejects the address of an array or map literal.
func (g *Generator) checkLitAddress(node ast.Node, lit *ast.CompositeLit) {
	switch g.types.TypeOf(lit).Underlying().(type) {
	case *types.Array:
		g.fail(node, "cannot take the address of an array literal; assign it to a variable first")
	case *types.Map:
		g.fail(node, "cannot take the address of a map literal; assign it to a variable first")
	}
}

// emitsCastOrUnary reports whether the emitted C expression starts with a cast
// or a unary operator. Go's grammar already requires parentheses around other
// unary forms, so only these two need special handling.
func (g *Generator) emitsCastOrUnary(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return true
	case *ast.CallExpr:
		// A conversion to a pointer type emits a cast. Every other conversion
		// emits a call, a compound literal, or the argument alone.
		tv, ok := g.types.Types[e.Fun]
		return ok && tv.IsType() && isPointerType(g.types.TypeOf(e))
	}
	return false
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
	if bin.Op == token.EQL || bin.Op == token.NEQ || isOrderCompare(bin.Op) {
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

// needsIntDivGuard reports whether a division or modulo of x by y needs the
// so_div or so_mod guard. C leaves two divisors undefined: 0 and -1 (which
// overflows for the minimum value). Floating-point division is well-defined
// (IEEE infinities/NaN), so only integers need the guard.
func (g *Generator) needsIntDivGuard(x, y ast.Expr) bool {
	if !isIntegerType(g.types.TypeOf(x)) {
		return false
	}
	if tv, ok := g.types.Types[y]; ok && tv.Value != nil {
		// Go's type checker rejects a constant 0 divisor,
		// but a constant -1 still needs the guard.
		n, ok := constant.Int64Val(constant.ToInt(tv.Value))
		return ok && n == -1
	}
	return true
}

// isFuncParam reports whether ident refers to
// a function parameter of the current function.
func (g *Generator) isFuncParam(ident *ast.Ident) bool {
	if g.state.fn.sig == nil {
		return false
	}
	obj := g.types.ObjectOf(ident)
	params := g.state.fn.sig.Params()
	for param := range params.Variables() {
		if param == obj {
			return true
		}
	}
	return false
}

// isSelfParenthesized reports whether expr emits its own parentheses.
func isSelfParenthesized(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	switch bin.Op {
	case token.SHL, token.SHR, token.AND, token.OR, token.XOR, token.AND_NOT:
		return true
	}
	return false
}

// isBraceInit reports whether expr emits a bare C brace initializer,
// which happens for array literals only.
func (g *Generator) isBraceInit(expr ast.Expr) bool {
	if _, ok := expr.(*ast.CompositeLit); !ok {
		return false
	}
	_, ok := g.types.TypeOf(expr).Underlying().(*types.Array)
	return ok
}

// isShift reports whether a token is a shift operator.
func isShift(op token.Token) bool {
	return op == token.SHL || op == token.SHR
}

// isOrderCompare reports whether a token is an ordering comparison
// operator, as opposed to == and !=.
func isOrderCompare(op token.Token) bool {
	switch op {
	case token.LSS, token.GTR, token.LEQ, token.GEQ:
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

// elemKind returns the basic kind of the slice element type. The element type
// can be a named type, as in "type Byte byte". A kind of types.Invalid marks
// an element type without a basic type, a struct for example.
func elemKind(sl *types.Slice) types.BasicKind {
	elem, ok := sl.Elem().Underlying().(*types.Basic)
	if !ok {
		return types.Invalid
	}
	return elem.Kind()
}
