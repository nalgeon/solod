package clang

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
)

// Escape analysis
//
// Solod has no garbage collector and no heap by default. Several builtins allocate
// their result in the current function's stack frame: string concatenation, rune
// and []rune conversions, make for slices and maps, array literals, and taking
// the address of a local, etc. Such a value is a "frame value": it is valid only
// until the function returns. Returning one hands the caller a dangling pointer.
//
// This checker rejects three ways a frame value can outlive its frame: a return,
// a store into a package-level variable, and a store through an out parameter.
// It runs per function.
//
// # Scope
//
// The checker catches the most common ways frame memory can leave a function,
// and only those. It doesn't guarantee memory safety: code that passes can
// still have dangling pointers. That's intentional. A fully accurate analysis
// would need interprocedural summaries and a real points-to analysis, which
// would be much more expensive than it's worth here.
//
// The rule for changing this file: only add to it when a pattern actually
// appears in real code, not just to cover every possible case. Every case
// handled here repeats information about lowering that's already in the emitter
// (for example, isFrameComposite knows what emitSliceLit does), and that
// duplication makes the checker expensive to maintain. It's not worth adding
// precision that no program actually needs.
//
// What it deliberately does not do:
//
//   - Interprocedural analysis. A call to a user or stdlib function is opaque:
//     its result is assumed not to be a frame value. The make, new and append
//     builtins and the builtin conversions are the known exceptions, handled
//     explicitly.
//   - A parameter copied into a local first. In p := out; *p = a + b the store
//     goes through p, which collectPointsTo does not track back to out.
//   - Pointer chains past one hop (see collectPointsTo).
//
// Marking is done per variable, not per field: storing to p.s marks all of p
// (see rootLocal). This overestimates, so a return might be flagged even when
// it's actually safe. It's better to have a false positive (which causes a
// compile error the author can fix) than a false negative, which leads to
// memory corruption in the generated C code.

// escapeMsg is the diagnostic reported for a return that escapes the frame.
const escapeMsg = "stack-allocated value escapes function frame"

// globalStoreMsg is the diagnostic reported for a store into a package-level variable.
const globalStoreMsg = "stack-allocated value escapes into a package-level variable"

// paramStoreMsg is the diagnostic reported for a store through a parameter.
const paramStoreMsg = "stack-allocated value escapes through a parameter"

// checkFrameValues fails the build on the first frame value that outlives its
// frame. It checks each of the package's functions which have a body.
func (g *Generator) checkFrameValues() {
	for _, file := range g.pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || g.hasExtern(g.types.Defs[fn.Name]) {
				continue
			}
			g.checkFuncFrame(fn)
		}
	}
}

// checkFuncFrame runs both frame-value checks over one function.
func (g *Generator) checkFuncFrame(fn *ast.FuncDecl) {
	c := newEscapeChecker(g.types, fn)
	if c == nil {
		return
	}
	for _, node := range c.globalStores(fn.Body) {
		g.fail(node, "%s", globalStoreMsg)
	}
	for _, node := range c.paramStores(fn.Body) {
		g.fail(node, "%s", paramStoreMsg)
	}
	// A generic function expands to a macro in the frame of the caller
	// (see emitMacroFuncDecl), so its result outlives the call.
	if isGenericFunc(fn) {
		return
	}
	for _, node := range c.escapes(fn) {
		g.fail(node, "%s", escapeMsg)
	}
}

// escapeChecker holds the per-function analysis state.
type escapeChecker struct {
	info   *types.Info
	locals map[types.Object]bool           // every object declared in the function
	params map[types.Object]bool           // the receiver and the parameters
	points map[types.Object][]types.Object // locals that a local pointer may point to
	fvars  map[types.Object]bool           // locals that hold a frame value
}

// newEscapeChecker runs the analysis passes over a function body and returns the
// result. It returns nil for a function with no body.
func newEscapeChecker(info *types.Info, decl *ast.FuncDecl) *escapeChecker {
	if decl.Body == nil {
		return nil
	}
	c := &escapeChecker{
		info:   info,
		locals: map[types.Object]bool{},
		params: map[types.Object]bool{},
		points: map[types.Object][]types.Object{},
		fvars:  map[types.Object]bool{},
	}
	c.collectLocals(decl)
	c.collectParams(decl)
	c.collectPointsTo(decl.Body)
	c.markFrameVars(decl.Body)
	return c
}

