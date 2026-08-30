package clang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

// checkSnippet type-checks a self-contained package and returns its type info
// and the parsed file. The snippet must use only builtin types (no imports).
func checkSnippet(t *testing.T, src string) (*types.Info, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	conf := types.Config{Error: func(error) {}} // tolerate unused-var noise
	conf.Check("x", fset, []*ast.File{f}, info)
	return info, f
}

func funcByName(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
			return fd
		}
	}
	return nil
}

func TestFindReturnEscapes(t *testing.T) {
	// Each function is named after whether it should be flagged.
	const src = `package x

type BoxStr struct{ s string }
type BoxCount struct {
	s string
	n int
}
type BoxArrStr struct{ a [2]string }
type BoxArrInt struct{ a [3]int }
type Arr [3]int

// --- should be flagged (returns frame-bound memory) ---

func escConcat(a, b string) string        { return a + b }
func escConcatLit(a string) string        { return "x" + a }
func escConcatVar(a, b string) string     { s := a + b; return s }
func escConcatChain(a, b string) string   { t := a + b; s := t; return s }
func escAddAssign(s, a string) string     { s += a; return s }
func escRunes(s string) []rune            { return []rune(s) }
func escRunesToStr(rs []rune) string      { return string(rs) }
func escByteToStr(b byte) string          { return string(b) }
func escRuneToStr(r rune) string          { return string(r) }
func escIntToStr(i int) string            { return string(i) }
func escArray() [2][3]int                 { return [2][3]int{} }
func escArrayPtr() *[3]int                { a := [3]int{}; return &a }
func escLocalArray() [3]int               { var a [3]int; return a }
func escLocalArrayDef() [3]int            { a := [3]int{}; return a }
func escNamedArray() Arr                  { var a Arr; return a }
func escMakeSlice(n int) []int            { return make([]int, n) }
func escMakeMap() map[string]int          { return make(map[string]int, 4) }
func escAddrLocal() *int                  { x := 5; return &x }
func escAddrParam(p int) *int             { return &p }
func escAddrComposite() *BoxStr           { return &BoxStr{} }
func escStruct(a, b string) BoxStr        { return BoxStr{s: a + b} }
func escNew() *BoxStr                     { return new(BoxStr) }
func escNewLit(a, b string) *BoxStr       { return new(BoxStr{s: a + b}) }
func escMapLit() map[string]int           { return map[string]int{"a": 1} }
func escMapLitEmpty() map[string]int      { return map[string]int{} }
func escSliceLit(a, b string) []string    { return []string{a + b} }
func escSliceLitConst() []int             { return []int{1, 2, 3} }
func escSliceLitVar() []int               { s := []int{1, 2, 3}; return s }
func escSubslice(a, b string) string      { s := a + b; return s[1:] }
func escAppend() []int                    { s := make([]int, 0, 4); return append(s, 1) }
func escAppendFrameDst() []byte           { var a [4]byte; s := a[:0]; return append(s, 1) }
func escAppendElem(dst []string, a, b string) []string { return append(dst, a+b) }
func escAppendSpreadLit(dst []string, a, b string) []string {
	return append(dst, []string{a + b}...)
}
func escAppendSpreadVar(dst []string, a, b string) []string {
	s := []string{a + b}
	return append(dst, s...)
}
func escAliasField(a, b string) BoxStr {
	var p BoxStr
	q := &p
	q.s = a + b
	return p
}
func escAliasCopy(a, b string) BoxStr {
	var p BoxStr
	q := &p
	r := q
	r.s = a + b
	return p
}
func escFieldAssign(a, b string) BoxStr {
	var p BoxStr
	p.s = a + b
	return p
}
func escFieldAssignPtr(a, b string) *BoxStr {
	p := new(BoxStr)
	p.s = a + b
	return p
}
func escIndexAssign(s []string, a, b string) []string {
	s[0] = a + b
	return s
}
func escMapAssign(m map[string]string, a, b string) map[string]string {
	m["k"] = a + b
	return m
}
func escStarAssign(p *string, a, b string) *string {
	*p = a + b
	return p
}
func escArrayLitElem(x, y string) BoxArrStr {
	return BoxArrStr{a: [2]string{x + y}}
}
func escArrayVarElem(x, y string) BoxArrStr {
	var a [2]string
	a[0] = x + y
	return BoxArrStr{a: a}
}
func escFieldRead(a, b string) string {
	var p BoxStr
	p.s = a + b
	return p.s
}
func escIndexRead(a, b string) string {
	s := []string{a + b}
	return s[0]
}
func escStarRead(a, b string) BoxStr {
	var p BoxStr
	q := &p
	q.s = a + b
	return *q
}

// --- should NOT be flagged (safe) ---

func okConstConcat() string               { return "a" + "b" }
func okConstToStr() string                { return string(65) }
func okConstConvConcat() string           { return string(65) + "b" }
func okReturnParam(s string) string       { return s }
func okReturnPtrParam(p *int) *int        { return p }
func okParamSubslice(s string) string     { return s[1:] }
func okBytesFromParam(s string) []byte    { return []byte(s) }
func okStrFromByteParam(bs []byte) string { return string(bs) }
func okLenOfConcat(a, b string) int       { return len(a + b) }
func okReassignSafe(a, b, c string) string { s := a + b; s = c; _ = s; return c }
func okArrayParam(a [3]int) [3]int        { return a }
func okNamedArrayParam(a Arr) Arr         { return a }
func okMapParam(m map[string]int) map[string]int { return m }
func okSliceParam(s []int) []int          { return s }
func okEmptySliceLit() []int              { return []int{} }
func okStructParam(p BoxStr) BoxStr       { return p }
func okStructPtrParam(p *BoxStr) *BoxStr  { return p }
func okStructWithArray() BoxArrInt        { return BoxArrInt{a: [3]int{1, 2, 3}} }
func okStructArrLit() BoxArrStr           { return BoxArrStr{a: [2]string{"x", "y"}} }
func okAppendParam(dst []byte, a [3]byte) []byte { return append(dst, a[:]...) }
func okAppendElem(dst []string, s string) []string { return append(dst, s) }
func okAppendSpreadParam(dst, src []string) []string { return append(dst, src...) }
func okAppendFrameSpread(dst []byte) []byte {
	// The elements are plain bytes, so copying them in carries no frame memory.
	var a [4]byte
	return append(dst, a[:]...)
}
func okFieldAssignSafe(p BoxStr, s string) BoxStr {
	p.s = s
	return p
}
func okFieldAssignNoReturn(a, b string) string {
	var p BoxStr
	p.s = a + b
	return "x"
}
func okIndexAssignSafe(s []string, v string) []string {
	s[0] = v
	return s
}
func okIntFieldRead(a, b string) int {
	// The field is copied whole, so it carries no frame memory even though
	// the assignment above marks all of p.
	var p BoxCount
	p.s = a + b
	return p.n
}
func okByteIndexRead(a, b string) byte {
	var arr [2]byte
	return arr[0]
}
func okAliasSafe(s string) BoxStr {
	var p BoxStr
	q := &p
	q.s = s
	return p
}

// --- known misses (NOT safe, but out of scope; see the Scope comment in escape.go) ---
//
// These do dangle. They are named ok so the test passes: it pins what the
// checker gives up on, so a future change that starts catching them is visible.

func okMissDoublePtr(a, b string) BoxStr {
	// Only one pointer hop is tracked, so pp is marked but p is not.
	var p BoxStr
	pp := &p
	qq := &pp
	(*qq).s = a + b
	return p
}
func okMissOutParam(p *string, a, b string) {
	// Nothing is returned, and only returns are checked.
	*p = a + b
}
func okMissOpaqueCall(a, b string) string {
	// The result of a call is opaque, even when it hands back frame memory.
	return identity(a + b)
}
func identity(s string) string { return s }

// --- methods (checked like functions) ---

func (p BoxStr) escMethodConcat(a string) string { return p.s + a }
func (p *BoxStr) escMethodAddr() *BoxStr         { return &BoxStr{} }
func (p BoxStr) okMethodField() string           { return p.s }
func (p *BoxStr) okMethodRecv() *BoxStr          { return p }
`

	info, f := checkSnippet(t, src)

	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		wantEscape := len(name) >= 3 && name[:3] == "esc"
		got := newEscapeChecker(info, fd).escapes(fd)
		if wantEscape && len(got) == 0 {
			t.Errorf("%s: expected an escape, found none", name)
		}
		if !wantEscape && len(got) != 0 {
			t.Errorf("%s: expected no escape, found %d", name, len(got))
		}
	}
}

