package os

import (
	"unsafe"

	"solod.dev/so/c"
	"solod.dev/so/io"
	"solod.dev/so/mem"
)

// Open flag constants.
//
//so:extern O_RDONLY
const O_RDONLY = 0x0000 // open the file read-only
//so:extern O_WRONLY
const O_WRONLY = 0x0001 // open the file write-only
//so:extern O_RDWR
const O_RDWR = 0x0002 // open the file read-write
//so:extern O_APPEND
const O_APPEND = 0x00000008 // open the file in append mode
//so:extern O_CREAT
const O_CREATE = 0x00000200 // create a new file if none exists
//so:extern O_EXCL
const O_EXCL = 0x00000800 // ensure that this call creates the file
//so:extern O_SYNC
const O_SYNC = 0x0080 // synchronous writes
//so:extern O_TRUNC
const O_TRUNC = 0x00000400 // truncate regular writable file when opened

// File represents an open file descriptor.
// A nil pointer and the zero File are invalid.
// The methods return [ErrInvalid] for an invalid file
// and [ErrClosed] for a closed file.
type File struct {
	fd     *os_file
	name   string
	closed bool
}

// Standard input, output, and error streams.
var stdin_ File
var Stdin *File
var stdout_ File
var Stdout *File
var stderr_ File
var Stderr *File

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
	return File{fd: fd, name: name}, nil
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
func Open(name string) (File, error) {
	fd := fopen(name, "rb")
	if fd == nil {
		return File{}, mapError()
	}
	return File{fd: fd, name: name}, nil
}

// OpenFile is the generalized open call; most users will use Open
// or Create instead. It opens the named file with specified flag
// ([O_RDONLY] etc.). If the file does not exist, and the [O_CREATE] flag
// is passed, it is created with mode perm (before umask);
// the containing directory must exist. If successful,
// methods on the returned File can be used for I/O.
func OpenFile(name string, flag int, perm FileMode) (File, error) {
	pmode := makePosixMode(perm)
	fd := posixOpen(name, c.Int(flag), c.Int(pmode))
	if fd < 0 {
		return File{}, mapError()
	}
	mode := fdopenMode(flag)
	fp := fdopen(fd, mode)
	if fp == nil {
		fdclose(fd)
		return File{}, mapError()
	}
	return File{fd: fp, name: name}, nil
}

// Name returns the name of the file as presented to Open/Create.
func (f *File) Name() string {
	return f.name
}

// Read reads up to len(b) bytes from the file and stores them in b.
// It returns the number of bytes read and any error encountered.
// At end of file, Read returns 0, io.EOF.
func (f *File) Read(b []byte) (int, error) {
	fd := f.stream()
	if fd == nil {
		return 0, f.invalidErr()
	}
	if len(b) == 0 {
		return 0, nil
	}
	n := int(fread(unsafe.SliceData(b), 1, uintptr(len(b)), fd))
	if n < len(b) {
		if ferror(fd) {
			return n, mapError()
		}
		if n == 0 {
			return 0, io.EOF
		}
	}
	return n, nil
}

// Write writes len(b) bytes from b to the file.
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (f *File) Write(b []byte) (int, error) {
	fd := f.stream()
	if fd == nil {
		return 0, f.invalidErr()
	}
	if len(b) == 0 {
		return 0, nil
	}
	n := int(fwrite(unsafe.SliceData(b), 1, uintptr(len(b)), fd))
	if n < len(b) {
		return n, mapError()
	}
	return n, nil
}

// Seek sets the offset for the next Read or Write on file to offset,
// interpreted according to whence: [io.SeekStart] means relative to
// the start of the file, [io.SeekCurrent] means relative to the current
// offset, and [io.SeekEnd] means relative to the end.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	fd := f.stream()
	if fd == nil {
		return 0, f.invalidErr()
	}
	if fseeko(fd, offset, c.Int(whence)) != 0 {
		return 0, mapError()
	}
	pos := ftello(fd)
	if pos < 0 {
		return 0, mapError()
	}
	return pos, nil
}