// collectLocals records every object declared inside the function: receiver,
// parameters, results and body-local variables.
func (c *escapeChecker) collectLocals(decl *ast.FuncDecl) {
	ast.Inspect(decl, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if obj := c.info.Defs[id]; obj != nil {
				c.locals[obj] = true
			}
		}
		return true
	})
}

// collectParams records the receiver and the parameters.
func (c *escapeChecker) collectParams(decl *ast.FuncDecl) {
	for _, list := range []*ast.FieldList{decl.Recv, decl.Type.Params} {
		if list == nil {
			continue
		}
		for _, field := range list.List {
			for _, name := range field.Names {
				if obj := c.info.Defs[name]; obj != nil {
					c.params[obj] = true
				}
			}
		}
	}
}

// collectPointsTo fills points with the locals that each local pointer may point
// to, so a store through the pointer also marks the target: in q := &p; q.s = a + b,
// it is p that ends up holding the frame value, not just q. Only the address
// of a local is tracked; a pointer from anywhere else points to memory this
// frame does not own.
//
// A pointer is tracked per root variable, so storing one into a field (q.next = &p)
// records p for all of q. That is imprecise but safe: it can only mark more locals
// than needed, never fewer.
//
// Only one hop is tracked. A store through a pointer to a pointer marks the middle
// pointer, not the final target: in pp := &p; qq := &pp; (*qq).s = a + b, pp is
// marked but p is not. Reading a pointer back out of a field or element (q := n.next)
// is not tracked either, only a copy of a whole pointer variable.
//
// It repeats to a fixpoint so it follows pointer copies (q := &p; r := q). The
// target sets only grow and are bounded by the number of locals, so it terminates.
func (c *escapeChecker) collectPointsTo(body *ast.BlockStmt) {
	for {
		changed := false
		c.walk(body, func(n ast.Node) {
			assignPairs(n, func(lhs, rhs ast.Expr) {
				changed = c.addPointsTo(lhs, rhs) || changed
			})
		})
		if !changed {
			return
		}
	}
}

// addPointsTo records that the local behind lhs may point to whatever rhs points to.
func (c *escapeChecker) addPointsTo(lhs, rhs ast.Expr) bool {
	obj := c.rootLocal(lhs)
	if obj == nil {
		return false
	}
	changed := false
	for _, target := range c.pointsTo(rhs) {
		if !slices.Contains(c.points[obj], target) {
			c.points[obj] = append(c.points[obj], target)
			changed = true
		}
	}
	return changed
}

// pointsTo returns the locals that the value of expr may point to.
func (c *escapeChecker) pointsTo(expr ast.Expr) []types.Object {
	switch x := expr.(type) {
	case *ast.ParenExpr:
		return c.pointsTo(x.X)
	case *ast.UnaryExpr:
		// &p, &p.s and &s[i] all point to the local they root at.
		if x.Op == token.AND {
			if obj := c.rootLocal(x.X); obj != nil {
				return []types.Object{obj}
			}
		}
	case *ast.Ident:
		// Copying a pointer copies where it points.
		if obj := c.info.ObjectOf(x); obj != nil {
			return c.points[obj]
		}
	}
	return nil
}

// markFrameVars fills fvars with the locals that hold a frame value.
// It repeats to a fixpoint so it follows assignment chains such as
//
//	t := a + b; s := t
//
// Marking only ever adds locals, so the loop terminates.
func (c *escapeChecker) markFrameVars(body *ast.BlockStmt) {
	for {
		changed := false
		c.walk(body, func(n ast.Node) {
			changed = c.markNode(n) || changed
		})
		if !changed {
			return
		}
	}
}

// markNode marks the locals that a single node turns into frame values.
func (c *escapeChecker) markNode(n ast.Node) bool {
	changed := false
	switch s := n.(type) {
	case *ast.Ident:
		changed = c.markArrayDecl(s)
	case *ast.AssignStmt:
		changed = c.markAddAssign(s)
	}
	assignPairs(n, func(lhs, rhs ast.Expr) {
		if c.isFrameValue(rhs) {
			changed = c.markVar(lhs) || changed
		}
	})
	return changed
}

// markArrayDecl marks a local array variable, which is itself a frame value: an
// array translates to a bare C array, so returning it by value returns a pointer
// into the frame. It acts on the defining ident, which lives in the body, so it
// skips parameters (an array parameter is already a caller pointer, safe to return).
func (c *escapeChecker) markArrayDecl(id *ast.Ident) bool {
	obj := c.info.Defs[id]
	if obj == nil || !isUnderlyingArray(obj.Type()) {
		return false
	}
	return c.markVar(id)
}

