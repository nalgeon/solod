#include "main.h"

// -- Forward declarations --
static void basicTest(void);
static void envTest(void);
static void fileTest(void);
static void procTest(void);
static void seekTest(void);
static void statTest(void);
static void tempTest(void);

// -- basic.go --

static void basicTest(void) {
    {
        // WriteFile, ReadFile.
        so_String name = so_str("test_rw.txt");
        so_Slice data = so_string_bytes(so_str("hello world"));
        so_Error err = os_WriteFile(name, data, 0666);
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        so_R_slice_err _res1 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res1.val;
        err = _res1.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadFile failed");
        }
        if (so_string_ne(so_bytes_string(b), so_bytes_string(data))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(name);
            so_panic("ReadFile: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
    {
        // Create, Write, Close.
        so_String name = so_str("test_file.txt");
        os_FileResult _res2 = os_Create(name);
        os_File f = _res2.val;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        // Write.
        so_R_int_err _res3 = os_File_Write(&f, so_string_bytes(so_str("abcdef")));
        so_int n = _res3.val;
        err = _res3.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Write failed");
        }
        if (n != 6) {
            os_Remove(name);
            so_panic("Write: wrong count");
        }
        // Close.
        err = os_File_Close(&f);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Close failed");
        }
        os_Remove(name);
    }
    {
        // Open, Read, Close.
        so_String name = so_str("test_file.txt");
        so_Slice data = so_string_bytes(so_str("abcdef"));
        so_Error err = os_WriteFile(name, data, 0666);
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        // Open.
        os_FileResult _res4 = os_Open(name);
        os_File f = _res4.val;
        err = _res4.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Open failed");
        }
        // Read.
        so_Slice buf = so_make_slice(so_byte, 10, 10);
        so_R_int_err _res5 = os_File_Read(&f, buf);
        so_int n = _res5.val;
        err = _res5.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Read failed");
        }
        if (n != 6) {
            os_Remove(name);
            so_panic("Read: wrong count");
        }
        if (so_string_ne(so_bytes_string(so_slice(so_byte, buf, 0, n)), so_str("abcdef"))) {
            os_Remove(name);
            so_panic("Read: wrong data");
        }
        // Close.
        err = os_File_Close(&f);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Close failed");
        }
        os_Remove(name);
    }
    {
        // WriteString.
        so_String name = so_str("test_writestr.txt");
        os_FileResult _res6 = os_Create(name);
        os_File f = _res6.val;
        so_Error err = _res6.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        so_R_int_err _res7 = os_File_WriteString(&f, so_str("hello"));
        so_int n = _res7.val;
        err = _res7.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("WriteString failed");
        }
        if (n != 5) {
            os_Remove(name);
            so_panic("WriteString: wrong count");
        }
        os_File_Close(&f);
        so_R_slice_err _res8 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res8.val;
        err = _res8.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadFile failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("hello"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(name);
            so_panic("WriteString: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
    {
        // Stdout, Stderr.
        so_R_int_err _res9 = os_File_WriteString(&os_Stdout, so_str("hello"));
        so_int n = _res9.val;
        so_Error err = _res9.err;
        if (err != NULL) {
            so_panic("Stdout failed");
        }
        if (n != 5) {
            so_panic("Stdout: wrong count");
        }
        so_R_int_err _res10 = os_File_WriteString(&os_Stderr, so_str("goodbye"));
        n = _res10.val;
        err = _res10.err;
        if (err != NULL) {
            so_panic("Stderr failed");
        }
        if (n != 7) {
            so_panic("Stderr: wrong count");
        }
        so_println("");
    }
}

// -- env.go --

static void envTest(void) {
    {
        // Setenv, Getenv.
        so_Error err = os_Setenv(so_str("SO_TEST_KEY"), so_str("test_value"));
        if (err != NULL) {
            so_panic("Setenv failed");
        }
        so_String val = os_Getenv(so_str("SO_TEST_KEY"));
        if (so_string_ne(val, so_str("test_value"))) {
            so_panic("Getenv: wrong value");
        }
    }
    {
        // LookupEnv - present.
        os_Setenv(so_str("SO_TEST_LOOKUP"), so_str("found"));
        so_R_str_bool _res1 = os_LookupEnv(so_str("SO_TEST_LOOKUP"));
        so_String val = _res1.val;
        bool ok = _res1.val2;
        if (!ok) {
            so_panic("LookupEnv: should be present");
        }
        if (so_string_ne(val, so_str("found"))) {
            so_panic("LookupEnv: wrong value");
        }
    }
    {
        // LookupEnv - absent.
        so_R_str_bool _res2 = os_LookupEnv(so_str("SO_TEST_NONEXISTENT_VAR_XYZ"));
        bool ok = _res2.val2;
        if (ok) {
            so_panic("LookupEnv: should not be present");
        }
    }
    {
        // Unsetenv.
        os_Setenv(so_str("SO_TEST_UNSET"), so_str("bye"));
        so_Error err = os_Unsetenv(so_str("SO_TEST_UNSET"));
        if (err != NULL) {
            so_panic("Unsetenv failed");
        }
        so_String val = os_Getenv(so_str("SO_TEST_UNSET"));
        if (so_string_ne(val, so_str(""))) {
            so_panic("Unsetenv: should be empty");
        }
    }
    {
        // Getenv on PATH (should always be set).
        so_String path = os_Getenv(so_str("PATH"));
        if (so_len(path) == 0) {
            so_panic("Getenv PATH: empty");
        }
    }
}

// -- file.go --

static void fileTest(void) {
    {
        // OpenFile with O_CREATE | O_WRONLY | O_TRUNC.
        so_String name = so_str("test_openfile.txt");
        os_FileResult _res1 = os_OpenFile(name, ((O_CREAT | O_WRONLY) | O_TRUNC), 0644);
        os_File f = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("OpenFile create failed");
        }
        os_File_Write(&f, so_string_bytes(so_str("openfile")));
        os_File_Close(&f);
        so_R_slice_err _res2 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res2.val;
        err = _res2.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadFile after OpenFile failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("openfile"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(name);
            so_panic("OpenFile: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
    {
        // OpenFile with O_RDONLY.
        so_String name = so_str("test_openfile_rd.txt");
        os_WriteFile(name, so_string_bytes(so_str("readonly")), 0666);
        os_FileResult _res3 = os_OpenFile(name, O_RDONLY, 0);
        os_File f = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("OpenFile rdonly failed");
        }
        so_Slice buf = so_make_slice(so_byte, 16, 16);
        so_R_int_err _res4 = os_File_Read(&f, buf);
        so_int n = _res4.val;
        err = _res4.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Read from rdonly failed");
        }
        if (so_string_ne(so_bytes_string(so_slice(so_byte, buf, 0, n)), so_str("readonly"))) {
            os_Remove(name);
            so_panic("OpenFile rdonly: wrong data");
        }
        os_File_Close(&f);
        os_Remove(name);
    }
    {
        // File.Name.
        so_String name = so_str("test_filename.txt");
        os_FileResult _res5 = os_Create(name);
        os_File f = _res5.val;
        so_Error err = _res5.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        if (so_string_ne(os_File_Name(&f), name)) {
            os_Remove(name);
            so_panic("Name: wrong");
        }
        os_File_Close(&f);
        os_Remove(name);
    }
    {
        // Link and Readlink (via symlink).
        so_String target = so_str("test_link_target.txt");
        os_WriteFile(target, so_string_bytes(so_str("linked")), 0666);
        // Hard link.
        so_String hard = so_str("test_hard_link.txt");
        so_Error err = os_Link(target, hard);
        if (err != NULL) {
            os_Remove(target);
            so_panic("Link failed");
        }
        so_R_slice_err _res6 = os_ReadFile((mem_Allocator){0}, hard);
        so_Slice b = _res6.val;
        err = _res6.err;
        if (err != NULL) {
            os_Remove(hard);
            os_Remove(target);
            so_panic("ReadFile hard link failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("linked"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(hard);
            os_Remove(target);
            so_panic("Hard link: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(hard);
        os_Remove(target);
    }
    {
        // Symlink and Readlink.
        so_String target = so_str("test_sym_target.txt");
        os_WriteFile(target, so_string_bytes(so_str("sym")), 0666);
        so_String link = so_str("test_sym_link");
        so_Error err = os_Symlink(target, link);
        if (err != NULL) {
            os_Remove(target);
            so_panic("Symlink failed");
        }
        so_byte rlBuf[4096] = {0};
        so_R_str_err _res7 = os_Readlink(so_array_slice(so_byte, rlBuf, 0, 4096, 4096), link);
        so_String dest = _res7.val;
        err = _res7.err;
        if (err != NULL) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Readlink failed");
        }
        if (so_string_ne(dest, target)) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Readlink: wrong target");
        }
        os_Remove(link);
        os_Remove(target);
    }
    {
        // Mkdir and Chdir.
        so_String dir = so_str("test_mkdir_dir");
        so_Error err = os_Mkdir(dir, 0755);
        if (err != NULL) {
            so_panic("Mkdir failed");
        }
        // Get current dir.
        so_byte wdBuf[4096] = {0};
        so_R_str_err _res8 = os_Getwd(so_array_slice(so_byte, wdBuf, 0, 4096, 4096));
        so_String origWd = _res8.val;
        err = _res8.err;
        if (err != NULL) {
            os_Remove(dir);
            so_panic("Getwd failed");
        }
        // Change to new dir.
        err = os_Chdir(dir);
        if (err != NULL) {
            os_Remove(dir);
            so_panic("Chdir failed");
        }
        // Verify we moved.
        so_byte wdBuf2[4096] = {0};
        so_R_str_err _res9 = os_Getwd(so_array_slice(so_byte, wdBuf2, 0, 4096, 4096));
        so_String newWd = _res9.val;
        err = _res9.err;
        if (err != NULL) {
            os_Remove(dir);
            so_panic("Getwd after Chdir failed");
        }
        if (so_string_eq(newWd, origWd)) {
            os_Remove(dir);
            so_panic("Chdir: dir did not change");
        }
        // Change back.
        os_Chdir(origWd);
        os_Remove(dir);
    }
    {
        // Truncate.
        so_String name = so_str("test_truncate.txt");
        os_WriteFile(name, so_string_bytes(so_str("abcdef")), 0666);
        so_Error err = os_Truncate(name, 3);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Truncate failed");
        }
        so_R_slice_err _res10 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res10.val;
        err = _res10.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadFile after Truncate failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("abc"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(name);
            so_panic("Truncate: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
    {
        // OpenFile with O_APPEND.
        so_String name = so_str("test_append.txt");
        os_WriteFile(name, so_string_bytes(so_str("hello")), 0666);
        os_FileResult _res11 = os_OpenFile(name, (O_WRONLY | O_APPEND), 0);
        os_File f = _res11.val;
        so_Error err = _res11.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("OpenFile append failed");
        }
        os_File_Write(&f, so_string_bytes(so_str(" world")));
        os_File_Close(&f);
        so_R_slice_err _res12 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res12.val;
        err = _res12.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadFile after append failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("hello world"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(name);
            so_panic("Append: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
    {
        // Chtimes - just verify it doesn't error.
        so_String name = so_str("test_chtimes.txt");
        os_WriteFile(name, so_string_bytes(so_str("times")), 0666);
        os_FileInfoResult _res13 = os_Stat(name);
        os_FileInfo fi = _res13.val;
        so_Error err = _res13.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Stat for Chtimes failed");
        }
        time_Time mt = os_FileInfo_ModTime(&fi);
        err = os_Chtimes(name, mt, mt);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Chtimes failed");
        }
        os_Remove(name);
    }
    {
        // Chown with -1, -1 (no change) - should succeed.
        so_String name = so_str("test_chown.txt");
        os_WriteFile(name, so_string_bytes(so_str("chown")), 0666);
        so_Error err = os_Chown(name, -1, -1);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Chown failed");
        }
        os_Remove(name);
    }
    {
        // Lchown with -1, -1 (no change) - should succeed.
        so_String name = so_str("test_lchown.txt");
        os_WriteFile(name, so_string_bytes(so_str("lchown")), 0666);
        so_Error err = os_Lchown(name, -1, -1);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Lchown failed");
        }
        os_Remove(name);
    }
    {
        // Remove.
        so_String name = so_str("test_remove.txt");
        so_Error err = os_WriteFile(name, so_string_bytes(so_str("tmp")), 0666);
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        err = os_Remove(name);
        if (err != NULL) {
            so_panic("Remove failed");
        }
        os_FileResult _res14 = os_Open(name);
        err = _res14.err;
        if (err == NULL) {
            so_panic("Open after Remove should fail");
        }
    }
    {
        // Rename.
        so_String oldName = so_str("test_old.txt");
        so_String newName = so_str("test_new.txt");
        os_WriteFile(oldName, so_string_bytes(so_str("renamed")), 0666);
        so_Error err = os_Rename(oldName, newName);
        if (err != NULL) {
            so_panic("Rename failed");
        }
        so_R_slice_err _res15 = os_ReadFile((mem_Allocator){0}, newName);
        so_Slice b = _res15.val;
        err = _res15.err;
        if (err != NULL) {
            os_Remove(newName);
            so_panic("ReadFile after Rename failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("renamed"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(newName);
            so_panic("Rename: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(newName);
    }
    {
        // ErrExist - try to create dir that already exists.
        so_String name = so_str("test_exist_dir");
        os_Mkdir(name, 0755);
        so_Error err = os_Mkdir(name, 0755);
        if (err != os_ErrExist) {
            os_Remove(name);
            so_panic("Mkdir existing: wrong error");
        }
        os_Remove(name);
    }
    {
        // ErrNotExist.
        os_FileResult _res16 = os_Open(so_str("nonexistent_file.txt"));
        so_Error err = _res16.err;
        if (err != os_ErrNotExist) {
            so_panic("Open nonexistent: wrong error");
        }
    }
    {
        // Verify OpenFile nonexistent returns ErrNotExist.
        os_FileResult _res17 = os_OpenFile(so_str("nonexistent_open.txt"), O_RDONLY, 0);
        so_Error err = _res17.err;
        if (err != os_ErrNotExist) {
            so_panic("OpenFile nonexistent: wrong error");
        }
    }
}

// -- main.go --

int main(void) {
    basicTest();
    envTest();
    fileTest();
    procTest();
    seekTest();
    statTest();
    tempTest();
}

// -- proc.go --

static void procTest(void) {
    {
        // Getpid.
        so_int pid = os_Getpid();
        if (pid <= 0) {
            so_panic("Getpid: invalid");
        }
    }
    {
        // Getppid.
        so_int ppid = os_Getppid();
        if (ppid < 0) {
            so_panic("Getppid: invalid");
        }
    }
    {
        // Getuid.
        so_int uid = os_Getuid();
        if (uid < 0) {
            so_panic("Getuid: invalid");
        }
    }
    {
        // Geteuid.
        so_int euid = os_Geteuid();
        if (euid < 0) {
            so_panic("Geteuid: invalid");
        }
    }
    {
        // Getgid.
        so_int gid = os_Getgid();
        if (gid < 0) {
            so_panic("Getgid: invalid");
        }
    }
    {
        // Getegid.
        so_int egid = os_Getegid();
        if (egid < 0) {
            so_panic("Getegid: invalid");
        }
    }
    {
        // Getwd.
        so_byte wdBuf[4096] = {0};
        so_R_str_err _res1 = os_Getwd(so_array_slice(so_byte, wdBuf, 0, 4096, 4096));
        so_String wd = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("Getwd failed");
        }
        if (so_len(wd) == 0) {
            so_panic("Getwd: empty");
        }
        // Should start with '/'.
        if (so_at(so_byte, wd, 0) != '/') {
            so_panic("Getwd: not absolute");
        }
    }
    {
        // Hostname.
        so_byte hostBuf[255] = {0};
        so_R_str_err _res2 = os_Hostname(so_array_slice(so_byte, hostBuf, 0, 255, 255));
        so_String name = _res2.val;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic("Hostname failed");
        }
        if (so_len(name) == 0) {
            so_panic("Hostname: empty");
        }
    }
}

// -- seek.go --

static void seekTest(void) {
    {
        // Seek.
        so_String name = so_str("test_seek.txt");
        os_FileResult _res1 = os_Create(name);
        os_File f = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        os_File_Write(&f, so_string_bytes(so_str("abcdef")));
        so_R_i64_err _res2 = os_File_Seek(&f, 0, io_SeekStart);
        int64_t pos = _res2.val;
        err = _res2.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Seek failed");
        }
        if (pos != 0) {
            os_Remove(name);
            so_panic("Seek: wrong position");
        }
        so_Slice buf = so_make_slice(so_byte, 6, 6);
        so_R_int_err _res3 = os_File_Read(&f, buf);
        so_int n = _res3.val;
        err = _res3.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Read after Seek failed");
        }
        if (so_string_ne(so_bytes_string(so_slice(so_byte, buf, 0, n)), so_str("abcdef"))) {
            os_Remove(name);
            so_panic("Seek: wrong data");
        }
        os_File_Close(&f);
        os_Remove(name);
    }
    {
        // ReadAt.
        so_String name = so_str("test_readat.txt");
        so_Error err = os_WriteFile(name, so_string_bytes(so_str("hello world")), 0666);
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        os_FileResult _res4 = os_Open(name);
        os_File f = _res4.val;
        err = _res4.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Open failed");
        }
        so_Slice buf = so_make_slice(so_byte, 5, 5);
        so_R_int_err _res5 = os_File_ReadAt(&f, buf, 6);
        so_int n = _res5.val;
        err = _res5.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadAt failed");
        }
        if (n != 5) {
            os_Remove(name);
            so_panic("ReadAt: wrong count");
        }
        if (so_string_ne(so_bytes_string(so_slice(so_byte, buf, 0, n)), so_str("world"))) {
            os_Remove(name);
            so_panic("ReadAt: wrong data");
        }
        os_File_Close(&f);
        os_Remove(name);
    }
    {
        // WriteAt.
        so_String name = so_str("test_writeat.txt");
        os_FileResult _res6 = os_Create(name);
        os_File f = _res6.val;
        so_Error err = _res6.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        os_File_Write(&f, so_string_bytes(so_str("hello world")));
        so_R_int_err _res7 = os_File_WriteAt(&f, so_string_bytes(so_str("WORLD")), 6);
        err = _res7.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("WriteAt failed");
        }
        os_File_Close(&f);
        so_R_slice_err _res8 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res8.val;
        err = _res8.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("ReadFile failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("hello WORLD"))) {
            mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
            os_Remove(name);
            so_panic("WriteAt: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
}

// -- stat.go --

static void statTest(void) {
    {
        // Stat on a regular file.
        so_String name = so_str("test_stat.txt");
        os_WriteFile(name, so_string_bytes(so_str("hello")), 0666);
        os_FileInfoResult _res1 = os_Stat(name);
        os_FileInfo fi = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Stat failed");
        }
        if (so_string_ne(os_FileInfo_Name(&fi), so_str("test_stat.txt"))) {
            os_Remove(name);
            so_panic("Stat: wrong name");
        }
        if (os_FileInfo_Size(&fi) != 5) {
            os_Remove(name);
            so_panic("Stat: wrong size");
        }
        if (!os_FileMode_IsRegular(os_FileInfo_Mode(&fi))) {
            os_Remove(name);
            so_panic("Stat: not regular");
        }
        if (os_FileInfo_IsDir(&fi)) {
            os_Remove(name);
            so_panic("Stat: should not be dir");
        }
        os_Remove(name);
    }
    {
        // Stat on a directory.
        so_String name = so_str("test_stat_dir");
        os_Mkdir(name, 0755);
        os_FileInfoResult _res2 = os_Stat(name);
        os_FileInfo fi = _res2.val;
        so_Error err = _res2.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Stat dir failed");
        }
        if (so_string_ne(os_FileInfo_Name(&fi), so_str("test_stat_dir"))) {
            os_Remove(name);
            so_panic("Stat dir: wrong name");
        }
        if (!os_FileInfo_IsDir(&fi)) {
            os_Remove(name);
            so_panic("Stat dir: should be dir");
        }
        if (os_FileMode_IsRegular(os_FileInfo_Mode(&fi))) {
            os_Remove(name);
            so_panic("Stat dir: should not be regular");
        }
        os_Remove(name);
    }
    {
        // Lstat on a symlink.
        so_String target = so_str("test_lstat_target.txt");
        so_String link = so_str("test_lstat_link");
        os_WriteFile(target, so_string_bytes(so_str("target")), 0666);
        os_Symlink(target, link);
        // Lstat returns info about the link itself.
        os_FileInfoResult _res3 = os_Lstat(link);
        os_FileInfo fi = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Lstat failed");
        }
        if (so_string_ne(os_FileInfo_Name(&fi), so_str("test_lstat_link"))) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Lstat: wrong name");
        }
        if ((os_FileInfo_Mode(&fi) & os_ModeSymlink) == 0) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Lstat: should be symlink");
        }
        // Stat follows the link.
        os_FileInfoResult _res4 = os_Stat(link);
        os_FileInfo fi2 = _res4.val;
        err = _res4.err;
        if (err != NULL) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Stat through link failed");
        }
        if (os_FileInfo_Size(&fi2) != 6) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Stat through link: wrong size");
        }
        if ((os_FileInfo_Mode(&fi2) & os_ModeSymlink) != 0) {
            os_Remove(link);
            os_Remove(target);
            so_panic("Stat through link: should not be symlink");
        }
        os_Remove(link);
        os_Remove(target);
    }
    {
        // SameFile.
        so_String name = so_str("test_samefile.txt");
        os_WriteFile(name, so_string_bytes(so_str("same")), 0666);
        os_FileInfoResult _res5 = os_Stat(name);
        os_FileInfo fi1 = _res5.val;
        so_Error err = _res5.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Stat 1 failed");
        }
        os_FileInfoResult _res6 = os_Stat(name);
        os_FileInfo fi2 = _res6.val;
        err = _res6.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Stat 2 failed");
        }
        if (!os_SameFile(fi1, fi2)) {
            os_Remove(name);
            so_panic("SameFile: should be same");
        }
        so_String name2 = so_str("test_samefile2.txt");
        os_WriteFile(name2, so_string_bytes(so_str("other")), 0666);
        os_FileInfoResult _res7 = os_Stat(name2);
        os_FileInfo fi3 = _res7.val;
        err = _res7.err;
        if (err != NULL) {
            os_Remove(name2);
            os_Remove(name);
            so_panic("Stat 3 failed");
        }
        if (os_SameFile(fi1, fi3)) {
            os_Remove(name2);
            os_Remove(name);
            so_panic("SameFile: should be different");
        }
        os_Remove(name2);
        os_Remove(name);
    }
    {
        // Stat on nonexistent file.
        os_FileInfoResult _res8 = os_Stat(so_str("nonexistent_stat.txt"));
        so_Error err = _res8.err;
        if (err != os_ErrNotExist) {
            so_panic("Stat nonexistent: wrong error");
        }
    }
    {
        // Chmod and permission check.
        so_String name = so_str("test_chmod.txt");
        os_WriteFile(name, so_string_bytes(so_str("chmod")), 0666);
        so_Error err = os_Chmod(name, 0644);
        if (err != NULL) {
            os_Remove(name);
            so_panic("Chmod failed");
        }
        os_FileInfoResult _res9 = os_Stat(name);
        os_FileInfo fi = _res9.val;
        err = _res9.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Stat after Chmod failed");
        }
        if (os_FileMode_Perm(os_FileInfo_Mode(&fi)) != 0644) {
            os_Remove(name);
            so_panic("Chmod: wrong perm");
        }
        os_Remove(name);
    }
}

