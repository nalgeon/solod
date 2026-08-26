// Package ctype declares extern C types. The main package does not import it,
// so ctype reaches main only through sub.
package ctype

//so:extern int
type Int int32