// ReadAt reads len(b) bytes from the file starting at byte offset off.
// It returns the number of bytes read and the error, if any.
// ReadAt always returns a non-nil error when n < len(b).
func (f *File) ReadAt(b []byte, off int64) (int, error) {
	fd := f.stream()
	if fd == nil {
		return 0, f.invalidErr()
	}
	if off < 0 {
		return 0, io.ErrOffset
	}
	cur := ftello(fd)
	if cur < 0 {
		return 0, mapError()
	}
	if fseeko(fd, off, io.SeekStart) != 0 {
		return 0, mapError()
	}
	n, err := f.Read(b)
	if fseeko(fd, cur, io.SeekStart) != 0 && err == nil {
		return n, mapError()
	}
	if n < len(b) && err == nil {
		err = io.EOF
	}
	return n, err
}

// WriteAt writes len(b) bytes to the file starting at byte offset off.
// It returns the number of bytes written and an error, if any.
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	fd := f.stream()
	if fd == nil {
		return 0, f.invalidErr()
	}
	if off < 0 {
		return 0, io.ErrOffset
	}
	cur := ftello(fd)
	if cur < 0 {
		return 0, mapError()
	}
	if fseeko(fd, off, io.SeekStart) != 0 {
		return 0, mapError()
	}
	n, err := f.Write(b)
	if fseeko(fd, cur, io.SeekStart) != 0 && err == nil {
		return n, mapError()
	}
	return n, err
}

// WriteString is like Write, but writes the contents of string s
// rather than a slice of bytes.
func (f *File) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}

// Sync writes the buffered data of the file to the operating system.
// The data can still wait in an operating system cache, so Sync does not
// guarantee that the data reached the storage device.
//
// A File writes through a buffered C stream. A program that ends abnormally
// loses the buffered data. Call Sync to keep the output that a later crash
// must not hide.
func (f *File) Sync() error {
	fd := f.stream()
	if fd == nil {
		return f.invalidErr()
	}
	if fflush(fd) != 0 {
		return mapError()
	}
	return nil
}

// Close closes the file, rendering it unusable for I/O.
// Close will return an error if it has already been called.
func (f *File) Close() error {
	fd := f.stream()
	if fd == nil {
		return f.invalidErr()
	}
	// fclose destroys the stream whatever it returns. The file is marked
	// closed before the result is checked, so that no later call can use
	// the destroyed stream.
	f.fd = nil
	f.closed = true
	if fclose(fd) != 0 {
		return mapError()
	}
	return nil
}

// stream returns the C stream of f, or nil if f is not usable for I/O.
// A nil pointer, an unopened file, and a closed file all give nil.
// Use [File.invalidErr] to report the reason.
func (f *File) stream() *os_file {
	if f == nil {
		return nil
	}
	return f.fd
}

// invalidErr returns the reason f is not usable for I/O.
// A closed file gives ErrClosed. A nil pointer or an unopened file
// gives ErrInvalid.
func (f *File) invalidErr() error {
	if f != nil && f.closed {
		return ErrClosed
	}
	return ErrInvalid
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat
// an EOF from Read as an error to be reported.
//
// If the allocator is nil, uses the system allocator.
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
// If the file does not exist, WriteFile creates it with permissions perm (before umask);
// otherwise WriteFile truncates it before writing, without changing permissions.
//
// Since WriteFile requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
func WriteFile(name string, data []byte, perm FileMode) error {
	f, err := OpenFile(name, O_WRONLY|O_CREATE|O_TRUNC, perm)
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

// fdopenMode returns the fdopen mode string for the given open flags.
func fdopenMode(flag int) string {
	switch flag & (O_RDONLY | O_WRONLY | O_RDWR) {
	case O_WRONLY:
		if flag&O_APPEND != 0 {
			return "ab"
		}
		return "wb"
	case O_RDWR:
		if flag&O_APPEND != 0 {
			return "a+b"
		}
		if flag&O_TRUNC != 0 {
			return "w+b"
		}
		return "r+b"
	default:
		return "rb"
	}
}

func init() {
	stdin_ = File{fd: stdin, name: "/dev/stdin"}
	Stdin = &stdin_
	stdout_ = File{fd: stdout, name: "/dev/stdout"}
	Stdout = &stdout_
	stderr_ = File{fd: stderr, name: "/dev/stderr"}
	Stderr = &stderr_
}
