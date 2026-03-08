package main

import (
	"github.com/nalgeon/solod/so/c"
	"github.com/nalgeon/solod/so/c/stdlib"
)

//so:include <string.h>

// char *strcat( char *dest, const char *src );
//
//so:extern
func strcat(dest *byte, src string) *byte

func main() {
	{
		// c.String: convert C string to So string.
		ptr := stdlib.Getenv("PATH")
		path := c.String(ptr)
		if len(path) == 0 {
			panic("want non-empty PATH")
		}
	}
	{
		// c.String: nil pointer returns empty string.
		ptr := stdlib.Getenv("SOLOD_NONEXISTENT_VAR")
		s := c.String(ptr)
		if len(s) != 0 {
			panic("want empty string for nil")
		}
	}
	{
		// c.Bytes: wrap a raw buffer into []byte.
		buf := stdlib.Malloc(4)
		if buf == nil {
			panic("malloc failed")
		}
		ptr := any(buf).(*byte)
		*ptr = 'H'
		slice := c.Bytes(ptr, 4)
		if len(slice) != 4 {
			panic("want len == 4")
		}
		if slice[0] != 'H' {
			panic("want slice[0] == 'H'")
		}
		stdlib.Free(buf)
	}
	{
		// Passing (char*) strings to C functions.
		var buf [64]byte
		strcat(c.CharPtr(&buf[0]), "Hello, ")
		strcat(c.CharPtr(&buf[0]), "world!")
		println(c.String(&buf[0]))
	}
	{
		// Returning (char*) strings from C functions.
		var buf [64]byte
		strcat(c.CharPtr(&buf[0]), "Hello, ")
		s := c.String(strcat(c.CharPtr(&buf[0]), "world!"))
		println(s)
	}
}
