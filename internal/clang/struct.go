package clang

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strconv"
	"strings"
)

// emitStructTypeSpec emits a typedef struct for a struct type declaration.
// dirs provides parsed so: directives for package-level declarations.
func (g *Generator) emitStructTypeSpec(w io.Writer, spec *ast.TypeSpec, dirs directives) {
	st := spec.Type.(*ast.StructType)
	cName := g.mapObjName(g.types.Defs[spec.Name])
	attr := dirs.attrString()
	if attr != "" {
		fmt.Fprintf(w, "%stypedef struct %s %s {\n", g.indent(), attr, cName)
	} else {
		fmt.Fprintf(w, "%stypedef struct %s {\n", g.indent(), cName)
	}
	g.state.depth++
	index := 0
	for _, field := range st.Fields.List {
		typ := g.types.TypeOf(field.Type)
		for _, name := range field.Names {
			fieldName := g.fieldNameOfAt(name, index)
			index++
			if innerSt, ok := field.Type.(*ast.StructType); ok {
				g.emitInlineStructField(w, innerSt, fieldName)
			} else if sig, ok := typ.(*types.Signature); ok {
				g.emitFuncPtrField(w, spec, fieldName, sig, cName)
			} else {
				// Regular struct field (arrays get dimension suffix).
				ct := g.mapTypeDecl(field, typ)
				fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(fieldName))
			}
		}
	}
	g.state.depth--
	fmt.Fprintf(w, "%s} %s;\n", g.indent(), cName)
}

// emitFuncPtrField emits a function pointer field in a struct typedef.
// Example: so_int (*ratingFn)(struct main_Movie);
//
// Parameter names are omitted, because they are not required in a C function
// pointer declaration. Emitting them could cause a conflict with a C keyword.
func (g *Generator) emitFuncPtrField(w io.Writer, node ast.Node, fieldName string, sig *types.Signature, enclosingStruct string) {
	retType := g.returnType(node, sig)
	var params []string
	for p := range sig.Params().Variables() {
		cType := g.mapParamType(node, p.Type())
		if cType == enclosingStruct || cType == enclosingStruct+"*" {
			cType = "struct " + cType
		}
		params = append(params, cType)
	}
	fmt.Fprintf(w, "%s%s (*%s)(%s);\n", g.indent(), retType, fieldName, funcParams(params))
}

// emitInlineStructField emits an anonymous struct field inline within a parent struct.
// Example: struct { so_int n; so_int i; } loop;
// Does not support function pointer fields within the inline struct.
func (g *Generator) emitInlineStructField(w io.Writer, st *ast.StructType, fieldName string) {
	fmt.Fprintf(w, "%sstruct {\n", g.indent())
	g.state.depth++
	index := 0
	for _, f := range st.Fields.List {
		typ := g.types.TypeOf(f.Type)
		ct := g.mapTypeDecl(f, typ)
		for _, name := range f.Names {
			fmt.Fprintf(w, "%s%s;\n", g.indent(), ct.Decl(g.fieldNameOfAt(name, index)))
			index++
		}
	}
	g.state.depth--
	fmt.Fprintf(w, "%s} %s;\n", g.indent(), fieldName)
}

