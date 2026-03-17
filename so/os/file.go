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
	n := fread(&b[0], 1, len(b), f.fd)
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
	n := fwrite(&b[0], 1, len(b), f.fd)
	if n < len(b) {
		return n, mapError()
	}
	return n, nil
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
