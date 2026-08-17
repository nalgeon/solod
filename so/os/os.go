// Package os provides a platform-independent interface to operating system
// functionality. The design is Unix-like, although the error handling is Go-like;
// failing calls return values of type error rather than error numbers.
package os

import (
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/errors"
	"solod.dev/so/time"
)

// MaxPathLen is the maximum length of a path.
const MaxPathLen = 4096

// MaxNameLen is the maximum length of a filename.
const MaxNameLen = 256

// IO-related errors that can be returned by functions in this package.
var ErrClosed = errors.New("os: file already closed")
var ErrExist = errors.New("os: file already exists")
var ErrInvalid = errors.New("os: invalid argument")
var ErrIsDir = errors.New("os: is a directory")
var ErrNotDir = errors.New("os: not a directory")
var ErrNotExist = errors.New("os: no such file or directory")
var ErrPermission = errors.New("os: permission denied")

// ErrIO is a generic I/O error that is returned when the error
// does not match any of the other, more specific errors.
var ErrIO = errors.New("os: i/o error")

// Chdir changes the current working directory to the named directory.
func Chdir(dir string) error {
	if chdir(dir) != 0 {
		return mapError()
	}
	return nil
}

// Chmod changes the mode of the named file to mode.
// If the file is a symbolic link, it changes the mode of the link's target.
func Chmod(name string, mode FileMode) error {
	pmode := makePosixMode(mode)
	if chmod(name, pmode) != 0 {
		return mapError()
	}
	return nil
}

// Chown changes the numeric uid and gid of the named file.
// If the file is a symbolic link, it changes the uid and gid of the link's target.
// A uid or gid of -1 means to not change that value.
func Chown(name string, uid, gid int) error {
	if chown(name, uid_t(uid), gid_t(gid)) != 0 {
		return mapError()
	}
	return nil
}

// Chtimes changes the access and modification times of the named
// file, similar to the Unix utime() or utimes() functions.
// A zero [time.Time] value will leave the corresponding file time unchanged.
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	var asec, ansec, msec, mnsec int64
	if atime.IsZero() {
		ansec = utimeOmit
	} else {
		asec, ansec = atime.Unix(), int64(atime.Nanosecond())
	}
	if mtime.IsZero() {
		mnsec = utimeOmit
	} else {
		msec, mnsec = mtime.Unix(), int64(mtime.Nanosecond())
	}
	if os_utimens(name, asec, ansec, msec, mnsec) != 0 {
		return mapError()
	}
	return nil
}

// Lchown changes the numeric uid and gid of the named file.
// If the file is a symbolic link, it changes the uid and gid of the link itself.
func Lchown(name string, uid, gid int) error {
	if lchown(name, uid_t(uid), gid_t(gid)) != 0 {
		return mapError()
	}
	return nil
}

// Link creates newname as a hard link to the oldname file.
func Link(oldname, newname string) error {
	if link(oldname, newname) != 0 {
		return mapError()
	}
	return nil
}

// Mkdir creates a new directory with the specified name and permission
// bits (before umask).
func Mkdir(name string, perm FileMode) error {
	pmode := makePosixMode(perm)
	if mkdir(name, pmode) != 0 {
		return mapError()
	}
	return nil
}

// Readlink returns the destination of the named symbolic link.
// If the link destination is relative, Readlink returns the relative path
// without resolving it to an absolute one.
//
// Writes the result into buf. Panics if buf is empty. A buf shorter than
// the destination gets the first len(buf) bytes of the destination,
// and Readlink reports no error. The returned string is a view into buf.
func Readlink(buf []byte, name string) (string, error) {
	bufPtr := unsafe.SliceData(buf)
	if bufPtr == nil {
		panic("os: empty buffer")
	}
	n := readlink(name, (*c.Char)(bufPtr), uintptr(len(buf)))
	if n < 0 {
		return "", mapError()
	}
	return unsafe.String(bufPtr, n), nil
}

// Remove removes the named file or (empty) directory.
func Remove(name string) error {
	if remove(name) != 0 {
		return mapError()
	}
	return nil
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

// SameFile reports whether fi1 and fi2 describe the same file.
// For example, on Unix this means that the device and inode fields
// of the two underlying structures are identical.
func SameFile(fi1, fi2 FileInfo) bool {
	return fi1.dev == fi2.dev && fi1.ino == fi2.ino
}

// Symlink creates newname as a symbolic link to oldname.
func Symlink(oldname, newname string) error {
	if symlink(oldname, newname) != 0 {
		return mapError()
	}
	return nil
}

// Truncate changes the size of the named file.
// If the file is a symbolic link, it changes the size of the link's target.
func Truncate(name string, size int64) error {
	if truncate(name, size) != 0 {
		return mapError()
	}
	return nil
}

// mapError maps errno to a sentinel error.
func mapError() error {
	if errno == eACCES {
		return ErrPermission
	}
	if errno == eEXIST {
		return ErrExist
	}
	if errno == eISDIR {
		return ErrIsDir
	}
	if errno == eNOENT {
		return ErrNotExist
	}
	if errno == eNOTDIR {
		return ErrNotDir
	}
	if errno == ePERM {
		return ErrPermission
	}
	return ErrIO
}