// emitAnonStructLit emits an anonymous struct literal.
// (e.g. struct{ x, y int }{1, 2} or struct{ x, y int }{ x: 1, y: 2 })
func (g *Generator) emitAnonStructLit(w io.Writer, n *ast.CompositeLit, st *ast.StructType) {
	// Struct fields declaration.
	fmt.Fprint(w, "(struct {\n")
	index := 0
	for _, field := range st.Fields.List {
		typ := g.types.TypeOf(field.Type)
		cType := g.mapTypeName(field, typ)
		for _, name := range field.Names {
			fmt.Fprintf(w, "%s    %s %s;\n", g.indent(), cType, g.fieldNameOfAt(name, index))
			index++
		}
	}
	fmt.Fprintf(w, "%s})", g.indent())

	// Struct fields initialization.
	if len(n.Elts) == 0 {
		fmt.Fprint(w, "{}")
		return
	}
	fmt.Fprint(w, "{\n")
	struc := g.types.TypeOf(n).Underlying().(*types.Struct)
	for i, elt := range n.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			key := kv.Key.(*ast.Ident)
			fmt.Fprintf(w, "%s    .%s = ", g.indent(), g.fieldNameOf(key))
			g.emitFieldValue(w, n, kv.Value, structFieldType(struc, key.Name))
		} else {
			fmt.Fprintf(w, "%s    .%s = ", g.indent(), g.fieldNameAt(struc.Field(i), i))
			g.emitFieldValue(w, n, elt, struc.Field(i).Type())
		}
		fmt.Fprint(w, ",\n")
	}
	fmt.Fprintf(w, "%s}", g.indent())
}

// emitStructLit emits a struct literal (e.g. Point{1, 2} or Point{x: 1, y: 2}).
func (g *Generator) emitStructLit(w io.Writer, n *ast.CompositeLit) {
	var typ types.Type
	if n.Type != nil {
		typ = g.types.TypeOf(n.Type)
	} else {
		typ = g.types.TypeOf(n)
	}
	cType := g.mapTypeName(n, typ)
	fmt.Fprintf(w, "(%s)", cType)
	g.emitBareStructInit(w, n)
}

// emitBareStructInit emits a struct literal as a bare initializer
// (e.g. {.n = 200, .i = 10}) without a compound literal cast prefix.
func (g *Generator) emitBareStructInit(w io.Writer, n *ast.CompositeLit) {
	struc := g.types.TypeOf(n).Underlying().(*types.Struct)
	fmt.Fprint(w, "{")
	for i, elt := range n.Elts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			key := kv.Key.(*ast.Ident)
			fmt.Fprintf(w, ".%s = ", g.fieldNameOf(key))
			if lit, ok := isAnonStructLit(kv.Value); ok {
				g.emitBareStructInit(w, lit)
			} else {
				g.emitFieldValue(w, n, kv.Value, structFieldType(struc, key.Name))
			}
		} else {
			if lit, ok := isAnonStructLit(elt); ok {
				g.emitBareStructInit(w, lit)
			} else {
				g.emitFieldValue(w, n, elt, struc.Field(i).Type())
			}
		}
	}
	fmt.Fprint(w, "}")
}

// emitFieldValue emits a struct literal field value.
// Unlike a plain assignment, an array field can only take an array literal.
func (g *Generator) emitFieldValue(w io.Writer, n ast.Node, expr ast.Expr, fieldType types.Type) {
	g.checkArrayValue(expr, fieldType)
	g.emitExprAsType(w, n, expr, fieldType)
}

// checkAnonStructAliases rejects aliases for anonymous structs.
func (g *Generator) checkAnonStructAliases() {
	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, ok := spec.Type.(*ast.StructType); ok && spec.Assign.IsValid() {
				g.fail(spec, "alias for an anonymous struct is not supported; declare a named struct type instead")
			}
			return true
		})
	}
}

// checkFieldNames rejects structs with two fields of the same C name.
func (g *Generator) checkFieldNames() {
	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if st, ok := n.(*ast.StructType); ok {
				g.checkStructFields(st)
			}
			return true
		})
	}
}

// checkStructFields rejects a struct with two fields of the same C name.
func (g *Generator) checkStructFields(st *ast.StructType) {
	seen := map[string]bool{} // C names emitted for this struct
	index := 0
	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			cName := g.fieldNameOfAt(name, index)
			index++
			if seen[cName] {
				g.fail(name, "duplicate C field name %q in struct", cName)
			}
			seen[cName] = true
		}
	}
}

// checkEmbeddedFields rejects struct fields declared without a name.
func (g *Generator) checkEmbeddedFields(st *ast.StructType) {
	for _, field := range st.Fields.List {
		if len(field.Names) > 0 {
			continue
		}
		typ := g.typeString(g.types.TypeOf(field.Type))
		g.fail(field, "embedded field %s is not supported; declare a named field instead", typ)
	}
}

