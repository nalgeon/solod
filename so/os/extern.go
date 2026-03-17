package os

//so:include <errno.h>
//so:include <stdio.h>

//so:embed os.h
var os_h string

//so:extern
var errno int

// Errno constants mapped to C macros in os.h.
//
//so:extern
const (
	os_EACCES = iota
	os_EEXIST
	os_EISDIR
	os_ENOENT
	os_ENOTDIR
)

//so:extern
type os_file struct{}

//so:extern
func fopen(path string, mode string) *os_file {
	_, _ = path, mode
	return nil
}

//so:extern
func fclose(stream *os_file) int {
	_ = stream
	return 0
}

//so:extern
func fread(ptr *byte, size int, count int, stream *os_file) int {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

//so:extern
func fwrite(ptr *byte, size int, count int, stream *os_file) int {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

//so:extern
func ferror(stream *os_file) bool {
	_ = stream
	return false
}

//so:extern
func remove(path string) int {
	_ = path
	return 0
}

//so:extern
func rename(oldpath string, newpath string) int {
	_, _ = oldpath, newpath
	return 0
}
