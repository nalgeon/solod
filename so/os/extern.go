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
func fread(ptr any, size uintptr, count uintptr, stream *os_file) uintptr {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

//so:extern
func fwrite(ptr any, size uintptr, count uintptr, stream *os_file) uintptr {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

//so:extern
func ferror(stream *os_file) bool {
	_ = stream
	return false
}

//so:extern
func fseeko(stream *os_file, offset int64, whence int) int {
	_, _, _ = stream, offset, whence
	return 0
}

//so:extern
func ftello(stream *os_file) int64 {
	_ = stream
	return 0
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

//so:extern
func getenv(name string) any {
	_ = name
	return nil
}

//so:extern
func exit(status int) {
	_ = status
}
