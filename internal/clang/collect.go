package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"
)

type symbolKind int

const (
	symbolFunc symbolKind = iota
	symbolMethod
	symbolType
)

type symbol struct {
	kind     symbolKind
	exported bool
	typeSpec *ast.TypeSpec
	funcDecl *ast.FuncDecl
}

// collectSymbols gathers all top-level type and function declarations
// into an ordered list. This list drives both header emission (exported
// symbols) and forward declarations in the .c file (unexported symbols).
func (g *Generator) collectSymbols() {
	for _, file := range g.pkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				if hasExternDirective(d.Doc) {
					continue
				}
				for _, spec := range d.Specs {
					ts := spec.(*ast.TypeSpec)
					if g.externs[ts.Name.Name] {
						continue
					}
					g.symbols = append(g.symbols, symbol{
						kind:     symbolType,
						exported: ast.IsExported(ts.Name.Name),
						typeSpec: ts,
					})
				}
			case *ast.FuncDecl:
				if d.Body == nil || d.Name.Name == "main" {
					continue
				}
				if g.externs[d.Name.Name] {
					continue
				}
				kind := symbolFunc
				exported := ast.IsExported(d.Name.Name)
				if d.Recv != nil {
					kind = symbolMethod
					if exported {
						exported = ast.IsExported(recvTypeName(d.Recv.List[0]))
					}
				}
				g.symbols = append(g.symbols, symbol{
					kind:     kind,
					exported: exported,
					funcDecl: d,
				})
			}
		}
	}
}

// collectExterns scans all files for extern symbols and #include directives.
// Body-less functions and declarations annotated with //so:extern are treated
// as external C symbols that should not be emitted.
func (g *Generator) collectExterns() {
	for _, file := range g.pkg.Syntax {
		// Collect // #include comments from the file.
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				text := strings.TrimPrefix(c.Text, "// ")
				if strings.HasPrefix(text, "#include") {
					g.includes = append(g.includes, text)
				}
			}
		}

		// Collect extern symbols from declarations.
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if !hasExternDirective(d.Doc) {
					continue
				}
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						g.externs[s.Name.Name] = true
					case *ast.ValueSpec:
						for _, name := range s.Names {
							g.externs[name.Name] = true
						}
					}
				}
			case *ast.FuncDecl:
				if d.Body == nil {
					g.externs[d.Name.Name] = true
				}
			}
		}
	}
}

// emitForwardDecls writes forward declarations for all unexported symbols.
// Types are emitted first, then functions/methods, so that type names
// are known before function prototypes reference them.
func (g *Generator) emitForwardDecls(w io.Writer) {
	// First pass: unexported types.
	for _, sym := range g.symbols {
		if sym.exported || sym.kind != symbolType {
			continue
		}
		g.emitForwardTypeDecl(w, sym.typeSpec)
	}
	// Second pass: unexported functions/methods.
	for _, sym := range g.symbols {
		if sym.exported || sym.kind == symbolType {
			continue
		}
		fn := newFuncDecl(g, sym.funcDecl)
		fmt.Fprintf(w, "%s%s %s(%s);\n", fn.spec, fn.returnType(), fn.name(), fn.params())
	}
}

// emitForwardTypeDecl writes a forward declaration for a type.
func (g *Generator) emitForwardTypeDecl(w io.Writer, spec *ast.TypeSpec) {
	cName := g.symbolName(spec.Name.Name)
	switch spec.Type.(type) {
	case *ast.StructType:
		fmt.Fprintf(w, "typedef struct %s %s;\n", cName, cName)
	case *ast.InterfaceType:
		iface := g.types.Defs[spec.Name].Type().Underlying().(*types.Interface)
		if iface.Empty() {
			cType := g.mapType(spec, iface)
			fmt.Fprintf(w, "typedef %s %s;\n", cType, cName)
		} else {
			fmt.Fprintf(w, "typedef struct %s %s;\n", cName, cName)
		}
	case *ast.FuncType:
		named := g.types.Defs[spec.Name].Type().(*types.Named)
		sig := named.Underlying().(*types.Signature)
		retType := g.returnType(spec, sig)
		var params []string
		for p := range sig.Params().Variables() {
			params = append(params, g.mapType(spec, p.Type()))
		}
		fmt.Fprintf(w, "typedef %s (*%s)(%s);\n", retType, cName, strings.Join(params, ", "))
	default:
		typ := g.types.Defs[spec.Name].Type()
		cType := g.mapType(spec, typ.Underlying())
		fmt.Fprintf(w, "typedef %s %s;\n", cType, cName)
	}
}

// hasExternDirective checks if a comment group contains the //so:extern directive.
func hasExternDirective(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.TrimSpace(c.Text) == "//so:extern" {
			return true
		}
	}
	return false
}
