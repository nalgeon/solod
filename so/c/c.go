// Package c provides convenience helpers for C interop.
// It bridges C's null-terminated strings and raw pointers
// with So's string and slice types.
package c

import "unsafe"

//so:embed c.h
var c_h string

// Char represents a C char type.
//
//so:extern char
type Char byte

// ConstChar represents a C char type with a const modifier.
//
//so:extern const char
type ConstChar byte

// SChar represents a C signed char type.
//
//so:extern signed char
type SChar int8

// UChar represents a C unsigned char type.
//
//so:extern unsigned char
type UChar uint8

// Short represents a C short type.
//
//so:extern short
type Short int16

// UShort represents a C unsigned short type.
//
//so:extern unsigned short
type UShort uint16

// Int represents a C int type.
//
//so:extern int
type Int int32

// UInt represents a C unsigned int type.
//
//so:extern unsigned int
type UInt uint32

// Long represents a C long type.
//
//so:extern long
type Long int64

// ULong represents a C unsigned long type.
//
//so:extern unsigned long
type ULong uint64

// LongLong represents a C long long type.
//
//so:extern long long
type LongLong int64

// ULongLong represents a C unsigned long long type.
//
//so:extern unsigned long long
type ULongLong uint64

// LongDouble represents a C long double type.
//
//so:extern long double
type LongDouble float64

// Size represents a C size_t type.
//
//so:extern size_t
type Size uint

// SSize represents a C ssize_t type.
// In freestanding builds it maps to intptr_t.
//
//so:extern so_ssize_t
type SSize int

// Ptrdiff represents a C ptrdiff_t type.
//
//so:extern ptrdiff_t
type Ptrdiff int

// Intptr represents a C intptr_t type.
//
//so:extern intptr_t
type Intptr int

// ConstVoid represents a C void type with a const modifier.
// Use *ConstVoid for a C const void* pointer.
//
//so:extern const void
type ConstVoid byte

// Alignof returns the alignment of type T in bytes.
//
//	alignof(T)
//
//so:extern
func Alignof[T any]() int {
	var v T
	return int(unsafe.Alignof(v))
}

// Alloca allocates an array of the given length
// on the stack and returns a pointer to it.
//
//so:extern
func Alloca[T any](n int) *T {
	v := make([]T, n)
	return &v[0]
}

// Assert panics with the given message if the condition is false.
// The -assert=off build flag removes the check entirely,
// so cond must be free of side effects.
//
//	so_assert(cond, msg)
//
//so:extern so_assert
func Assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

// Assume tells the C compiler that cond is always true.
// It generates no code in any build, and the -assert=off build flag
// has no effect on it. The behavior is undefined if cond is false.
//
// Use Assume only for conditions that are provably true, such as when
// a pointer is known to be non-null but the compiler cannot see it.
// Use [Assert] for all other conditions.
//
//	so_assume(cond)
//
//so:extern so_assume
func Assume(cond bool) {
	if !cond {
		panic("c: false assumption")
	}
}

// Bytes wraps a raw byte pointer and length into a []byte without copying.
// If ptr is nil, returns nil.
//
//	(so_Slice){ptr, n, n}
//
//so:extern
func Bytes(ptr *byte, n int) []byte {
	if ptr == nil {
		return nil
	}
	return unsafe.Slice(ptr, n)
}

// CString converts a So string to a null-terminated C string.
// Allocates memory on the stack using alloca until the calling function returns.
//
//so:extern so_cstr nodecay
func CString(s string) *Char {
	return (*Char)(unsafe.StringData(s))
}

// PtrAdd adds offset to a pointer and returns the result.
// The offset counts elements of type T, not bytes.
//
//	ptr + offset
//
//so:extern
func PtrAdd[T any](ptr *T, offset int) *T {
	raw := ptrVal(ptr)
	p := unsafe.Add(raw, offset*Sizeof[T]())
	return (*T)(p)
}

// PtrAs casts a raw pointer (void*) to *T.
//
//	(T*)(ptr)
//
//so:extern
func PtrAs[T any](ptr any) *T {
	raw := ptrVal(ptr)
	return (*T)(raw)
}

// PtrAt returns a pointer to the element at the given index in an array or slice.
//
//	&ptr[index]
//
//so:extern
func PtrAt[T any](ptr *T, index int) *T {
	return PtrAdd(ptr, index)
}

// Raw emits a raw block of C code.
//
//so:extern
func Raw(code string) { _ = code }

// Sizeof returns the size of type T in bytes.
//
//	sizeof(T)
//
//so:extern
func Sizeof[T any]() int {
	var v T
	return int(unsafe.Sizeof(v))
}

// Slice wraps a raw pointer and length into a []T without copying.
// If ptr is nil, returns nil.
//
//	(so_Slice){ptr, len, cap}
//
//so:extern
func Slice[T any](ptr *T, len int, cap int) []T {
	if ptr == nil {
		return nil
	}
	s := unsafe.Slice(ptr, cap)
	return s[:len]
}

// SliceData returns a pointer to the underlying array of the slice,
// typed as *T. If the slice is nil, returns nil.
//
//	(T*)(s.ptr)
//
//so:extern
func SliceData[T, V any](s []V) *T {
	return (*T)(unsafe.Pointer(unsafe.SliceData(s)))
}

// String converts a null-terminated C string to a So string.
// If ptr is nil, returns "".
//
//	(so_String){s, strlen(s)}
//
//so:extern
func String[T Char | ConstChar](ptr *T) string { _ = ptr; return "" }

// StringData returns a pointer to the underlying bytes of the string,
// typed as *T. The bytes are read-only, so a write through the pointer
// is undefined behavior.
//
//	(T*)(s.ptr)
//
//so:extern
func StringData[T any](s string) *T {
	return (*T)(unsafe.Pointer(unsafe.StringData(s)))
}

// Val emits a typed C expression.
//
//so:extern
func Val[T any](expr string) T {
	var v T
	_ = expr
	return v
}

// Zero returns the zero value of type T.
//
//	{}
//
//so:extern
func Zero[T any]() T {
	var v T
	return v
}

// ptrVal extracts a raw pointer from an interface containing any pointer type.
// For testing only: in C any is void*, so unwrapping is unnecessary.
//
//so:extern
func ptrVal(v any) unsafe.Pointer {
	type iface struct {
		_    unsafe.Pointer
		data unsafe.Pointer
	}
	return (*iface)(unsafe.Pointer(&v)).data
}
