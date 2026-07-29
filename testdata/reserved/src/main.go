package main

// An exported identifier gets a package prefix,
// so it doesn't need mangling.
var NULL = 0

// C keywords used as parameter names.
func scale(long int, register int) int {
	total := long * register
	return total
}

// A mangled parameter (long -> long_) and a same-named local in a nested
// block are a legal C shadow, not a collision, so both are accepted.
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

// A function pointer field with a reserved parameter name.
type movie struct {
	rate func(long int) int
}

// An interface method with a reserved parameter name.
type rater interface {
	rate(register int) int
}

func main() {
	// C keywords used as local variables.
	long := 10
	short := 20
	value := scale(long, short)
	_ = value
	_ = shadow(value)
	_ = switchTemp(1)
	_ = resultTemp()

	// The name should be mangled everywhere it is used.
	for bool := 0; bool < long; bool++ {
		b := bool
		_ = b
	}

	// Reference the reserved-parameter types so they are emitted.
	var m movie
	var r rater
	_ = m
	_ = r
}
