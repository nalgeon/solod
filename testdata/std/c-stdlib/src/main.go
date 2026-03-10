package main

import (
	"unsafe"

	"github.com/nalgeon/solod/so/c/stdlib"
)

func main() {
	{
		// Constants.
		status := stdlib.ExitSuccess
		if status == stdlib.ExitFailure {
			panic("unexpected failure")
		}
	}
	{
		// String-to-number conversion.
		n := stdlib.Atoi("42")
		if n != 42 {
			panic("want n == 42")
		}

		f := stdlib.Atof("3.14")
		if f < 3.0 {
			panic("want f >= 3.0")
		}
	}
	{
		// Memory management.
		ptr := stdlib.Malloc(unsafe.Sizeof(int(0)))
		if ptr == nil {
			panic("malloc failed")
		}
		stdlib.Free(ptr)

		ptr = stdlib.Calloc(10, unsafe.Sizeof(int(0)))
		if ptr == nil {
			panic("calloc failed")
		}
		ptr = stdlib.Realloc(ptr, 20*unsafe.Sizeof(int(0)))
		if ptr == nil {
			panic("realloc failed")
		}
		stdlib.Free(ptr)
	}
	{
		// Environment.
		env := stdlib.Getenv("PATH")
		if env == nil {
			panic("PATH not set")
		}
	}
	{
		// Exit (must be last).
		stdlib.Exit(0)
	}
}
