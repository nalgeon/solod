package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"math"
	"strconv"
	"strings"
)

// emitFloatConst emits a computed constant float expression as a single
// literal and reports whether it did.
//
// Go computes a constant expression in arbitrary precision and rounds once, at
// the end. C rounds every intermediate result to the type it evaluates in, so
// an operator tree can produce a different value. 1/Ln10 differs by one ulp,
// 0.1+0.2 differs by more, and 1e200*1e200/1e200 becomes infinity. The
// generator emits the value Go computed, so C and Go agree.
func (g *Generator) emitFloatConst(w io.Writer, n ast.Expr) bool {
	switch n.(type) {
	case *ast.Ident, *ast.SelectorExpr, *ast.BasicLit:
		return false
	}
	tv := g.types.Types[n]
	if tv.Value == nil || !isFloatType(tv.Type) {
		return false
	}
	fmt.Fprint(w, g.floatLit(n, tv.Value, tv.Type))
	return true
}

// emitFloatLit emits a literal of a float type. val is the literal text without
// the digit separators.
func (g *Generator) emitFloatLit(w io.Writer, n *ast.BasicLit, val string, tv types.TypeAndValue) {
	num := g.floatVal(n, tv.Value, tv.Type)
	if n.Kind == token.INT {
		// The literal is an integer in a float context. 1e22 written out in
		// full does not fit any C integer type, and 25 would make x / 25
		// integer division.
		fmt.Fprint(w, formatFloat(num, tv.Type))
		return
	}
	fmt.Fprint(w, val)
	if isFloat32(tv.Type) {
		// Without the f suffix, C reads the literal as a double and promotes
		// the other operand to match. Go says `x == 0.1` is true for a
		// float32 x, and C says it is false.
		fmt.Fprint(w, "f")
	}
}

// floatVal rounds a float constant to the emitted C type, and reports an error
// if the constant does not fit. Out of range means infinity.
// Infinity is ordinary IEEE behavior, so C never reports the problem.
func (g *Generator) floatVal(node ast.Node, val constant.Value, typ types.Type) float64 {
	// Round to the emitted type. The shortest form and the out of range limit
	// both depend on that type.
	if isFloat32(typ) {
		num, _ := constant.Float32Val(val)
		if math.IsInf(float64(num), 0) {
			g.fail(node, "constant %s overflows float32", val)
		}
		return float64(num)
	}
	num, _ := constant.Float64Val(val)
	if math.IsInf(num, 0) {
		g.fail(node, "constant %s overflows float64", val)
	}
	return num
}

// floatLit formats a constant as a C float literal of the given type.
func (g *Generator) floatLit(node ast.Node, val constant.Value, typ types.Type) string {
	return formatFloat(g.floatVal(node, val, typ), typ)
}

// formatFloat formats a float value as a C literal of the given type.
func formatFloat(num float64, typ types.Type) string {
	bits, suffix := 64, ""
	if isFloat32(typ) {
		bits, suffix = 32, "f"
	}
	// Use the shortest form that reads back as the same value. Go has already
	// rounded the constant to the destination type, so the shortest form
	// reproduces the Go value.
	lit := strconv.FormatFloat(num, 'g', -1, bits)

	// Make C read the literal as floating point. A value like 25 formats as
	// "25". C reads "25" as an int and turns x / 25 into integer division.
	if !strings.ContainsAny(lit, ".eE") {
		lit += ".0"
	}
	return lit + suffix
}

// isFloatType reports whether typ is a float type.
func isFloatType(typ types.Type) bool {
	basic, ok := typ.Underlying().(*types.Basic)
	return ok && basic.Info()&types.IsFloat != 0
}

// isFloat32 reports whether typ maps to a C float, not to a double.
func isFloat32(typ types.Type) bool {
	basic, ok := typ.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.Float32
}