// markAddAssign marks the local that a string += turns into a frame value:
// it emits so_string_add, whose result is fresh frame memory.
func (c *escapeChecker) markAddAssign(s *ast.AssignStmt) bool {
	if s.Tok != token.ADD_ASSIGN || len(s.Lhs) != 1 {
		return false
	}
	if !isStringExpr(c.info, s.Lhs[0]) {
		return false
	}
	return c.markVar(s.Lhs[0])
}

// markVar records that a write to expr produces a frame value and reports whether
// that changed the set. Assigning a variable marks just that variable: q := &p
// makes q a frame value but leaves p untouched. Storing through it (p.s, s[i], *p)
// also marks whatever the variable points to, since the frame value lands there
// rather than in the variable.
func (c *escapeChecker) markVar(expr ast.Expr) bool {
	obj := c.rootLocal(expr)
	if obj == nil {
		return false
	}
	changed := c.mark(obj)
	if _, ok := ast.Unparen(expr).(*ast.Ident); ok {
		return changed
	}
	for _, target := range c.points[obj] {
		changed = c.mark(target) || changed
	}
	return changed
}

// mark records obj as holding a frame value and reports whether that changed the set.
func (c *escapeChecker) mark(obj types.Object) bool {
	if c.fvars[obj] {
		return false
	}
	c.fvars[obj] = true
	return true
}

// rootLocal returns the local variable that an assignment target writes to.
// Marking the whole variable instead of just the field isn't precise, but it's
// safer: once an aggregate contains a frame pointer, returning a copy is unsafe.
// It returns nil if the target doesn't root at a local variable.
func (c *escapeChecker) rootLocal(expr ast.Expr) types.Object {
	obj := c.rootObject(expr)
	if obj == nil || !c.locals[obj] {
		return nil
	}
	return obj
}

// rootGlobal returns the package-level variable that an assignment target writes
// to. It returns nil if the target doesn't root at a package-level variable.
func (c *escapeChecker) rootGlobal(expr ast.Expr) types.Object {
	obj := c.rootObject(expr)
	if obj == nil || !isPkgVar(obj) {
		return nil
	}
	return obj
}

// rootObject returns the variable that an assignment target writes to:
// p.s and p.x.y both root at p, s[i] roots at s, and *p roots at p. A qualified
// name roots at the variable of the other package, not at the package name.
func (c *escapeChecker) rootObject(expr ast.Expr) types.Object {
	for {
		switch x := expr.(type) {
		case *ast.ParenExpr:
			expr = x.X
		case *ast.SelectorExpr:
			if obj := c.info.Uses[x.Sel]; isPkgVar(obj) {
				return obj
			}
			expr = x.X
		case *ast.IndexExpr:
			expr = x.X
		case *ast.StarExpr:
			expr = x.X
		case *ast.Ident:
			return c.info.ObjectOf(x)
		default:
			return nil
		}
	}
}

// isFrameValue reports whether the value of expr lives in the current frame. It
// follows only the parts that pass the value along (a subexpression, a slice of
// it), never the opaque result of a call.
func (c *escapeChecker) isFrameValue(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.ParenExpr:
		return c.isFrameValue(x.X)
	case *ast.Ident:
		obj := c.info.ObjectOf(x)
		return obj != nil && c.fvars[obj]
	case *ast.BinaryExpr:
		return c.isFrameConcat(x)
	case *ast.UnaryExpr:
		return x.Op == token.AND && c.isFrameAddress(x.X)
	case *ast.CompositeLit:
		return c.isFrameComposite(x)
	case *ast.SliceExpr:
		return c.isFrameValue(x.X)
	case *ast.CallExpr:
		return c.isFrameCall(x)
	case *ast.SelectorExpr, *ast.IndexExpr, *ast.StarExpr:
		return c.isFrameRead(e)
	}
	return false
}

// isFrameRead reports whether reading out of a local carries frame memory with it:
// p.s and s[i] on a local that holds a frame value, *p on a pointer to one. Marking
// is per variable (see rootLocal), so the read checks the root. It carries frame
// memory only when what it reads can reference memory elsewhere: p.n on an int field
// is a plain copy even when p is marked.
func (c *escapeChecker) isFrameRead(expr ast.Expr) bool {
	obj := c.rootLocal(expr)
	if obj == nil || !c.fvars[obj] {
		return false
	}
	return carriesPointers(c.info.TypeOf(expr))
}

