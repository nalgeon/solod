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
// the same expression emitted as an operator tree can produce a different
// value: 1/Ln10 differs by one ulp, 0.1+0.2 by more, and 1e200*1e200/1e200
// becomes infinity. Emitting the computed Go value keeps the two in agreement.
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

// emitFloatLit emits a literal of a float type. val is the literal text with
// the digit separators removed.
func (g *Generator) emitFloatLit(w io.Writer, n *ast.BasicLit, val string, tv types.TypeAndValue) {
	num := g.floatVal(n, tv.Value, tv.Type)
	if n.Kind == token.INT {
		// An integer literal in a float context: 1e22 written out in full does
		// not fit any C integer type, and 25 would make x / 25 integer division.
		fmt.Fprint(w, formatFloat(num, tv.Type))
		return
	}
	fmt.Fprint(w, val)
	if isFloat32(tv.Type) {
		// Without the f suffix C reads the literal as a double and promotes the
		// other operand to match, so `x == 0.1` with a float32 x is false in C
		// where Go says it is true.
		fmt.Fprint(w, "f")
	}
}

// floatVal rounds a float constant to the C type it is emitted as, and reports
// an error if it does not fit. Out of range means infinity, which is ordinary
// IEEE behavior, so C never diagnoses it.
func (g *Generator) floatVal(node ast.Node, val constant.Value, typ types.Type) float64 {
	// Round to the type the literal is emitted as: both the shortest form and
	// what counts as out of range depend on it.
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
	// The shortest form that reads back as the same value. Go has already
	// rounded the constant to the destination type, so this reproduces it.
	lit := strconv.FormatFloat(num, 'g', -1, bits)

	// Make C read the literal as floating point. A value like 25 formats as
	// "25", which is an int in C, turning x / 25 into integer division.
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

// isFloat32 reports whether typ is emitted as a C float rather than a double.
func isFloat32(typ types.Type) bool {
	basic, ok := typ.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.Float32
}
