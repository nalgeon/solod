// Package os provides a platform-independent interface to operating system
// functionality. The design is Unix-like, although the error handling is Go-like;
// failing calls return values of type error rather than error numbers.
package os

import (
	"github.com/nalgeon/solod/so/errors"
	"github.com/nalgeon/solod/so/io"
	"github.com/nalgeon/solod/so/mem"
)

var ErrClosed = errors.New("file already closed")
var ErrExist = errors.New("file already exists")
var ErrIsDir = errors.New("is a directory")
var ErrNotDir = errors.New("not a directory")
var ErrNotExist = errors.New("no such file or directory")
var ErrPermission = errors.New("permission denied")

var ErrIO = errors.New("i/o error")

// FileResult is a helper struct for returning
// a File and an error from a function.
type FileResult struct {
	val File
	err error
}

// Create creates or truncates the named file. If the file already exists,
// it is truncated. If the file does not exist, it is created with mode 0o666
// (before umask). If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
// The directory containing the file must already exist.
func Create(name string) (File, error) {
	fd := fopen(name, "w+b")
	if fd == nil {
		return File{}, mapError()
	}
	return File{fd: fd}, nil
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
func Open(name string) (File, error) {
	fd := fopen(name, "rb")
	if fd == nil {
		return File{}, mapError()
	}
	return File{fd: fd}, nil
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat
// an EOF from Read as an error to be reported.
//
// The returned slice is allocated; the caller owns it.
func ReadFile(a mem.Allocator, name string) ([]byte, error) {
	f, err := Open(name)
	if err != nil {
		return []byte{}, err
	}
	b, err := io.ReadAll(a, &f)
	f.Close()
	return b, err
}

// WriteFile writes data to the named file, creating it if necessary.
// If the file does not exist, WriteFile creates it.
// If the file exists, WriteFile truncates it before writing.
func WriteFile(name string, data []byte) error {
	f, err := Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// Rename renames (moves) oldpath to newpath. If newpath already exists
// and is not a directory, Rename replaces it. If newpath already exists
// and is a directory, Rename returns an error. OS-specific restrictions
// may apply when oldpath and newpath are in different directories.
// Even within the same directory, on non-Unix platforms Rename
// is not an atomic operation.
func Rename(oldpath, newpath string) error {
	if rename(oldpath, newpath) != 0 {
		return mapError()
	}
	return nil
}

// Remove removes the named file or (empty) directory.
func Remove(name string) error {
	if remove(name) != 0 {
		return mapError()
	}
	return nil
}

// mapError maps errno to a sentinel error.
func mapError() error {
	if errno == os_EACCES {
		return ErrPermission
	}
	if errno == os_EEXIST {
		return ErrExist
	}
	if errno == os_EISDIR {
		return ErrIsDir
	}
	if errno == os_ENOENT {
		return ErrNotExist
	}
	if errno == os_ENOTDIR {
		return ErrNotDir
	}
	return ErrIO
}
