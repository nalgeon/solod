package clang

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
)

// constFolder decides how the generator emits a constant integer expression.
// The two options are the C operators and the single value Go computed for the
// expression.
//
// The generator emits the operators by default. A constant may depend on the
// width of int, and the target decides that width. The generator emits the
// computed value only when C cannot do the arithmetic.
type constFolder struct {
	info *types.Info
}

// folder returns the constant folder for the emitted package.
func (g *Generator) folder() constFolder {
	return constFolder{info: g.types}
}

// needsFold reports whether the generator emits a constant integer expression
// as a single literal. If not, the generator emits C operators.
//
// An expression folds at the lowest node that can hold the value. If C cannot
// compute a node, that node gives the decision to its parent. The parent keeps
// the decision only while it can write the value as a C literal. A fold higher
// up computes the width dependent operands on the host, so the emitted value
// carries the host's word size and the target gets the wrong number.
func (f constFolder) needsFold(n ast.Expr) bool {
	return f.fitsCLiteral(n) && f.beyondC(n)
}

// fitsCLiteral reports whether C can write the value of an expression as an
// integer literal. C cannot write an intermediate above uint64. For such a
// value, the enclosing expression makes the fold decision.
func (f constFolder) fitsCLiteral(n ast.Expr) bool {
	val := f.info.Types[n].Value
	if val == nil || val.Kind() != constant.Int {
		return false
	}
	// A value above both limits has no C literal.
	return !exceedsInt64(val) || !exceedsUint64(val)
}

// beyondC reports whether C would fail to reproduce the value of an expression.
// The failure can happen at the node itself, or in a descendant that cannot
// fold.
//
// The result is true for any of the following:
//   - the value does not fit the emitted C type;
//   - the operation mixes an untyped value above int64 with a negative one;
//   - the shift count reaches the width of the type the shift evaluates in.
//     C does not define such a shift.
//   - the node is a float literal of an integer expression. C reads such a
//     literal as a double and computes the whole expression in double.
//
// Otherwise the result is false, and C can compute the value.
func (f constFolder) beyondC(n ast.Expr) bool {
	if !f.fitsCType(n) || f.mixesSignedness(n) || f.shiftExceedsWidth(n) || isFloatLit(n) {
		return true
	}
	for _, child := range exprChildren(n) {
		if f.beyondC(child) && !f.fitsCLiteral(child) {
			return true
		}
	}
	return false
}

// fitsCType reports whether the value of an expression fits the emitted C type.
// Go checks the final value of a constant expression against its type, but Go
// checks no untyped intermediate. In maxUint64 * 3 / 3 the product is above
// uint64. C wraps the product before the division brings it back into range.
func (f constFolder) fitsCType(n ast.Expr) bool {
	val := f.info.Types[n].Value
	if val == nil || val.Kind() != constant.Int {
		return true
	}
	typ := f.cTypeOf(n)
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok || basic.Info()&types.IsInteger == 0 {
		return true
	}
	width := uint(cIntWidth(typ))
	if basic.Info()&types.IsUnsigned != 0 {
		num, ok := constant.Uint64Val(val)
		return ok && (width == 64 || num>>width == 0)
	}
	num, ok := constant.Int64Val(val)
	if !ok || width == 64 {
		return ok
	}
	lim := int64(1) << (width - 1)
	return num >= -lim && num < lim
}

// mixesSignedness reports whether an operation combines an untyped value above
// int64 with a negative one. C evaluates such an operation in uint64 and turns
// the negative side into a large positive one. A sum or a product wraps back to
// the value Go computed. But a comparison against a signed operand does not
// compile, and the C compiler decides the result of a conversion to a signed type.
func (f constFolder) mixesSignedness(n ast.Expr) bool {
	var operands []ast.Expr
	switch expr := n.(type) {
	case *ast.BinaryExpr:
		operands = []ast.Expr{expr.X, expr.Y}
	case *ast.UnaryExpr:
		// The minus in -(1 << 63). The operand is unsigned,
		// and the result is negative.
		operands = []ast.Expr{expr.X}
	default:
		return false
	}
	big, negative := false, isNegative(f.info.Types[n].Value)
	for _, op := range operands {
		tv := f.info.Types[op]
		big = big || emitsAsUint64(tv.Type, tv.Value)
		negative = negative || isNegative(tv.Value)
	}
	return big && negative
}

// shiftExceedsWidth reports whether a constant shift count reaches
// the width of the C type the shift evaluates in.
func (f constFolder) shiftExceedsWidth(n ast.Expr) bool {
	bin, ok := n.(*ast.BinaryExpr)
	if !ok || (bin.Op != token.SHL && bin.Op != token.SHR) {
		return false
	}
	count := f.info.Types[bin.Y].Value
	if count == nil || count.Kind() != constant.Int {
		return false
	}
	bits, ok := constant.Int64Val(count)
	if !ok {
		return true // the count is above int64, so it exceeds every width
	}
	return bits >= cIntWidth(f.cTypeOf(bin))
}

// isFloatLit reports whether an expression is a float literal, for example 1e9.
func isFloatLit(n ast.Expr) bool {
	lit, ok := n.(*ast.BasicLit)
	return ok && lit.Kind == token.FLOAT
}

// cTypeOf returns the Go type that decides the C type of an integer expression.
// This is the type go/types recorded, with one exception. An untyped value above
// int64 goes in uint64. Untyped normally maps to int64, which cannot hold such
// a value.
func (f constFolder) cTypeOf(n ast.Expr) types.Type {
	tv := f.info.Types[n]
	if emitsAsUint64(tv.Type, tv.Value) {
		return types.Typ[types.Uint64]
	}
	return tv.Type
}

// cIntWidth returns the width in bits of the emitted C type. The types int,
// uint and uintptr are pointer sized, so the target decides their width. These
// three types are 64 bits at most.
func cIntWidth(typ types.Type) int64 {
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return 64
	}
	switch basic.Kind() {
	case types.Int8, types.Uint8:
		return 8
	case types.Int16, types.Uint16:
		return 16
	case types.Int32, types.Uint32, types.UntypedRune:
		return 32
	default:
		return 64
	}
}

// exprChildren returns the immediate sub-expressions of a constant expression.
// A constant expression can also contain an identifier, a qualified name, or a
// literal. Each of those is a leaf.
func exprChildren(n ast.Expr) []ast.Expr {
	switch expr := n.(type) {
	case *ast.BinaryExpr:
		return []ast.Expr{expr.X, expr.Y}
	case *ast.UnaryExpr:
		return []ast.Expr{expr.X}
	case *ast.ParenExpr:
		return []ast.Expr{expr.X}
	case *ast.CallExpr:
		return expr.Args
	}
	return nil
}
