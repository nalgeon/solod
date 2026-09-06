package main

// Blank package-level constants.
const _ = 1
const _ = 2
const (
	_ = 3.14
	_ = "4"
)

// Iota with blank identifiers.
const (
	_ = iota
	_ = iota
	_ = iota
	iotaD
	iotaE
)

// Blank package-level variables.
var _ = 1
var _ = 2
var (
	_ = 3.14
	_ = "4"
)

// Blank package-level variables of different types.
var _ = 11
var _ int = 11
var _ float64 = 22.2
var _ string = "33"
var _ []int = []int{1, 2, 3}
var _ = Value{42}
var _ Value = Value{42}
var value = Value{42}
var _ *Value = &value
var _ any = 42
var _ any = &value
var _ any = nil

// A named struct with blank fields.
type Point struct {
	x int
	_ int
	y int
	_ int
}

// An inner struct field with blank fields.
type Wrapper struct {
	inner struct {
		n    int
		_, _ string
	}
}

// Struct with blank and unnamed parameters.
type Value struct{ x int }

// A named array type.
type Nums [3]int

// Unnamed receiver.
func (Value) One() int { return 1 }

// Blank receiver.
func (_ Value) Two() int { return 2 }

// Unnamed method parameters.
func (v Value) Decr1(int) int           { return v.x - 1 }
func (v *Value) Decr2(int, float32) int { return v.x - 2 }

// Blank method parameters.
func (v Value) Incr1(_ int) int                    { return v.x + 1 }
func (v Value) Incr2(n int, _ float32) int         { return v.x + n }
func (v *Value) Incr3(_ int, _ float32, n int) int { return v.x + n }

// Interface with unnamed and blank parameters.
type Valuer interface {
	Decr2(int, float32) int
	Incr3(_ int, _ float32, n int) int
}

// Unnamed function parameters.
func unnamed1(int) int          { return 1 }
func unnamed2(int, float32) int { return 2 }

// Blank function parameters.
func blank1(_ int) int                   { return 1 }
func blank2(n int, _ int) int            { return n }
func blank3(_ int, _ float32, n int) int { return n }

// Unnamed generic function parameters.

//so:inline
func unnamedGen1[T any](T) int { return 1 }

//so:inline
func unnamedGen2[T any](T, float32) int { return 2 }

// Blank generic function parameters.

//so:inline
func blankGen1[T any](_ T) int { return 1 }

//so:inline
func blankGen2[T any](n T, _ int) T { return n }

//so:inline
func blankGen3[T any](_ T, _ float32, n T) T { return n }

func main() {
	{
		if iotaD != 3 || iotaE != 4 {
			panic("unexpected iota values")
		}
	}
	{
		// Keyed literal, field assignment, and field access.
		p := Point{x: 1, y: 2}
		p.x = 3
		if p.x != 3 || p.y != 2 {
			panic("unexpected Point value")
		}
	}
	{
		// Positional literal also fills the blank fields.
		q := Point{4, 0, 5, 0}
		if q.x != 4 || q.y != 5 {
			panic("unexpected Point value")
		}
	}
	{
		// Inner struct field.
		var w Wrapper
		w.inner.n = 6
		if w.inner.n != 6 {
			panic("unexpected Wrapper value")
		}
	}
	{
		// Local anonymous struct.
		var st struct {
			f    float64
			_, _ string
		}
		st.f = 7
		if st.f != 7 {
			panic("unexpected st value")
		}
	}
	{
		// Anonymous struct literal.
		lit := struct {
			n int
			_ string
		}{8, "skip"}
		if lit.n != 8 {
			panic("unexpected lit value")
		}
	}
	{
		// Unnamed or blank method receiver.
		v := Value{42}
		if v.One() != 1 {
			panic("unexpected Value.One result")
		}
		if v.Two() != 2 {
			panic("unexpected Value.Two result")
		}
	}
	{
		// Unnamed or blank method parameter.
		v := Value{5}
		if v.Decr1(1) != 4 {
			panic("unexpected Value.Decr1 result")
		}
		if v.Decr2(1, 2) != 3 {
			panic("unexpected Value.Decr2 result")
		}
		if v.Incr1(1) != 6 {
			panic("unexpected Value.Incr1 result")
		}
		if v.Incr2(5, 6) != 10 {
			panic("unexpected Value.Incr2 result")
		}
		if v.Incr3(5, 6, 7) != 12 {
			panic("unexpected Value.Incr3 result")
		}
	}
	{
		// Interface methods with unnamed and blank parameters.
		var v Valuer = &Value{5}
		if v.Decr2(1, 2) != 3 {
			panic("unexpected Valuer.Decr2 result")
		}
		if v.Incr3(5, 6, 7) != 12 {
			panic("unexpected Valuer.Incr3 result")
		}
	}
	{
		// Unnamed or blank function parameters.
		if unnamed1(5) != 1 {
			panic("unexpected unnamed1 result")
		}
		if unnamed2(5, 6) != 2 {
			panic("unexpected unnamed2 result")
		}
		if blank1(5) != 1 {
			panic("unexpected blank1 result")
		}
		if blank2(5, 6) != 5 {
			panic("unexpected blank2 result")
		}
		if blank3(5, 6, 7) != 7 {
			panic("unexpected blank3 result")
		}
	}
	{
		// Unnamed or blank generic function parameters.
		if unnamedGen1(5) != 1 {
			panic("unexpected unnamedGen1 result")
		}
		if unnamedGen2(5, 6) != 2 {
			panic("unexpected unnamedGen2 result")
		}
		if blankGen1(5) != 1 {
			panic("unexpected blankGen1 result")
		}
		if blankGen2(5, 6) != 5 {
			panic("unexpected blankGen2 result")
		}
		if blankGen3(5, 6, 7) != 7 {
			panic("unexpected blankGen3 result")
		}
	}
	{
		// Discarding values with blank identifier.
		var v1, _ = 11, 12
		var _, v2 = 21, 22
		var _, _ = 31, 32
		var _ = 41

		v3, _ := 51, 52
		_, v4 := 61, 62
		_, _ = 71, 72
		_ = 81

		_ = v1
		_ = v2
		_ = v3
		_ = v4
	}
	{
		// Discarding an array literal.
		_ = [3]int{1, 2, 3}
		_ = [...]int{1, 2, 3}
		_ = [...][]int{{}, {}}
		_ = Nums{1, 2, 3}
		n1, _ := 1, [...]int{2, 3}
		_ = n1
	}
}
