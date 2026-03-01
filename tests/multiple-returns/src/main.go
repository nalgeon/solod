package main

type Reader interface {
	Read(buf int) (int, error)
}

type File struct {
	size int
}

var file File

func (f *File) Read(buf int) (int, error) {
	_ = buf
	return f.size, nil
}

func divide(a, b int) (int, error) {
	return a / b, nil
}

func returnRune() (rune, error) {
	return 'x', nil
}

func returnString() (string, error) {
	return "hello", nil
}

func returnSlice() ([]int, error) {
	return []int{1, 2, 3}, nil
}

// Returning struct values is not supported.
// func returnStruct() (File, error) {
// 	return File{size: 42}, nil
// }

func returnPtr() (*File, error) {
	return &file, nil
}

// Returning interface values is not supported.
// func returnIface() (Reader, error) {
// 	return &file, nil
// }

func forwardCall() (int, error) {
	return divide(10, 3)
}

func main() {
	{
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

	{
		// If-init with multi-return.
		f := File{size: 42}
		if n, err := f.Read(64); err != nil {
			_ = n
		}
	}

	{
		// Various return types.
		var err error
		_ = err
		run, err := returnRune()
		_ = run
		str, err := returnString()
		_ = str
		slice, err := returnSlice()
		_ = slice
		// struc, err := returnStruct()
		// _ = struc
		ptr, err := returnPtr()
		_ = ptr
		// iface, err := returnIface()
		// _ = iface
	}

	{
		// Forward call.
		q, err := forwardCall()
		_ = q
		_ = err
	}
}
