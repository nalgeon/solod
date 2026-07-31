package main

import "example/sub"

// File-level constants.
const fInt int = 42
const fString string = "file"

// Typedefed constant group.
type HttpStatus int

const (
	StatusOK       HttpStatus = 200
	StatusNotFound HttpStatus = 404
	StatusError    HttpStatus = 500
	statusSecret   HttpStatus = 999
)

// Regular constant group.
type ServerState string

const (
	StateIdle      ServerState = "idle"
	StateConnected ServerState = "connected"
	StateError     ServerState = "error"
)

// Iota constant group.
type Day int

const (
	Sunday Day = iota
	Monday
	Tuesday
)

// Using constants in other definitions.
const Zero = 42
const FortyTwo = Zero + 42

// Untyped constants above math.MaxInt64 are declared as uint64.
const MaxUint64 = 18446744073709551615
const halfUint64 = 9223372036854775808

const (
	bigIota = 1<<63 + iota
	bigIotaNext
)

// Float constant expressions are folded.
const (
	ln10   = 2.30258509299404568401799145468436420760110148862877297603332790
	log10E = 1 / ln10
)

// An integral value still has to read as a float in C.
const area float64 = 3*3 + 4*4

// An integer literal in a float context: too large for a C integer type.
const bigFloat float64 = 100000000000000000000000

// Iota in float constants.
const (
	step0 float64 = 0.25 * iota
	step1
	step2
)

type Point struct {
	X int
	Y int
}

var PointZero = Point{X: Zero, Y: Zero}
var PointSubZero = Point{X: sub.Zero, Y: sub.Zero}

func main() {
	{
		// Local constants.
		const lInt = 500000000
		const lFloat = 3e20 / lInt
		const lString = "local"
	}
	{
		// Using constants in expressions.
		status := StatusOK
		_ = status != StatusNotFound

		secret := statusSecret
		_ = secret > StatusOK

		state := StateConnected
		_ = state == StateIdle
	}
	{
		// Using iota constants.
		day := Monday
		_ = day == Sunday
	}
	{
		// Arithmetic on constants above math.MaxInt64 stays unsigned.
		var third uint64 = MaxUint64 / 3
		if third != 6148914691236517205 {
			panic("MaxUint64 / 3")
		}
		var shifted uint64 = MaxUint64 >> 1
		if shifted != 9223372036854775807 {
			panic("MaxUint64 >> 1")
		}
		var half uint64 = halfUint64
		if half != 9223372036854775808 {
			panic("halfUint64")
		}
		var first uint64 = bigIota
		if first != 9223372036854775808 {
			panic("bigIota")
		}
		var next uint64 = bigIotaNext
		if next != 9223372036854775809 {
			panic("bigIotaNext")
		}
	}
	{
		// Float constant expressions are folded.
		// Compare through a variable, so that the comparison happens at
		// runtime on the emitted value instead of at compile time.
		var quo float64 = log10E
		if quo != 0.4342944819032518 {
			panic("1 / ln10")
		}
		var whole float64 = area
		if whole != 25 {
			panic("3*3 + 4*4")
		}
		var sum float64 = 0.1 + 0.2
		if sum != 0.3 {
			panic("0.1 + 0.2")
		}
		var wide float64 = 1e200 * 1e200 / 1e200
		if wide != 1e200 {
			panic("1e200 * 1e200 / 1e200")
		}
		var tiny float64 = 1e-300 * 1e-300 * 1e300
		if tiny != 1e-300 {
			panic("1e-300 * 1e-300 * 1e300")
		}
		var lost float64 = (1e20 + 1) - 1e20
		if lost != 1 {
			panic("(1e20 + 1) - 1e20")
		}
		var step float64 = step2
		if step != 0.5 {
			panic("0.25 * iota")
		}
	}
	{
		// An integer literal in a float context.
		var big float64 = 10000000000000000000000
		if big != 1e22 {
			panic("1e22")
		}
		var wider float64 = bigFloat
		if wider != 1e23 {
			panic("bigFloat")
		}
		var hex float64 = 0xFF
		if hex != 255 {
			panic("0xFF")
		}
	}
	{
		// Same for constants narrowed to float32.
		var sum float32 = 0.1 + 0.2
		if sum != 0.3 {
			panic("float32 0.1 + 0.2")
		}
		var wide float32 = 1e200 * 1e200 / 1e200 / 1e195
		if wide != 100000 {
			panic("float32 1e200 * 1e200 / 1e200 / 1e195")
		}
		var tiny float32 = 1e-40 * 1e40 * 1e-40
		if tiny != 1e-40 {
			panic("float32 1e-40 * 1e40 * 1e-40")
		}
		// A float32 literal compares as a float, not as a double.
		var lit float32 = 0.1
		if lit != 0.1 {
			panic("float32 0.1")
		}
	}
}
