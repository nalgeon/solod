package os

import "github.com/nalgeon/solod/so/io"

// File represents an open file descriptor.
type File struct {
	fd     *os_file
	closed bool
}

// Read reads up to len(b) bytes from the file and stores them in b.
// It returns the number of bytes read and any error encountered.
// At end of file, Read returns 0, io.EOF.
func (f *File) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	n := int(fread(&b[0], 1, uintptr(len(b)), f.fd))
	if n < len(b) {
		if ferror(f.fd) {
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
	if len(b) == 0 {
		return 0, nil
	}
	n := int(fwrite(&b[0], 1, uintptr(len(b)), f.fd))
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
	if fseeko(f.fd, offset, whence) != 0 {
		return 0, mapError()
	}
	pos := ftello(f.fd)
	if pos < 0 {
		return 0, mapError()
	}
	return pos, nil
}

// ReadAt reads len(b) bytes from the file starting at byte offset off.
// It returns the number of bytes read and the error, if any.
// ReadAt always returns a non-nil error when n < len(b).
func (f *File) ReadAt(b []byte, off int64) (int, error) {
	if off < 0 {
		return 0, io.ErrOffset
	}
	cur := ftello(f.fd)
	if cur < 0 {
		return 0, mapError()
	}
	if fseeko(f.fd, off, io.SeekStart) != 0 {
		return 0, mapError()
	}
	n, err := f.Read(b)
	if fseeko(f.fd, cur, io.SeekStart) != 0 && err == nil {
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
	if off < 0 {
		return 0, io.ErrOffset
	}
	cur := ftello(f.fd)
	if cur < 0 {
		return 0, mapError()
	}
	if fseeko(f.fd, off, io.SeekStart) != 0 {
		return 0, mapError()
	}
	n, err := f.Write(b)
	if fseeko(f.fd, cur, io.SeekStart) != 0 && err == nil {
		return n, mapError()
	}
	return n, err
}

// WriteString is like Write, but writes the contents of string s
// rather than a slice of bytes.
func (f *File) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}

// Close closes the file, rendering it unusable for I/O.
// Close will return an error if it has already been called.
func (f *File) Close() error {
	if f.closed {
		return ErrClosed
	}
	if fclose(f.fd) != 0 {
		return mapError()
	}
	f.closed = true
	return nil
}
