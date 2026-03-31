package os

//so:include <fcntl.h>
//so:include <sys/stat.h>
//so:include <unistd.h>

// MaxPathLen is the maximum length of a path.
const MaxPathLen = 4096

// MaxHostnameLen is the maximum length of a hostname.
const MaxHostnameLen = 255

//so:extern S_IFMT
const sIFMT = 0170000 // type of file mask
//so:extern S_IFIFO
const sIFIFO = 0010000 // named pipe (fifo)
//so:extern S_IFCHR
const sIFCHR = 0020000 // character special
//so:extern S_IFDIR
const sIFDIR = 0040000 // directory
//so:extern S_IFBLK
const sIFBLK = 0060000 // block special
//so:extern S_IFREG
const sIFREG = 0100000 // regular
//so:extern S_IFLNK
const sIFLNK = 0120000 // symbolic link
//so:extern S_IFSOCK
const sIFSOCK = 0140000 // socket
//so:extern S_ISUID
const sISUID = 0004000 // set user id on execution
//so:extern S_ISGID
const sISGID = 0002000 // set group id on execution
//so:extern S_ISVTX
const sISVTX = 0001000 // directory restricted delete

//so:extern UTIME_OMIT
const utimeOmit = 0 // sentinel: leave file time unchanged

//so:extern
type gid_t uint32 // group ID
//so:extern
type mode_t uint32 // file mode bits
//so:extern
type pid_t int32 // process ID
//so:extern
type uid_t uint32 // user ID

//so:extern
type os_statResult struct {
	size    int64
	mode    mode_t
	modSec  int64
	modNsec int64
	dev     uint64
	ino     uint64
	ok      bool
}

// int chdir(const char *path);
//
//so:extern
func chdir(path string) int {
	_ = path
	return 0
}

// int chmod(const char *path, mode_t mode);
//
//so:extern
func chmod(path string, mode mode_t) int {
	_, _ = path, mode
	return 0
}

// int chown(const char *path, uid_t owner, gid_t group);
//
//so:extern
func chown(path string, uid uid_t, gid gid_t) int {
	_, _, _ = path, uid, gid
	return 0
}

// int close(int fd);
//
//so:extern close
func fdclose(fd int) int {
	_ = fd
	return 0
}

// FILE *fdopen(int fd, const char *mode);
//
//so:extern
func fdopen(fd int, mode string) *os_file {
	_, _ = fd, mode
	return &os_file{}
}

// os_getcwd wraps getcwd with a null check to avoid
// glibc fortify-source nonnull warning when buf comes from a slice.
// char* getcwd(char *buf, size_t size);
//
//so:extern
func os_getcwd(buf *byte, size int) *byte {
	_, _ = buf, size
	return nil
}

// uid_t geteuid(void);
//
//so:extern
func geteuid() uid_t {
	return 0
}

// gid_t getegid(void);
//
//so:extern
func getegid() gid_t {
	return 0
}

// gid_t getgid(void);
//
//so:extern
func getgid() gid_t {
	return 0
}

// pid_t getpid(void);
//
//so:extern
func getpid() pid_t {
	return 0
}

// pid_t getppid(void);
//
//so:extern
func getppid() pid_t {
	return 0
}

// uid_t getuid(void);
//
//so:extern
func getuid() uid_t {
	return 0
}

// os_gethostname wraps gethostname with a null check to avoid
// glibc fortify-source nonnull warning when buf comes from a slice.
// int gethostname(char* name, size_t namelen);
//
//so:extern
func os_gethostname(buf *byte, size int) int {
	_, _ = buf, size
	return 0
}

// int lchown(const char *path, uid_t owner, gid_t group);
//
//so:extern
func lchown(path string, uid uid_t, gid gid_t) int {
	_, _, _ = path, uid, gid
	return 0
}

// int link(const char *path1, const char *path2);
//
//so:extern
func link(old, new string) int {
	_, _ = old, new
	return 0
}

// lstat wrapper (fills os_statResult).
// int lstat(const char* restrict path, struct stat* restrict buf);
//
//so:extern
func os_lstat(path string) os_statResult {
	_ = path
	return os_statResult{size: 42, mode: 0o777, ok: true}
}

// int mkdir(const char *path, mode_t mode);
//
//so:extern
func mkdir(path string, mode mode_t) int {
	_, _ = path, mode
	return 0
}

// int mkstemp(char *template);
//
//so:extern
func mkstemp(tmpl *byte) int {
	_ = tmpl
	return 0
}

// char* mkdtemp(char *template);
//
//so:extern
func mkdtemp(tmpl *byte) any {
	_ = tmpl
	b := []byte("example/tmpdir")
	return &b[0]
}

// int open(const char *path, int oflag, ...);
//
//so:extern open
func posixOpen(path string, flags int, mode uint32) int {
	_, _, _ = path, flags, mode
	return 42
}

// os_readlink wraps readlink with a null check to avoid
// glibc fortify-source nonnull warning when buf comes from a slice.
// ssize_t readlink(const char* restrict path, char* restrict buf, size_t bufsize);
//
//so:extern
func os_readlink(path string, buf *byte, bufsiz int) int {
	_, _, _ = path, buf, bufsiz
	return 0
}

// stat wrapper (fills os_statResult).
// int stat(const char* restrict path, struct stat* restrict buf);
//
//so:extern
func os_stat(path string) os_statResult {
	_ = path
	return os_statResult{}
}

// int symlink(const char *path1, const char *path2);
//
//so:extern
func symlink(old, new string) int {
	_, _ = old, new
	return 0
}

// int truncate(const char *path, off_t length);
//
//so:extern
func truncate(path string, size int64) int {
	_, _ = path, size
	return 0
}

// int setenv(const char *name, const char *value, int overwrite);
//
//so:extern
func setenv(key, value string, overwrite int) int {
	_, _, _ = key, value, overwrite
	return 0
}

// int unsetenv(const char *name);
//
//so:extern
func unsetenv(key string) int {
	_ = key
	return 0
}

// utimensat wrapper (passes timespec values as separate arguments).
//
//so:extern
func os_utimens(path string, asec, ansec, msec, mnsec int64) int {
	_, _, _, _, _ = path, asec, ansec, msec, mnsec
	return 0
}
