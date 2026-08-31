package clang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// State holds the code generation state for the current scope.
// All of it is transient: it is set up when emission enters a function
// body and cleared when emission leaves it.
type State struct {
	// Current indentation depth (0 = not indented at file level).
	depth int
	// Function whose body is being emitted.
	fn funcScope
	// Deferred generic calls to emit before returns, panics, and function end.
	defers []string
	// Counter for unique temp variable names.
	tempCount int
	// Whether we are emitting inside a #define macro body.
	inMacro bool
	// Non-type macro parameter names. They are suffixed with _ and parenthesized
	// to avoid name collisions (b->val = val) and syntax errors (&b->val) in macro bodies.
	macroParams map[string]bool
}

// funcScope describes the function whose body is being emitted.
type funcScope struct {
	decl *ast.FuncDecl
	sig  *types.Signature
}

// enterFunc starts a function body scope. Scopes never nest: Solod has no
// function literals, and generic functions are always top-level.
func (s *State) enterFunc(decl *ast.FuncDecl, sig *types.Signature) {
	*s = State{fn: funcScope{decl: decl, sig: sig}}
}

// enterMacro starts the macro body scope for a generic function. The body is
// written to a buffer and then inserted into a #define, so it starts one level deep.
// params holds the macro's non-type parameter names.
func (s *State) enterMacro(decl *ast.FuncDecl, sig *types.Signature, params map[string]bool) {
	*s = State{fn: funcScope{decl: decl, sig: sig}, depth: 1, inMacro: true, macroParams: params}
}

// leaveFunc ends the current function or macro body scope.
func (s *State) leaveFunc() {
	*s = State{}
}

// atTopLevel reports whether emission is at package scope, outside any function body.
func (s *State) atTopLevel() bool {
	return s.fn.sig == nil
}

// indent returns the indentation for the current scope.
func (s *State) indent() string {
	return strings.Repeat("    ", s.depth)
}

// Prefixes for generated temporary names. The prefix describes what the
// temporary holds.
const (
	tempResult = "res" // result of a call
	tempSwitch = "sw"  // switch tag
	tempAssign = "asg" // value from the right side of a multiple assignment
)

// newTemp returns a fresh name for a temporary variable at node's position.
// It skips names that are visible there, so the temporary does not redeclare
// an existing identifier in the same C block.
func (g *Generator) newTemp(node ast.Node, prefix string) string {
	for {
		g.state.tempCount++
		name := fmt.Sprintf("_%s%d", prefix, g.state.tempCount)
		if !g.isVisible(node.Pos(), name) {
			return name
		}
	}
}

// isVisible reports whether name is declared in any scope enclosing pos.
func (g *Generator) isVisible(pos token.Pos, name string) bool {
	scope := g.pkg.Types.Scope().Innermost(pos)
	if scope == nil {
		scope = g.pkg.Types.Scope()
	}
	// The position within the scope does not matter: a name declared
	// after pos still shares a C block with the temporary.
	_, obj := scope.LookupParent(name, token.NoPos)
	return obj != nil
}
