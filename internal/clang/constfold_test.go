package clang

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
	"strings"
	"testing"
)

type foldCase struct {
	typ  string // the type of the expression, or empty to keep it untyped
	expr string
	want bool // whether the whole expression folds into one literal
}

func TestConstFolderNeedsFold(t *testing.T) {
	cases := []foldCase{
		// C computes these at the target's width.
		{"uint", "^uint(0)", false},
		{"uint64", "1<<63 + (-1 + 2)", false},
		{"int32", "1<<40 >> 20", false},
		{"int64", "1<<20 | 1<<10", false},
		{"int64", "1<<20 - 1", false},
		{"uint64", "maxUint64 / 3", false},
		{"int64", "8 * ptrSize", false},

		// The value is above int64, but C would compute it in int64.
		{"uint64", "1<<64 - 1", true},
		{"uint64", "1<<100 >> 37", true},

		// No C type can hold this intermediate. Only fitsCType reports the
		// problem, because the expression has no shift and no negative value.
		{"uint64", "maxUint64 * 3 / 3", true},

		// The operation combines an untyped value above int64 with a negative
		// one. In the literal form, the operand of the minus is the untyped
		// 9223372036854775808.
		{"int64", "-9223372036854775808", true},
		{"int64", "-(1 << 63)", true},
		{"uint64", "maxUint64 + (-1)", true},

		// The shift count reaches the width of the type it shifts. This
		// happens for a left shift and for a right shift.
		{"uint64", "0 << 64", true},
		{"int64", "1 >> 64", true},

		// The left operand of the shift is negative. This happens for a left
		// shift and for a right shift.
		{"int64", "-1 << 63", true},
		{"int", "-1 >> 30", true},
		{"int64", "-1 << 63 >> 63", true},
		{"int", "^0 << 8", true},

		// An untyped rune is 32 bits wide. A value outside int32 does not fit
		// the emitted C type.
		{"", "'a' + 1<<40", true},
		{"", "'a' - 1<<40", true},

		// C cannot compute this value, and the value is above uint64. C has no
		// literal for it, so the enclosing expression folds.
		{"", "1 << 64", false},

		// The generator never folds a float constant as an integer, even when
		// C cannot compute the shift inside the constant.
		{"float64", "1<<64 * 1.0", false},

		// C cannot compute the intermediate, but the whole expression still
		// does not fold. See TestConstFolderFoldsLowest.
		{"uint64", "(1<<64 - 1) >> (64 - 8*ptrSize)", false},
	}

	info, file := checkSnippet(t, foldSrc(cases))
	folder := constFolder{info: info}
	for i, c := range cases {
		expr := varValue(t, file, fmt.Sprintf("v%d", i))
		if info.Types[expr].Value == nil {
			t.Fatalf("%s: not a constant, check the snippet", c.expr)
		}
		typ := c.typ
		if typ == "" {
			typ = "untyped"
		}
		if got := folder.needsFold(expr); got != c.want {
			t.Errorf("needsFold(%s as %s) = %v, want %v", c.expr, typ, got, c.want)
		}
	}
}

func TestConstFolderFoldsLowest(t *testing.T) {
	// Checks that an expression folds at the lowest node that can hold the
	// value. The generator emits the width dependent part around that node as
	// C operators.
	const src = `package x

const ptrSize = 4 << (uint64(^uintptr(0)) >> 63)

var v uint64 = (1<<64 - 1) >> (64 - 8*ptrSize)
`
	info, file := checkSnippet(t, src)
	folder := constFolder{info: info}

	shift := varValue(t, file, "v").(*ast.BinaryExpr)
	if folder.needsFold(shift) {
		t.Error("the shift folded, so the value carries the host's word size")
	}
	// The generator emits the parens as parens, so the mask folds one node
	// further in.
	mask := shift.X.(*ast.ParenExpr).X
	if !folder.needsFold(mask) {
		t.Error("the mask did not fold, so C has to compute 1<<64")
	}
}

func TestConstFolderIgnoresFloats(t *testing.T) {
	// Checks that the float path keeps a node with a float type, even when the
	// value of the node is an integer. Only the outermost node of a float
	// expression has a float value. A node below the outermost one can hold an
	// integer, and an integer above int64 looks like a value C cannot compute.
	const src = `package x

var v float64 = 1<<63 * 1.0
`
	info, file := checkSnippet(t, src)
	folder := constFolder{info: info}

	product := varValue(t, file, "v").(*ast.BinaryExpr)
	if folder.needsFold(product) {
		t.Error("the product folded as an integer")
	}
	if folder.needsFold(product.X) {
		t.Error("the shift folded as an integer")
	}
}

func TestCIntWidth(t *testing.T) {
	// Checks the width of every emitted C type. The types int, uint and uintptr
	// are 64 bits, because 64 is the widest a target can make them. A target
	// width flag would change these three types and no other.
	cases := []struct {
		kind types.BasicKind
		want int64
	}{
		{types.Int8, 8}, {types.Uint8, 8},
		{types.Int16, 16}, {types.Uint16, 16},
		{types.Int32, 32}, {types.Uint32, 32},
		{types.Int64, 64}, {types.Uint64, 64},
		{types.Int, 64}, {types.Uint, 64}, {types.Uintptr, 64},
		{types.UntypedInt, 64}, {types.UntypedRune, 32},
	}
	for _, c := range cases {
		typ := types.Typ[c.kind]
		if got := cIntWidth(typ); got != c.want {
			t.Errorf("cIntWidth(%s) = %d, want %d", typ, got, c.want)
		}
	}
	// A type with no width of its own, such as a pointer, cannot narrow a value.
	if got := cIntWidth(types.NewPointer(types.Typ[types.Int8])); got != 64 {
		t.Errorf("cIntWidth(*int8) = %d, want 64", got)
	}
}

