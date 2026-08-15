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

// constCast returns the C type to which an untyped integer constant should be
// converted at a particular use, and reports whether the use requires the
// conversion.
//
// An untyped constant gets its type from each expression that uses it, but its
// declaration has a single C type. Using it with a narrower type or the opposite
// signedness causes C to perform the arithmetic using the declared type.
// The constCast conversion gives C the Go type used by the expression instead.
func (g *Generator) constCast(n ast.Expr, obj types.Object) (string, bool) {
	c, ok := obj.(*types.Const)
	if !ok {
		return "", false
	}
	if basic, ok := types.Unalias(c.Type()).(*types.Basic); !ok || basic.Info()&types.IsUntyped == 0 {
		// The declaration is not an untyped integer constant.
		return "", false
	}
	use := g.types.Types[n].Type
	if use == nil || !isIntegerType(use) {
		// The use is not an integer expression.
		return "", false
	}
	if basic, ok := types.Unalias(use).(*types.Basic); ok && basic.Info()&types.IsUntyped != 0 {
		// The use is in another constant expression, which has no type yet.
		return "", false
	}
	decl := g.constType(n, c)
	if cIntWidth(use) >= cIntWidth(decl) && isUnsignedType(use) == isUnsignedType(decl) {
		// The use type is at least as wide as the declaration, and has the same signedness.
		return "", false
	}
	// The use type is narrower than the declaration, or has the opposite signedness.
	return g.mapTypeName(n, use), true
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

// isUnsignedType reports whether t is an unsigned integer type.
func isUnsignedType(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsUnsigned != 0
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