func TestFindGlobalStores(t *testing.T) {
	// Each function is named after whether it should be flagged.
	const src = `package x

type Box struct{ s string }

var gStr string
var gInt int
var gPtr *int
var gStrPtr *string
var gBox Box
var gStrs = []string{""}
var gInts []int
var gRunes []rune

func mk() string { return "ok" }

// --- should be flagged (stores frame-bound memory) ---

func escConcat(a, b string)     { gStr = a + b }
func escAddAssign(a string)     { gStr += a }
func escChain(a, b string)      { t := a + b; gStr = t }
func escMake()                  { gInts = make([]int, 3) }
func escRunes(s string)         { gRunes = []rune(s) }
func escAddr()                  { n := 1; gPtr = &n }
func escField(a, b string)      { gBox.s = a + b }
func escElem(a, b string)       { gStrs[0] = a + b }
func escDeref(a, b string)      { *gStrPtr = a + b }
func escLit()                   { gInts = []int{1, 2, 3} }

// --- should not be flagged ---

func okCall()                   { gStr = mk() }
func okLitConcat()              { gStr = "x" + "y" }
func okAddrGlobal()             { gPtr = &gInt }
func okInt(n int)               { gInt = n }
func okLocalConcat(a, b string) { s := a + b; _ = s }
func okLocalAddAssign(a string) { s := "x"; s += a; _ = s }
func okReadGlobal()             { s := gStr; _ = s }
`

	info, f := checkSnippet(t, src)

	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		wantStore := len(name) >= 3 && name[:3] == "esc"
		got := newEscapeChecker(info, fd).globalStores(fd.Body)
		if wantStore && len(got) == 0 {
			t.Errorf("%s: expected a global store, found none", name)
		}
		if !wantStore && len(got) != 0 {
			t.Errorf("%s: expected no global store, found %d", name, len(got))
		}
	}
}