// isFrameConcat reports whether a binary expression is a string + that emits
// so_string_add, whose result is a fresh frame value. A + of two string literals
// folds to a static constant and is safe.
func (c *escapeChecker) isFrameConcat(x *ast.BinaryExpr) bool {
	if x.Op != token.ADD || !isStringExpr(c.info, x.X) {
		return false
	}
	return !(isStringLit(c.info, x.X) && isStringLit(c.info, x.Y))
}

// isFrameAddress reports whether &operand points into the current frame: the
// address of a local or parameter, or of a composite-literal temporary.
func (c *escapeChecker) isFrameAddress(operand ast.Expr) bool {
	switch o := operand.(type) {
	case *ast.CompositeLit:
		return true
	case *ast.Ident:
		obj := c.info.ObjectOf(o)
		return obj != nil && c.locals[obj]
	}
	return false
}

// isFrameComposite reports whether a composite literal is a frame value.
func (c *escapeChecker) isFrameComposite(x *ast.CompositeLit) bool {
	switch c.info.TypeOf(x).Underlying().(type) {
	case *types.Array:
		// An array literal translates to a bare C array in the frame.
		return true
	case *types.Map:
		// A map literal translates to so_map_lit, which calls so_make_map;
		// an empty one translates to a &(so_Map){} compound literal. Both
		// live in the frame.
		return true
	case *types.Slice:
		// A non-empty slice literal translates to a so_Slice over a
		// compound-literal backing array in the frame (see emitSliceLit);
		// an empty one translates to a null slice and is safe.
		return len(x.Elts) > 0
	}
	// A struct literal is a frame value when one of its elements is (e.g. BoxStr{s: a + b}).
	return c.hasFrameElem(x.Elts)
}

// isFrameCall reports whether a call produces a frame value. Conversions and the
// make, new and append builtins can; every other call is opaque and assumed not to.
func (c *escapeChecker) isFrameCall(call *ast.CallExpr) bool {
	if tv, ok := c.info.Types[call.Fun]; ok && tv.IsType() {
		return c.isFrameConversion(tv.Type, call)
	}
	id, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	b, ok := c.info.Uses[id].(*types.Builtin)
	if !ok {
		return false
	}
	switch b.Name() {
	case "make":
		// make of a slice or map emits so_make_slice / so_make_map.
		switch c.info.TypeOf(call).Underlying().(type) {
		case *types.Slice, *types.Map:
			return true
		}
	case "new":
		// Every form of new emits the address of a compound literal
		// in the frame (see emitNewCall).
		return true
	case "append":
		return c.isFrameAppend(call)
	}
	return false
}

// isFrameAppend reports whether an append produces a frame value.
//
// append(dst, el1, el2, ... eln) copies the elements into the existing backing
// array of the dst slice (it never reallocates, just checks capacity), so the
// result shares that backing array. It is a frame value if either dst or any
// of the elements are frame values.
func (c *escapeChecker) isFrameAppend(call *ast.CallExpr) bool {
	if len(call.Args) == 0 {
		return false
	}
	if c.isFrameValue(call.Args[0]) {
		return true
	}
	if call.Ellipsis.IsValid() {
		return c.isFrameSpread(call)
	}
	return slices.ContainsFunc(call.Args[1:], c.isFrameElem)
}

// isFrameSpread reports whether append(dst, src...) carries frame memory in
// from src. It copies the elements of src, not its header, so src being a frame
// value only matters when the element type can reference the frame: appending a
// frame []byte copies plain bytes, appending a frame []string copies headers
// that point into it.
func (c *escapeChecker) isFrameSpread(call *ast.CallExpr) bool {
	dst, ok := c.info.TypeOf(call.Args[0]).Underlying().(*types.Slice)
	if !ok || !carriesPointers(dst.Elem()) {
		return false
	}
	return c.isFrameValue(call.Args[1])
}

// hasFrameElem reports whether any of the elements stored into an aggregate
// carries frame memory into it.
func (c *escapeChecker) hasFrameElem(elts []ast.Expr) bool {
	for _, el := range elts {
		if kv, ok := el.(*ast.KeyValueExpr); ok {
			el = kv.Value
		}
		if c.isFrameElem(el) {
			return true
		}
	}
	return false
}