// -- temp.go --

static void tempTest(void) {
    so_Slice buf = so_make_slice(so_byte, os_MaxPathLen, os_MaxPathLen);
    {
        // TempDir.
        so_String td = os_TempDir();
        if (so_len(td) == 0) {
            so_panic("TempDir: empty");
        }
    }
    {
        // CreateTemp.
        os_FileResult _res1 = os_CreateTemp(buf, so_str(""), so_str("sotest"));
        os_File f = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("CreateTemp failed");
        }
        so_String name = os_File_Name(&f);
        if (so_len(name) == 0) {
            so_panic("CreateTemp: empty name");
        }
        // Name should contain the pattern prefix.
        if (!strings_Contains(name, so_str("sotest"))) {
            so_panic("CreateTemp: name missing pattern");
        }
        os_File_Write(&f, so_string_bytes(so_str("temp data")));
        os_File_Close(&f);
        // Verify the file exists.
        so_R_slice_err _res2 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res2.val;
        err = _res2.err;
        if (err != NULL) {
            so_panic("ReadFile temp failed");
        }
        if (so_string_ne(so_bytes_string(b), so_str("temp data"))) {
            so_panic("CreateTemp: wrong data");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
        os_Remove(name);
    }
    {
        // CreateTemp with specific dir.
        so_String td = os_TempDir();
        os_FileResult _res3 = os_CreateTemp(buf, td, so_str("myprefix"));
        os_File f = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            so_panic("CreateTemp dir failed");
        }
        so_String name = os_File_Name(&f);
        if (!strings_Contains(name, so_str("myprefix"))) {
            so_panic("CreateTemp dir: missing pattern");
        }
        if (!strings_HasPrefix(name, td)) {
            so_panic("CreateTemp dir: wrong dir");
        }
        os_File_Close(&f);
        os_Remove(name);
    }
    {
        // MkdirTemp.
        so_R_str_err _res4 = os_MkdirTemp(buf, so_str(""), so_str("sotest"));
        so_String dir = _res4.val;
        so_Error err = _res4.err;
        if (err != NULL) {
            so_panic("MkdirTemp failed");
        }
        if (so_len(dir) == 0) {
            so_panic("MkdirTemp: empty");
        }
        if (!strings_Contains(dir, so_str("sotest"))) {
            so_panic("MkdirTemp: name missing pattern");
        }
        // Verify it's a directory.
        os_FileInfoResult _res5 = os_Stat(dir);
        os_FileInfo fi = _res5.val;
        err = _res5.err;
        if (err != NULL) {
            so_panic("Stat MkdirTemp failed");
        }
        if (!os_FileInfo_IsDir(&fi)) {
            so_panic("MkdirTemp: not a directory");
        }
        os_Remove(dir);
    }
}
