package main

// An exported identifier gets a package prefix.
var NULL = 0

// A conflicting unexported package-level identifier.
var stderr = 1

func typeof() int {
	return stderr
}

// Conflicting C keywords used as parameter names.
func scale(long int, register int) int {
	total := long * register
	return total
}

// A mangled parameter (long -> long_) and a same-named local in a nested
// block are a legal C shadow, not a conflict, so both are accepted.
func shadow(long int) int {
	if long > 0 {
		long_ := 99
		return long_
	}
	return long
}

// A name that looks like a generated temporary is not reserved: the switch
// tag temporary picks the next free name instead of shadowing the variable.
func switchTemp(x int) int {
	_sw1 := 99
	switch x {
	case 1:
		return _sw1
	}
	return 0
}

func pair() (int, int) {
	return 1, 2
}

// The same for the temporary that holds a multi-value call result.
func resultTemp() int {
	_res1 := 99
	a, b := pair()
	return _res1 + a + b
}

// A function pointer field with a conflicting parameter name.
type movie struct {
	rate func(long int) int
}

// An interface method with a conflicting parameter name.
type rater interface {
	rate(register int) int
}

// Conflicting local function variables.
func varDecl() int {
	var double int = 5
	var union, enum = 1, 2
	return double + union + enum
}

func main() {
	// Conflicting local variables.
	long := 10
	short := 20
	value := scale(long, short)
	_ = value
	_ = shadow(value)
	_ = switchTemp(1)
	_ = typeof()
	_ = resultTemp()
	_ = varDecl()

	// The name is mangled everywhere it is used.
	for bool := 0; bool < long; bool++ {
		b := bool
		_ = b
	}

	var m movie
	var r rater
	_ = m
	_ = r
}
