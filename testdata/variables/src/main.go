package main

import "unsafe"

type person struct {
	age int
}

type number *int

// Package-level variables.
var (
	pkgInt    int     = 42
	pkgFloat  float64 = 3.14
	pkgBool   bool    = true
	pkgByte   byte    = 'x'
	pkgRune   rune    = '本'
	pkgString string  = "hello"
	pkgSlice  []int   = []int{1, 2, 3}
	pkgStruct person  = person{age: 42}
	pkgPtr    *person = &pkgStruct
	pkgAnyVal any     = 42
	// pkgAnyPtr any     = pkgPtr
	pkgNil any = nil
)

// A package-level variable is visible to a function, so
// a call can read an assignment the same statement makes.
var counter int

func readCounter() int { return counter }

func main() {
	{
		// Definition with var and explicit type.
		var vInt int = 42
		_ = vInt
		var vFloat float64 = 3.14
		_ = vFloat
		var vBool bool = true
		_ = vBool
		var vByte byte = 'x'
		_ = vByte
		var vRune rune = '本'
		_ = vRune
		var vString string = "hello"
		_ = vString
		var vSlice []int = []int{1, 2, 3}
		_ = vSlice
		var vStruct = person{age: 42}
		var vPtr = &vStruct
		_ = vPtr
		var vAnyVal any = 42
		_ = vAnyVal
		var vAnyPtr any = vPtr
		_ = vAnyPtr
		var vNil any = nil
		_ = vNil
	}
	{
		// Definition with var and type inference.
		var vInt = 42
		_ = vInt
		var vFloat = 3.14
		_ = vFloat
		var vBool = true
		_ = vBool
		var vByte = 'x'
		_ = vByte
		var vRune = '本'
		_ = vRune
		var vString = "hello"
		_ = vString
		var vSlice = []int{1, 2, 3}
		_ = vSlice
		var vStruct = person{age: 42}
		var vPtr = &vStruct
		_ = vPtr
		var vAnyVal = any(42)
		_ = vAnyVal
		var vAnyPtr = any(vPtr)
		_ = vAnyPtr
		var vNil = any(nil)
		_ = vNil
	}
	{
		// Definition with short variable declaration.
		vInt := 42
		_ = vInt
		vFloat := 3.14
		_ = vFloat
		vBool := true
		_ = vBool
		vByte := 'x'
		_ = vByte
		vRune := '本'
		_ = vRune
		vString := "hello"
		_ = vString
		vSlice := []int{1, 2, 3}
		_ = vSlice
		vStruct := person{age: 42}
		vPtr := &vStruct
		_ = vPtr
		vAnyVal := any(42)
		_ = vAnyVal
		vAnyPtr := any(vPtr)
		_ = vAnyPtr
		vNil := any(nil)
		_ = vNil
	}
	{
		// Zero values.
		var vInt int
		_ = vInt
		var vFloat float64
		_ = vFloat
		var vBool bool
		_ = vBool
		var vByte byte
		_ = vByte
		var vRune rune
		_ = vRune
		var vString string
		_ = vString
		var vSlice []int
		_ = vSlice
		var vStruct person
		_ = vStruct
		var vPtr *person
		_ = vPtr
		var vNil any
		_ = vNil
	}
	{
		// Multiple typed variable declaration.
		var a, b, c int = 11, 22, 33
		_ = a
		_ = b
		_ = c
		var b1, b2 byte = 'a', 'b'
		_ = b1
		_ = b2
		var s1, s2 string = "foo", "bar"
		_ = s1
		_ = s2
		var a1, a2 []int = []int{1, 2}, []int{3, 4}
		_ = a1
		_ = a2
		var p1, p2 person = person{age: 42}, person{age: 43}
		_ = p1
		_ = p2
		var ptr1, ptr2 *person = &p1, &p2
		_ = ptr1
		_ = ptr2
		var n1, n2 number = &p1.age, &p2.age
		_ = n1
		_ = n2
	}
	{
		// Multiple untyped variable declaration.
		var vInt, vFloat, vBool = 42, 3.14, true
		_ = vInt
		_ = vFloat
		_ = vBool
		var vByte, vRune, vString = 'x', '本', "hello"
		_ = vByte
		_ = vRune
		_ = vString
		var vSlice, vStruct = []int{1, 2, 3}, person{age: 42}
		_ = vSlice
		_ = vStruct
		var ptr1, ptr2 = &vStruct, &vStruct
		_ = ptr1
		_ = ptr2
		var n1, n2 = number(&vStruct.age), number(&vStruct.age)
		_ = n1
		_ = n2
	}
	{
		// Multiple variable declaration with short variable declaration.
		vInt, vFloat, vBool := 42, 3.14, true
		_ = vInt
		_ = vFloat
		_ = vBool
		vByte, vRune, vString := 'x', '本', "hello"
		_ = vByte
		_ = vRune
		_ = vString
		vSlice, vStruct := []int{1, 2, 3}, person{age: 42}
		_ = vSlice
		_ = vStruct
		ptr1, ptr2 := &vStruct, &vStruct
		_ = ptr1
		_ = ptr2
		n1, n2 := number(&vStruct.age), number(&vStruct.age)
		_ = n1
		_ = n2
		u1, u2 := unsafe.Pointer(&vStruct), unsafe.Pointer(&vStruct)
		_ = u1
		_ = u2
	}
	{
		// Partial redeclaration with short variable declaration.
		a, x := 11, 100
		b, x := 22, 200
		x, c := 300, 33
		_ = a
		_ = b
		_ = c
		_ = x
	}
	{
		// Multiple assignment.
		a, b := 11, 22
		a, b = 33, 44
		x, y := 55, 66
		a, b = x, y
		if a != 55 || b != 66 {
			panic("multiple assignment failed")
		}
		p := person{age: 42}
		var ptr1, ptr2 *person
		ptr1, ptr2 = &p, &p
		_ = ptr1
		_ = ptr2
		var n1, n2 number
		n1, n2 = number(&p.age), number(&p.age)
		_ = n1
		_ = n2
	}
	{
		// Evaluates the whole right side before it assigns anything,
		// so a swap works and a call sees the values from before.
		a, b := 11, 22
		a, b = b, a
		if a != 22 || b != 11 {
			panic("swap failed")
		}
		s := []int{10, 20}
		i, j := 0, 1
		s[i], s[j] = s[j], s[i]
		if s[0] != 20 || s[1] != 10 {
			panic("slice swap failed")
		}
		counter = 7
		counter, b = 8, readCounter()
		if counter != 8 || b != 7 {
			panic("assignment order failed")
		}
	}
	{
		// Variable shadowing.
		age := 30
		p := person{age: age}
		{
			age := p.age
			_ = age
		}
		{
			age := person{age: 40}
			_ = age
		}
	}
}