// isFrameElem reports whether storing el into an aggregate carries frame memory into it.
func (c *escapeChecker) isFrameElem(el ast.Expr) bool {
	t := c.info.TypeOf(el)
	// A pointer-free element is copied whole into the aggregate, so its frame
	// storage is copied too and does not escape.
	if !carriesPointers(t) {
		return false
	}
	// An array literal is written out element by element (BoxArrStr{a: {x, y}}),
	// so it carries frame memory only when one of its own elements does.
	if lit, ok := el.(*ast.CompositeLit); ok && isUnderlyingArray(t) {
		return c.hasFrameElem(lit.Elts)
	}
	return c.isFrameValue(el)
}

// isFrameConversion reports whether a type conversion produces a frame value.
// []rune(s) allocates a fresh buffer in the frame; []byte(s) is a zero-copy view,
// so it is a frame value only when s is. Conversions to string are handled by
// isFrameStringConv. Any other conversion reuses its argument's storage, so it is
// a frame value only when the argument is.
func (c *escapeChecker) isFrameConversion(target types.Type, call *ast.CallExpr) bool {
	if len(call.Args) != 1 {
		return false
	}
	// A constant conversion becomes a literal. A literal needs no frame storage.
	if c.info.Types[call].Value != nil {
		return false
	}
	arg := call.Args[0]
	argT := c.info.TypeOf(arg)
	switch t := target.Underlying().(type) {
	case *types.Slice:
		if b, ok := t.Elem().Underlying().(*types.Basic); ok && isStringType(argT) {
			switch b.Kind() {
			case types.Int32:
				return true // []rune(s): so_string_runes
			case types.Byte:
				return c.isFrameValue(arg) // []byte(s): view
			}
		}
	case *types.Basic:
		if t.Kind() == types.String {
			return c.isFrameStringConv(argT, arg)
		}
	}
	return c.isFrameValue(arg)
}

// isFrameStringConv reports whether a conversion to string produces a frame
// value. string([]rune) and string(integer) allocate a fresh buffer in the
// frame. string([]byte) is a zero-copy view, so it is a frame value only when
// its argument is.
func (c *escapeChecker) isFrameStringConv(argT types.Type, arg ast.Expr) bool {
	switch a := argT.Underlying().(type) {
	case *types.Slice:
		if b, ok := a.Elem().Underlying().(*types.Basic); ok {
			switch b.Kind() {
			case types.Int32:
				return true // string([]rune): so_runes_string
			case types.Byte:
				return c.isFrameValue(arg) // string([]byte): view
			}
		}
	case *types.Basic:
		if a.Info()&types.IsInteger != 0 {
			return true // string(integer): so_int_string
		}
	}
	return c.isFrameValue(arg)
}

// escapes returns the return-value expressions that are frame values. It skips
// nested closures. A bare return has no results, so it never escapes (Solod has no
// named results, so bare returns occur only in functions that return nothing).
func (c *escapeChecker) escapes(decl *ast.FuncDecl) []ast.Node {
	var found []ast.Node
	c.walk(decl.Body, func(n ast.Node) {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return
		}
		for _, r := range ret.Results {
			if c.isFrameValue(r) {
				found = append(found, r)
			}
		}
	})
	return found
}

// globalStores returns the expressions that store a frame value
// into a package-level variable.
func (c *escapeChecker) globalStores(body *ast.BlockStmt) []ast.Node {
	return c.frameStores(body, func(target ast.Expr) bool {
		return c.rootGlobal(target) != nil
	})
}

// paramStores returns the expressions that store a frame value
// into the out parameter of the caller.
func (c *escapeChecker) paramStores(body *ast.BlockStmt) []ast.Node {
	return c.frameStores(body, c.writesThroughParam)
}

// frameStores returns the expressions that store a frame value
// into a target the outside predicate accepts.
func (c *escapeChecker) frameStores(body *ast.BlockStmt, outside func(ast.Expr) bool) []ast.Node {
	var found []ast.Node
	c.walk(body, func(n ast.Node) {
		// A string += emits so_string_add, whose result is fresh frame memory.
		if target, ok := addAssignTarget(c.info, n); ok && outside(target) {
			found = append(found, target)
		}
		assignPairs(n, func(lhs, rhs ast.Expr) {
			if outside(lhs) && c.isFrameValue(rhs) {
				found = append(found, rhs)
			}
		})
	})
	return found
}

