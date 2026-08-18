package main

import "example/sub"

type Reader interface {
	Read(buf int) (int, error)
}

type File struct {
	size int
}

func makeFile(size int) (File, error) {
	return File{size: size}, nil
}

func (f *File) Read(buf int) (int, error) {
	_ = buf
	return f.size, nil
}

var file File

type point struct {
	x, y int
}

func makePoint(x, y int) (point, error) {
	return point{x: x, y: y}, nil
}

func makeSubPoint(x, y int) (sub.Point, error) {
	return sub.Point{X: x, Y: y}, nil
}

func divide(a, b int) (int, error) {
	return a / b, nil
}

func returnInt() (int, error)            { return 42, nil }
func returnRune() (rune, error)          { return 'x', nil }
func returnString() (string, error)      { return "hello", nil }
func returnSlice(s []int) ([]int, error) { return s, nil }
func returnStruct() (File, error)        { return File{size: 42}, nil }
func returnAny() (any, error)            { return &file, nil }
func returnPtr() (*File, error)          { return &file, nil }

// Nil and untyped values need the same conversion as a single return.
func returnNilSlice() ([]int, error)  { return nil, nil }
func returnNilPtr() (*File, error)    { return nil, nil }
func returnAnyInt(v any) (any, error) { return v, nil }
func returnAnyPair(v any) (any, bool) { return v, true }

// func returnIface() (Reader, error)  { return &file, nil }

func forwardInt() (int, error)            { return returnInt() }
func forwardRune() (rune, error)          { return returnRune() }
func forwardString() (string, error)      { return returnString() }
func forwardSlice(s []int) ([]int, error) { return returnSlice(s) }
func forwardStruct() (File, error)        { return returnStruct() }
func forwardAny() (any, error)            { return returnAny() }
func forwardPtr() (*File, error)          { return returnPtr() }

func testBasic() {
	// Destructure into new variables.
	q, err := divide(10, 3)
	_ = q
	_ = err

	// Blank identifier.
	_, err2 := divide(10, 3)
	_ = err2
	r3, _ := divide(10, 3)
	_ = r3

	// Partial reassignment.
	r4, err2 := divide(10, 3)
	_ = r4

	// Assign to existing variables.
	q = 0
	err = nil
	q, err = divide(20, 7)
}

func testIf() {
	if n, err := divide(10, 3); err != nil {
		_ = n
	}
}

func testReturnTypes() {
	var err error
	_ = err
	i, err := returnInt()
	_ = i
	run, err := returnRune()
	_ = run
	str, err := returnString()
	_ = str
	slice, err := returnSlice(nil)
	_ = slice
	struc, err := returnStruct()
	_ = struc
	// iface, err := returnIface()
	// _ = iface
	a, err := returnAny()
	_ = a
	ptr, err := returnPtr()
	_ = ptr
}

func testNilValues() {
	nilSlice, err := returnNilSlice()
	if nilSlice != nil || len(nilSlice) != 0 || err != nil {
		panic("Nil slice in multi-return failed")
	}
	nilPtr, err := returnNilPtr()
	if nilPtr != nil || err != nil {
		panic("Nil pointer in multi-return failed")
	}
	n := 42
	anyInt, err := returnAnyInt(&n)
	if anyInt.(int) != n || err != nil {
		panic("Any value in multi-return failed")
	}
	anyPair, ok := returnAnyPair(&n)
	if anyPair.(int) != n || !ok {
		panic("Any value in (T, T) return failed")
	}
}

func testForwarding() {
	var err error
	_ = err
	i, err := forwardInt()
	_ = i
	r, err := forwardRune()
	_ = r
	str, err := forwardString()
	_ = str
	slice, err := forwardSlice(nil)
	_ = slice
	struc, err := forwardStruct()
	_ = struc
	a, err := forwardAny()
	_ = a
	ptr, err := forwardPtr()
	_ = ptr
}

func testStructExported() {
	f, err := makeFile(42)
	if f.size != 42 || err != nil {
		panic("Custom exported struct failed")
	}
}

func testStructUnexported() {
	p, err := makePoint(1, 2)
	if p.x != 1 || p.y != 2 || err != nil {
		panic("Custom unexported struct failed")
	}
}

func testStructOtherPackage() {
	sp1, err := makeSubPoint(1, 2)
	if sp1.X != 1 || sp1.Y != 2 || err != nil {
		panic("Custom struct from another package failed")
	}
	sp2, err := sub.MakePoint(3, 4)
	if sp2.X != 3 || sp2.Y != 4 || err != nil {
		panic("Custom struct from another package failed")
	}
}

func main() {
	testBasic()
	testIf()
	testReturnTypes()
	testNilValues()
	testForwarding()
	testStructExported()
	testStructUnexported()
	testStructOtherPackage()
}
