package main

// Conflicts with div from stdlib.h.
func div(d, r int64) int64 {
	return d/r + 1
}

// Conflicts with remove from stdio.h.
var remove = 10

// Conflicts with index from string.h.
const index = 2

// Conflicts with rand from stdlib.h.
type rand struct {
	// A struct field does not conflict with an stdlib.h function.
	free int
}

// An exported name gets the package prefix.
func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	{
		// A block-scope shadows the file-scope one.
		div := 12
		free := 3
		_ = div
		_ = free
	}

	n := div(12, 5)
	_ = n
	_ = Abs(-1)

	w := rand{free: remove + index}
	_ = w.free
}
