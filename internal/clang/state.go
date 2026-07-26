package clang

import (
	"fmt"
	"go/types"
	"strings"
)

// State holds the code generation state for the current scope.
// All of it is transient: it is set up when emission enters a function
// body and cleared when emission leaves it.
type State struct {
	// Current indentation depth (0 = not indented at file level).
	depth int
	// Current function's signature (for multi-return).
	funcSig *types.Signature
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

// enterFunc starts a function body scope. Scopes never nest: So has no
// function literals, and generic functions are always top-level.
func (s *State) enterFunc(sig *types.Signature) {
	*s = State{funcSig: sig}
}

// enterMacro starts the macro body scope for a generic function. The body is
// written to a buffer and then inserted into a #define, so it starts one level deep.
// params holds the macro's non-type parameter names.
func (s *State) enterMacro(sig *types.Signature, params map[string]bool) {
	*s = State{funcSig: sig, depth: 1, inMacro: true, macroParams: params}
}

// leaveFunc ends the current function or macro body scope.
func (s *State) leaveFunc() {
	*s = State{}
}

// atTopLevel reports whether emission is at package scope, outside any function body.
func (s *State) atTopLevel() bool {
	return s.funcSig == nil
}

// indent returns the indentation for the current scope.
func (s *State) indent() string {
	return strings.Repeat("    ", s.depth)
}

// newTemp returns a fresh name for a temporary variable in the current function.
func (s *State) newTemp() string {
	s.tempCount++
	return fmt.Sprintf("_res%d", s.tempCount)
}
