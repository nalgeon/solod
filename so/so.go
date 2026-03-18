// Package so provides common utilities for other stdlib packages.
package so

import _ "embed"

//so:embed so.h
var so_h string

//so:extern
const MinInt32 = int32(-1 << 31)

//so:extern
const MaxInt32 = int32(1<<31 - 1)

//so:extern
const MinInt64 = int64(-1 << 63)

//so:extern
const MaxInt64 = int64(1<<63 - 1)

//so:extern
const MaxUint32 = uint32(1<<32 - 1)

//so:extern
const MaxUint64 = uint64(1<<64 - 1)
