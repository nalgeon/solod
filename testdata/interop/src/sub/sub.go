package sub

import "example/interop/src/sub/ctype"

//so:embed sub.h
var sub_h string

//so:extern Stream
type Stream struct {
	Write func(format string, args ...any)
}

//so:extern Discard
func Discard(format string, args ...any) {}

// The parameter type comes from the ctype package.
func Scale(n ctype.Int) int32 {
	return int32(n) * 2
}