func TestConstFolderRuneConst(t *testing.T) {
	// Documents an untyped rune constant whose value is outside int32. The
	// folder wants to fold even a reference to the constant, because the value
	// does not fit the emitted C type. emitIntConst refuses to fold a leaf, so
	// the value never replaces the reference.
	//
	// The declaration itself is still wrong. constType gives the constant the
	// type so_rune, and the value overflows so_rune. The generator today widens
	// only an untyped int above int64.
	const src = `package x

const runeBig = 'a' + 1<<40

var v rune = runeBig - 1<<40
`
	info, file := checkSnippet(t, src)
	folder := constFolder{info: info}

	diff := varValue(t, file, "v").(*ast.BinaryExpr)
	ref := diff.X.(*ast.Ident)
	if !folder.needsFold(ref) {
		t.Error("the folder no longer folds a reference to an out-of-range rune constant")
	}
	if !folder.needsFold(diff.Y) {
		t.Error("a shift that reaches the width of a rune no longer folds")
	}
	// The result of the subtraction is inside int32, so C can compute it.
	if folder.needsFold(diff) {
		t.Error("the subtraction folded, though its value fits a rune")
	}
}

func TestConstFolderFitsCType(t *testing.T) {
	// Checks the width arithmetic on its own. go/types rejects most of the
	// out-of-range combinations below before the folder sees them, so this test
	// builds the values by hand. A narrow target word size produces the same
	// shape. Go accepts a value for uint or int on a 64-bit host, and the
	// generator emits the value as a 32-bit C type.
	cases := []struct {
		kind types.BasicKind
		val  constant.Value
		want bool
	}{
		{types.Uint8, constant.MakeInt64(255), true},
		{types.Uint8, constant.MakeInt64(256), false},
		{types.Uint16, constant.MakeInt64(65535), true},
		{types.Uint16, constant.MakeInt64(65536), false},
		{types.Uint32, constant.MakeInt64(1<<32 - 1), true},
		{types.Uint32, constant.MakeInt64(1 << 32), false},
		{types.Int8, constant.MakeInt64(127), true},
		{types.Int8, constant.MakeInt64(128), false},
		{types.Int8, constant.MakeInt64(-128), true},
		{types.Int8, constant.MakeInt64(-129), false},
		{types.Int32, constant.MakeInt64(-1 << 31), true},
		{types.Int32, constant.MakeInt64(-1<<31 - 1), false},
		// A 64-bit type is as wide as a C literal, so it holds every value
		// fitsCLiteral accepts.
		{types.Uint64, constant.MakeUint64(1<<64 - 1), true},
		{types.Int64, constant.MakeInt64(-1 << 63), true},
		// An unsigned type cannot hold a negative value.
		{types.Uint32, constant.MakeInt64(-1), false},
		// There is nothing to check. The node has no value, or the type has no
		// width of its own.
		{types.Float64, constant.MakeInt64(1 << 40), true},
		{types.Int8, nil, true},
	}
	for _, c := range cases {
		expr := &ast.Ident{Name: "x"}
		info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{
			expr: {Type: types.Typ[c.kind], Value: c.val},
		}}
		folder := constFolder{info: info}
		if got := folder.fitsCType(expr); got != c.want {
			t.Errorf("fitsCType(%v as %s) = %v, want %v",
				c.val, types.Typ[c.kind], got, c.want)
		}
	}
}

func TestExprChildren(t *testing.T) {
	// Checks the node kinds a constant expression can contain. If a kind is
	// missing here, the fold decision never descends into that node.
	x, y := &ast.Ident{Name: "x"}, &ast.Ident{Name: "y"}
	cases := []struct {
		node ast.Expr
		want []ast.Expr
	}{
		{&ast.BinaryExpr{X: x, Y: y}, []ast.Expr{x, y}},
		{&ast.UnaryExpr{X: x}, []ast.Expr{x}},
		{&ast.ParenExpr{X: x}, []ast.Expr{x}},
		{&ast.CallExpr{Fun: y, Args: []ast.Expr{x}}, []ast.Expr{x}},
		{x, nil},
		{&ast.SelectorExpr{X: x, Sel: y}, nil},
		{&ast.BasicLit{Value: "1"}, nil},
	}
	for _, c := range cases {
		got := exprChildren(c.node)
		if !slices.Equal(got, c.want) {
			t.Errorf("exprChildren(%T) = %v, want %v", c.node, got, c.want)
		}
	}
}

// foldSrc builds a package that binds each expression to a name: v0, v1 and so
// on. A test then looks up a case by its index. A case with no type becomes an
// untyped constant. That is the only way to keep an untyped rune, or a value
// above uint64, out of a type that would reject it.
func foldSrc(cases []foldCase) string {
	var b strings.Builder
	b.WriteString("package x\n\n")
	b.WriteString("const maxUint64 = 1<<64 - 1\n")
	b.WriteString("const ptrSize = 4 << (uint64(^uintptr(0)) >> 63)\n\n")
	for i, c := range cases {
		if c.typ == "" {
			fmt.Fprintf(&b, "const v%d = %s\n", i, c.expr)
			continue
		}
		fmt.Fprintf(&b, "var v%d %s = %s\n", i, c.typ, c.expr)
	}
	return b.String()
}

// varValue returns the value of the package-level declaration
// with the given name.
func varValue(t *testing.T, f *ast.File, name string) ast.Expr {
	t.Helper()
	for _, d := range f.Decls {
		gen, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if ok && vs.Names[0].Name == name {
				return vs.Values[0]
			}
		}
	}
	t.Fatalf("no variable %s", name)
	return nil
}
