package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"
)

// emitStructTypeSpec emits a typedef struct for a struct type declaration.
func (g *Generator) emitStructTypeSpec(w io.Writer, spec *ast.TypeSpec) {
	st := spec.Type.(*ast.StructType)
	cName := g.symbolName(spec.Name.Name)
	fmt.Fprintf(w, "typedef struct %s {\n", cName)
	g.state.indent++
	for _, field := range st.Fields.List {
		typ := g.types.TypeOf(field.Type)
		for _, name := range field.Names {
			if sig, ok := typ.(*types.Signature); ok {
				g.emitFuncPtrField(w, spec, name.Name, sig, cName)
			} else {
				// Regular struct field.
				cType := g.mapType(field, typ)
				fmt.Fprintf(w, "%s%s %s;\n", g.indent(), cType, name.Name)
			}
		}
	}
	g.state.indent--
	fmt.Fprintf(w, "} %s;\n", cName)
}

// emitFuncPtrField emits a function pointer field in a struct typedef.
// Example: so_int (*ratingFn)(struct main_Movie m);
func (g *Generator) emitFuncPtrField(w io.Writer, node ast.Node, fieldName string, sig *types.Signature, enclosingStruct string) {
	retType := g.returnType(node, sig)
	var params []string
	for p := range sig.Params().Variables() {
		cType := g.mapType(node, p.Type())
		if cType == enclosingStruct {
			cType = "struct " + cType
		}
		params = append(params, cType+" "+p.Name())
	}
	fmt.Fprintf(w, "%s%s (*%s)(%s);\n", g.indent(), retType, fieldName, strings.Join(params, ", "))
}

// emitMethodDecl emits a method as a C function with void* self parameter.
func (g *Generator) emitMethodDecl(decl *ast.FuncDecl) {
	w := g.state.writer
	fn := newFuncDecl(g, decl)

	recv := decl.Recv.List[0]
	cStructType := g.symbolName(recvTypeName(recv))
	named := len(recv.Names) > 0 // does the receiver have a name?

	if named {
		// For value receivers, track the name so emitSelectorExpr uses ->
		// (all receivers become pointers in C via void* self cast).
		if _, ok := recv.Type.(*ast.Ident); ok {
			g.state.recvName = recv.Names[0].Name
			defer func() { g.state.recvName = "" }()
		}
	}

	g.rejectNamedReturns(decl, fn.sig)
	g.state.funcSig = fn.sig
	g.state.tempCount = 0
	if !emitDocComment(w, decl.Doc) {
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "%s%s %s(%s) {\n", fn.spec, fn.returnType(), fn.name(), fn.params())
	g.state.indent++

	if named {
		recvName := recv.Names[0].Name
		fmt.Fprintf(w, "%s%s* %s = (%s*)self;\n", g.indent(), cStructType, recvName, cStructType)
	} else {
		fmt.Fprintf(w, "%s(void)self;\n", g.indent())
	}

	for _, stmt := range decl.Body.List {
		ast.Walk(g, stmt)
	}
	g.state.indent--
	fmt.Fprintf(w, "}\n")
	g.state.funcSig = nil
}

// emitAnonStructLit emits an anonymous struct literal.
// (e.g. struct{ x, y int }{1, 2} or struct{ x, y int }{ x: 1, y: 2 })
func (g *Generator) emitAnonStructLit(n *ast.CompositeLit, st *ast.StructType) {
	w := g.state.writer
	// Struct fields declaration.
	fmt.Fprintf(w, "(struct {\n")
	for _, field := range st.Fields.List {
		typ := g.types.TypeOf(field.Type)
		cType := g.mapType(field, typ)
		for _, name := range field.Names {
			fmt.Fprintf(w, "%s    %s %s;\n", g.indent(), cType, name.Name)
		}
	}
	fmt.Fprintf(w, "%s})", g.indent())

	// Struct fields initialization.
	fmt.Fprintf(w, "{\n")
	fields := collectFieldNames(st)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprintf(w, ",\n")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fmt.Fprintf(w, "%s    .%s = ", g.indent(), kv.Key.(*ast.Ident).Name)
			g.emitExpr(kv.Value)
		} else {
			fmt.Fprintf(w, "%s    .%s = ", g.indent(), fields[i])
			g.emitExpr(elt)
		}
	}
	fmt.Fprintf(w, ",\n")
	fmt.Fprintf(w, "%s}", g.indent())
}

// emitStructLit emits a struct literal (e.g. Point{1, 2} or Point{x: 1, y: 2}).
func (g *Generator) emitStructLit(n *ast.CompositeLit) {
	w := g.state.writer
	typ := g.types.TypeOf(n.Type)
	cType := g.mapType(n, typ)
	fmt.Fprintf(w, "(%s){", cType)
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			fmt.Fprintf(w, ".%s = ", kv.Key.(*ast.Ident).Name)
			g.emitExpr(kv.Value)
		} else {
			g.emitExpr(elt)
		}
	}
	fmt.Fprintf(w, "}")
}

// emitMethodCall emits a method call.
func (g *Generator) emitMethodCall(sel *ast.SelectorExpr, args []ast.Expr) {
	w := g.state.writer
	selection := g.types.Selections[sel]
	recv := selection.Recv()

	// Get the struct type name.
	var named *types.Named
	if ptr, ok := recv.(*types.Pointer); ok {
		named = ptr.Elem().(*types.Named)
	} else {
		named = recv.(*types.Named)
	}

	// Interface method dispatch: s.Perim(2) → s.Perim(s.self, 2)
	if isInterfaceType(named) {
		g.emitExpr(sel.X)
		fmt.Fprintf(w, ".%s(", sel.Sel.Name)
		g.emitExpr(sel.X)
		fmt.Fprintf(w, ".self")
		for _, arg := range args {
			fmt.Fprintf(w, ", ")
			g.emitExpr(arg)
		}
		fmt.Fprintf(w, ")")
		return
	}

	// Regular method call: r.Area() → main_Rect_Area(&r)
	cStructType := g.symbolName(named.Obj().Name())
	cName := cStructType + "_" + sel.Sel.Name
	fmt.Fprintf(w, "%s(", cName)

	// Pass receiver: add & if it's a value, pass directly if already a pointer.
	xType := g.types.TypeOf(sel.X)
	if _, ok := xType.Underlying().(*types.Pointer); ok {
		g.emitExpr(sel.X)
	} else {
		fmt.Fprintf(w, "&")
		g.emitExpr(sel.X)
	}

	// Pass method arguments.
	for _, arg := range args {
		fmt.Fprintf(w, ", ")
		g.emitExpr(arg)
	}
	fmt.Fprintf(w, ")")
}

// collectFieldNames returns the field names from a struct type in order.
func collectFieldNames(st *ast.StructType) []string {
	var names []string
	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}