// writesThroughParam reports whether writing to expr lands in the caller-owned
// memory behind a parameter. The write must go through a pointer or similar
// indirection; otherwise, it just writes to the parameter's own storage.
func (c *escapeChecker) writesThroughParam(expr ast.Expr) bool {
	if !crossesIndirection(c.info, expr) {
		return false
	}
	obj := c.rootLocal(expr)
	return obj != nil && c.params[obj]
}

// addAssignTarget returns the target of a string +=.
func addAssignTarget(info *types.Info, n ast.Node) (ast.Expr, bool) {
	s, ok := n.(*ast.AssignStmt)
	if !ok || s.Tok != token.ADD_ASSIGN || len(s.Lhs) != 1 {
		return nil, false
	}
	if !isStringExpr(info, s.Lhs[0]) {
		return nil, false
	}
	return s.Lhs[0], true
}

// walk visits every node under root but does not enter nested closures, which
// have their own frame.
func (c *escapeChecker) walk(root ast.Node, visit func(ast.Node)) {
	ast.Inspect(root, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		visit(n)
		return true
	})
}

// assignPairs calls fn for each (target, value) pair that n assigns, covering
// both plain assignments and var declarations with initializers. A single
// multi-value RHS (x, y := f()) is an opaque call, so it pairs nothing.
func assignPairs(n ast.Node, fn func(lhs, rhs ast.Expr)) {
	switch s := n.(type) {
	case *ast.AssignStmt:
		if s.Tok == token.ASSIGN || s.Tok == token.DEFINE {
			pairUp(s.Lhs, s.Rhs, fn)
		}
	case *ast.DeclStmt:
		gd, ok := s.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			return
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			names := make([]ast.Expr, len(vs.Names))
			for i, name := range vs.Names {
				names[i] = name
			}
			pairUp(names, vs.Values, fn)
		}
	}
}

// pairUp calls fn for each lhs/rhs pair, or nothing if they do not line up.
func pairUp(lhs, rhs []ast.Expr, fn func(lhs, rhs ast.Expr)) {
	if len(lhs) != len(rhs) {
		return
	}
	for i := range lhs {
		fn(lhs[i], rhs[i])
	}
}

// carriesPointers reports whether a value of type t can reference memory outside
// itself. Copying a pointer-free value into an aggregate copies the whole value,
// so it can never carry frame memory along with it.
func carriesPointers(t types.Type) bool {
	if t == nil {
		return true
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		// A string is a pointer and a length; every other basic type is inline.
		return isStringType(u) || u.Kind() == types.UnsafePointer
	case *types.Array:
		return carriesPointers(u.Elem())
	case *types.Struct:
		for field := range u.Fields() {
			if carriesPointers(field.Type()) {
				return true
			}
		}
		return false
	}
	// Pointers, slices, maps, channels, interfaces and functions
	// all reference memory elsewhere.
	return true
}

// crossesIndirection reports whether an assignment target reaches
// its memory through a pointer, a slice, a map or an array.
func crossesIndirection(info *types.Info, expr ast.Expr) bool {
	for {
		switch x := expr.(type) {
		case *ast.ParenExpr:
			expr = x.X
		case *ast.StarExpr:
			return true
		case *ast.SelectorExpr:
			if isIndirect(info.TypeOf(x.X)) {
				return true
			}
			expr = x.X
		case *ast.IndexExpr:
			if isIndirect(info.TypeOf(x.X)) {
				return true
			}
			expr = x.X
		default:
			return false
		}
	}
}

// isIndirect reports whether a write into a value of type t
// reaches memory elsewhere.
func isIndirect(t types.Type) bool {
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Array:
		return true
	}
	return false
}

// isPkgVar reports whether obj is a variable declared at package level.
func isPkgVar(obj types.Object) bool {
	v, ok := obj.(*types.Var)
	if !ok || v.Pkg() == nil {
		return false
	}
	return v.Parent() == v.Pkg().Scope()
}

// isStringExpr reports whether expr has string type.
func isStringExpr(info *types.Info, expr ast.Expr) bool {
	return isStringType(info.TypeOf(expr))
}

// isStringType reports whether t is a string type.
func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && (b.Kind() == types.String || b.Kind() == types.UntypedString)
}

// isUnderlyingArray reports whether t is an array, named or not. Solod returns an
// array by value as a pointer into the frame (see returnType), so a local array
// is a frame value. A struct that wraps an array is copied by value and stays
// safe.
func isUnderlyingArray(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Array)
	return ok
}
