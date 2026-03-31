package os

//so:include <errno.h>

//so:embed os.h
var os_h string

//so:embed os.c
var os_c string

//so:extern
var errno int

// Errno constants.
//
//so:extern EACCES
const eACCES = 0 // Permission denied
//so:extern EEXIST
const eEXIST = 0 // File exists
//so:extern EISDIR
const eISDIR = 0 // Is a directory
//so:extern ENOENT
const eNOENT = 0 // No such file or directory
//so:extern ENOTDIR
const eNOTDIR = 0 // Not a directory
//so:extern EPERM
const ePERM = 0 // Operation not permitted

//so:extern FILE
type os_file struct{}

//so:extern
var (
	stdin  *os_file
	stdout *os_file
	stderr *os_file
)

// FILE *fopen(const char *restrict filename, const char *restrict mode);
//
//so:extern
func fopen(path string, mode string) *os_file {
	_, _ = path, mode
	return &os_file{}
}

// int fclose( FILE *stream );
//
//so:extern
func fclose(stream *os_file) int {
	_ = stream
	return 0
}

// size_t fread(void *restrict buffer, size_t size, size_t count, FILE *restrict stream);
//
//so:extern
func fread(ptr any, size uintptr, count uintptr, stream *os_file) uintptr {
	_, _, _, _ = ptr, size, count, stream
	return 0
}

// size_t fwrite(const void* restrict buffer, size_t size, size_t count, FILE* restrict stream);
//
//so:extern
func fwrite(ptr any, size uintptr, count uintptr, stream *os_file) uintptr {
	_, _, _, _ = ptr, size, count, stream
	return count
}

// int ferror(FILE *stream);
//
//so:extern
func ferror(stream *os_file) bool {
	_ = stream
	return false
}

// int fseeko(FILE *stream, off_t offset, int whence);
//
//so:extern
func fseeko(stream *os_file, offset int64, whence int) int {
	_, _, _ = stream, offset, whence
	return 0
}

// off_t ftello(FILE *stream);
//
//so:extern
func ftello(stream *os_file) int64 {
	_ = stream
	return 0
}

// int remove(const char* fname);
//
//so:extern
func remove(path string) int {
	_ = path
	return 0
}

// int rename(const char* old_filename, const char* new_filename);
//
//so:extern
func rename(oldpath string, newpath string) int {
	_, _ = oldpath, newpath
	return 0
}

// char *getenv(const char *name);
//
//so:extern
func getenv(name string) any {
	_ = name
	b := []byte("getenv:" + name)
	return &b[0]
}

// void exit(int exit_code);
//
//so:extern
func exit(status int) {
	_ = status
}