// checkStructConv rejects a conversion between two incompatible struct types.
// The types are compatible only if they have the same underlying type.
func (g *Generator) checkStructConv(n *ast.CallExpr, target types.Type, arg ast.Expr) {
	targetStruct := target.Underlying()
	if _, ok := targetStruct.(*types.Struct); !ok {
		return
	}
	argType := g.types.TypeOf(arg)
	if targetStruct == argType.Underlying() {
		return
	}
	g.fail(n, "cannot convert %s to %s; copy the fields instead",
		g.typeString(argType), g.typeString(target))
}

// anonStructType returns the inline C declaration of an anonymous struct type,
// for example `struct { so_int x; so_int y; }`.
func (g *Generator) anonStructType(node ast.Node, st *types.Struct) string {
	var sb strings.Builder
	sb.WriteString("struct {")
	for i := range st.NumFields() {
		field := st.Field(i)
		ct := g.mapTypeDecl(node, field.Type())
		sb.WriteString(" ")
		sb.WriteString(ct.Decl(g.fieldNameAt(field, i)))
		sb.WriteString(";")
	}
	sb.WriteString(" }")
	return sb.String()
}

// fieldNameOf resolves the C name for a field named by ident, whether ident is
// a field access selector, a field declaration, or a composite literal key.
// It falls back to the identifier text when ident does not denote a field, so
// output is unchanged for anonymous struct fields go/types may not record.
func (g *Generator) fieldNameOf(ident *ast.Ident) string {
	obj := g.types.Uses[ident]
	if obj == nil {
		obj = g.types.Defs[ident]
	}
	if field, ok := obj.(*types.Var); ok {
		return g.fieldName(field)
	}
	return ident.Name
}

// fieldName returns the C name emitted for a struct field: the override from a
// `c:"..."` tag if present, otherwise the field's Go name. Every field name
// reaches the generated C through this single function.
func (g *Generator) fieldName(field *types.Var) string {
	// The override map is keyed by the field's canonical Var. Origin normalize
	// an instantiated generic field back to its declaration, so accesses on a
	// Box[int] resolve the same override recorded on Box's declaration.
	if name, ok := g.fieldNames[field.Origin()]; ok {
		return name
	}
	return field.Name()
}

// fieldNameOfAt resolves the C name for the field declared at position index by
// ident. A blank field has no Go name, so it gets a generated C name.
func (g *Generator) fieldNameOfAt(ident *ast.Ident, index int) string {
	if ident.Name == "_" {
		return blankFieldName(index)
	}
	return g.fieldNameOf(ident)
}

// fieldNameAt returns the C name of the struct field at position index.
// A blank field has no Go name, so it gets a generated C name.
func (g *Generator) fieldNameAt(field *types.Var, index int) string {
	if field.Name() == "_" {
		return blankFieldName(index)
	}
	return g.fieldName(field)
}

// blankFieldName returns the C name of the blank ("_") field at position index.
// C needs a distinct name for every struct member.
func blankFieldName(index int) string {
	return "_" + strconv.Itoa(index)
}

// structFieldType returns the type of a struct field by name.
func structFieldType(st *types.Struct, name string) types.Type {
	for field := range st.Fields() {
		if field.Name() == name {
			return field.Type()
		}
	}
	panic("structFieldType: field not found: " + name)
}

// isAnonStruct reports whether typ is an anonymous struct type.
func isAnonStruct(typ types.Type) bool {
	// Named struct types are *types.Named, not *types.Struct.
	_, ok := types.Unalias(typ).(*types.Struct)
	return ok
}

// isAnonStructLit checks if an expression is an anonymous struct composite literal.
func isAnonStructLit(expr ast.Expr) (*ast.CompositeLit, bool) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}
	_, ok = lit.Type.(*ast.StructType)
	return lit, ok
}
