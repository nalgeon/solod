package clang

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestReservedLabel(t *testing.T) {
	// A reserved-name label is rejected rather than renamed: go/types records
	// the label as a *types.Label in both Defs and Uses, and a label is
	// neither a local value nor a package-level object.
	src := `package x
func f() {
	long:
	for {
		break long
	}
	goto long
}`
	info, file := checkSnippet(t, src)

	var defObj, useObj types.Object
	ast.Inspect(file, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name != "long" {
			return true
		}
		if o := info.Defs[id]; o != nil {
			defObj = o
		}
		if o := info.Uses[id]; o != nil {
			useObj = o
		}
		return true
	})

	if defObj == nil {
		t.Fatal("label def not recorded in Defs; expected a *types.Label")
	}
	if useObj == nil {
		t.Fatal("label use not recorded in Uses; expected a *types.Label")
	}
	if _, ok := defObj.(*types.Label); !ok {
		t.Fatalf("label def object = %T, want *types.Label", defObj)
	}
	g := &Generator{pkg: &packages.Package{Types: defObj.Pkg(), Name: "x"}}
	if g.canRename(defObj) {
		t.Error("canRename(label) = true; a label is not renamable")
	}
}

func TestObjName(t *testing.T) {
	g, pkg := testPkg()
	fnScope := types.NewScope(pkg.Scope(), token.NoPos, token.NoPos, "f")

	newVar := func(scope *types.Scope, name string) *types.Var {
		v := types.NewVar(token.NoPos, pkg, name, types.Typ[types.Int])
		scope.Insert(v)
		return v
	}
	newNamed := func(name string, under types.Type) *types.Named {
		tn := types.NewTypeName(token.NoPos, pkg, name, nil)
		pkg.Scope().Insert(tn)
		return types.NewNamed(tn, under, nil)
	}

	rect := newNamed("Rect", types.NewStruct(nil, nil))
	box := newNamed("box", types.NewStruct(nil, nil))
	rater := newNamed("rater", types.NewInterfaceType(nil, nil).Complete())

	tests := []struct {
		name       string
		obj        types.Object
		wantName   string
		wantRename bool
	}{
		{"exported", newVar(pkg.Scope(), "Long"), "x_Long", true},
		{"package-level", newVar(pkg.Scope(), "long"), "long", true},
		{"local", newVar(fnScope, "long"), "long", true},
		{"field", types.NewField(token.NoPos, pkg, "long", types.Typ[types.Int], false), "long", false},
		{"method", newMethod(pkg, rect, "long"), "x_Rect_long", false},
		{"pointer method", newMethod(pkg, types.NewPointer(box), "long"), "box_long", false},
		{"interface method", newMethod(pkg, rater, "long"), "long", false},
	}
	for _, tt := range tests {
		if got := g.baseObjName(tt.obj); got != tt.wantName {
			t.Errorf("baseObjName(%s) = %q, want %q", tt.name, got, tt.wantName)
		}
		if got := g.canRename(tt.obj); got != tt.wantRename {
			t.Errorf("canRename(%s) = %v, want %v", tt.name, got, tt.wantRename)
		}
	}
}

func TestMakesCName(t *testing.T) {
	g, pkg := testPkg()
	other := types.NewPackage("example.com/y", "y")

	// A package-level object of an imported package has a scope of its own.
	// Renaming its identifier would rename only the use.
	imported := types.NewVar(token.NoPos, other, "EOF", types.Typ[types.Int])
	other.Scope().Insert(imported)

	extern := types.NewVar(token.NoPos, pkg, "errno", types.Typ[types.Int])
	pkg.Scope().Insert(extern)
	g.externs = map[types.Object]externInfo{extern: {}}

	tests := []struct {
		name string
		obj  types.Object
		want bool
	}{
		{"predeclared", types.Universe.Lookup("true"), false},
		{"imported", imported, false},
		{"import alias", types.NewPkgName(token.NoPos, pkg, "index", other), false},
		{"extern", extern, false},
		{"package-level", types.NewVar(token.NoPos, pkg, "div", types.Typ[types.Int]), true},
	}
	for _, tt := range tests {
		if got := g.makesCName(tt.obj); got != tt.want {
			t.Errorf("makesCName(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsReserved(t *testing.T) {
	tests := []struct {
		name      string
		fileScope bool
		want      bool
	}{
		// A keyword or a macro conflicts in every scope.
		{"long", false, true},
		{"long", true, true},
		{"stderr", false, true},
		// A C library function conflicts only in file scope.
		{"div", false, false},
		{"div", true, true},
		{"total", false, false},
		{"total", true, false},
	}
	for _, tt := range tests {
		if got := isReserved(tt.name, tt.fileScope); got != tt.want {
			t.Errorf("isReserved(%q, %v) = %v, want %v", tt.name, tt.fileScope, got, tt.want)
		}
	}
}

// testPkg builds a generator for package x and the objects to name.
func testPkg() (*Generator, *types.Package) {
	pkg := types.NewPackage("example.com/x", "x")
	g := &Generator{pkg: &packages.Package{Types: pkg, Name: "x"}}
	return g, pkg
}

// newMethod returns a method named name with a receiver of type recv.
func newMethod(pkg *types.Package, recv types.Type, name string) *types.Func {
	recvVar := types.NewVar(token.NoPos, pkg, "r", recv)
	sig := types.NewSignatureType(recvVar, nil, nil, nil, nil, false)
	return types.NewFunc(token.NoPos, pkg, name, sig)
}
