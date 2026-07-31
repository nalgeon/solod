package main

import "example/sub"

// File-level constants.
const fInt int = 42
const fString string = "file"

// Using _ on file level is not supported.
// var _ = fInt
// var _ = fString

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
		_ = lInt
		const lFloat = 3e20 / lInt
		_ = lFloat
		const lString = "local"
		_ = lString
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
		// Using _ on file level is not supported,
		// so silence the unused file-level constants here.
		_ = fInt
		_ = fString
	}
}