func TestFindParamStores(t *testing.T) {
	// Each function is named after whether it should be flagged.
	const src = `package x

type Box struct{ s string }

func mk() string { return "ok" }

// --- should be flagged (stores frame-bound memory in the caller) ---

func escPtr(out *string, a, b string)         { *out = a + b }
func escPtrField(out *Box, a, b string)       { out.s = a + b }
func escPtrAddAssign(out *string, a string)   { *out += a }
func escSlice(out []string, a, b string)      { out[0] = a + b }
func escArray(out [2]string, a, b string)     { out[0] = a + b }
func escMap(m map[string]string, a, b string) { m["k"] = a + b }
func escMake(out *[]int, n int)               { *out = make([]int, n) }
func escRunes(out *[]rune, s string)          { *out = []rune(s) }
func escAddr(out **int)                       { n := 1; *out = &n }
func escChain(out *string, a, b string)       { t := a + b; *out = t }

func (b *Box) escMethod(a, c string) { b.s = a + c }

// --- should not be flagged ---

func okPtrInt(out *int, n int)            { *out = n * 2 }
func okPtrCopy(out *string, s string)     { *out = s }
func okPtrCall(out *string)               { *out = mk() }
func okPtrLit(out *string)                { *out = "a" + "b" }
func okByte(dst []byte, s string)         { dst[0] = s[0] }
func okAppend(dst *[]int, v int)          { *dst = append(*dst, v) }
func okValueParam(s string)               { s += "y"; _ = s }
func okValueField(b Box, a, c string)     { b.s = a + c }
func okLocalSlice(a, b string)            { buf := make([]string, 2); buf[0] = a + b }
func okLocalArray(a, b string)            { var arr [2]string; arr[0] = a + b }
func okLocalPtr(a, b string)              { var v Box; p := &v; p.s = a + b }

func (b Box) okValueMethod(a, c string) { b.s = a + c }
`

	info, f := checkSnippet(t, src)

	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		wantStore := len(name) >= 3 && name[:3] == "esc"
		got := newEscapeChecker(info, fd).paramStores(fd.Body)
		if wantStore && len(got) == 0 {
			t.Errorf("%s: expected a param store, found none", name)
		}
		if !wantStore && len(got) != 0 {
			t.Errorf("%s: expected no param store, found %d", name, len(got))
		}
	}
}
