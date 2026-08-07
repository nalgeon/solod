package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"io"
	"math"
)

// emitIntConst emits a computed constant integer expression as a single
// literal and reports whether it did. See [constFolder] for when that happens.
func (g *Generator) emitIntConst(w io.Writer, n ast.Expr) bool {
	switch n.(type) {
	case *ast.Ident, *ast.SelectorExpr, *ast.BasicLit:
		return false
	}
	if !g.folder().needsFold(n) {
		return false
	}
	// Go has already checked that the value fits the destination type.
	fmt.Fprint(w, intLit(g.types.Types[n].Value))
	return true
}

// cIntType returns the name of the C type an integer expression evaluates to.
func (g *Generator) cIntType(n ast.Expr) string {
	return g.mapTypeName(n, g.folder().cTypeOf(n))
}

// intLit formats an integer constant as a C literal.
func intLit(val constant.Value) string {
	if exceedsInt64(val) {
		return val.ExactString() + "u" // unsigned suffix for C
	}
	if num, ok := constant.Int64Val(val); ok && num == math.MinInt64 {
		// C has no negative literals. In C, -9223372036854775808 is a unary
		// minus on a value that no signed C type can hold.
		return "INT64_MIN"
	}
	return val.ExactString()
}

// emitsAsUint64 reports whether an untyped integer constant maps to uint64_t.
// Untyped int normally goes in int64_t, and a value above int64
// needs the unsigned type. A value above uint64 fits neither type, and
// constType rejects it.
func emitsAsUint64(typ types.Type, val constant.Value) bool {
	basic, ok := types.Unalias(typ).(*types.Basic)
	return ok && basic.Kind() == types.UntypedInt && exceedsInt64(val)
}

// isIntegerType reports whether t is an integer type (named or not).
func isIntegerType(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsInteger != 0
}

// isNegative reports whether an integer constant is below zero.
func isNegative(val constant.Value) bool {
	if val == nil || val.Kind() != constant.Int {
		return false
	}
	return constant.Sign(val) < 0
}

// exceedsInt64 reports whether an integer constant is too large for int64.
func exceedsInt64(val constant.Value) bool {
	if val == nil || val.Kind() != constant.Int {
		return false
	}
	_, ok := constant.Int64Val(val)
	return !ok
}

// exceedsUint64 reports whether an integer constant is too large for uint64.
func exceedsUint64(val constant.Value) bool {
	if val == nil || val.Kind() != constant.Int {
		return false
	}
	_, ok := constant.Uint64Val(val)
	return !ok
}
