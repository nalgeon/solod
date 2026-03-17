#include "main.h"

// -- Implementation --

int main(void) {
    {
        // WriteFile, ReadFile.
        so_String name = so_str("test_rw.txt");
        so_Slice data = so_string_bytes(so_str("hello world"));
        so_Error err = os_WriteFile(name, data);
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        so_Result _res1 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res1.val.as_slice;
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
        so_Result _res3 = os_File_Write(&f, so_string_bytes(so_str("abcdef")));
        so_int n = _res3.val.as_int;
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
        so_Error err = os_WriteFile(name, data);
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
        so_Result _res5 = os_File_Read(&f, buf);
        so_int n = _res5.val.as_int;
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
        // Remove.
        so_String name = so_str("test_remove.txt");
        so_Error err = os_WriteFile(name, so_string_bytes(so_str("tmp")));
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        err = os_Remove(name);
        if (err != NULL) {
            so_panic("Remove failed");
        }
        os_FileResult _res6 = os_Open(name);
        err = _res6.err;
        if (err == NULL) {
            so_panic("Open after Remove should fail");
        }
    }
    {
        // Rename.
        so_String oldName = so_str("test_old.txt");
        so_String newName = so_str("test_new.txt");
        os_WriteFile(oldName, so_string_bytes(so_str("renamed")));
        so_Error err = os_Rename(oldName, newName);
        if (err != NULL) {
            so_panic("Rename failed");
        }
        so_Result _res7 = os_ReadFile((mem_Allocator){0}, newName);
        so_Slice b = _res7.val.as_slice;
        err = _res7.err;
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
        // ErrNotExist.
        os_FileResult _res8 = os_Open(so_str("nonexistent_file.txt"));
        so_Error err = _res8.err;
        if (err != os_ErrNotExist) {
            so_panic("Open nonexistent: wrong error");
        }
    }
    {
        // Seek.
        so_String name = so_str("test_seek.txt");
        os_FileResult _res9 = os_Create(name);
        os_File f = _res9.val;
        so_Error err = _res9.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        os_File_Write(&f, so_string_bytes(so_str("abcdef")));
        so_Result _res10 = os_File_Seek(&f, 0, io_SeekStart);
        int64_t pos = _res10.val.as_i64;
        err = _res10.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Seek failed");
        }
        if (pos != 0) {
            os_Remove(name);
            so_panic("Seek: wrong position");
        }
        so_Slice buf = so_make_slice(so_byte, 6, 6);
        so_Result _res11 = os_File_Read(&f, buf);
        so_int n = _res11.val.as_int;
        err = _res11.err;
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
        so_Error err = os_WriteFile(name, so_string_bytes(so_str("hello world")));
        if (err != NULL) {
            so_panic("WriteFile failed");
        }
        os_FileResult _res12 = os_Open(name);
        os_File f = _res12.val;
        err = _res12.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("Open failed");
        }
        so_Slice buf = so_make_slice(so_byte, 5, 5);
        so_Result _res13 = os_File_ReadAt(&f, buf, 6);
        so_int n = _res13.val.as_int;
        err = _res13.err;
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
        os_FileResult _res14 = os_Create(name);
        os_File f = _res14.val;
        so_Error err = _res14.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        os_File_Write(&f, so_string_bytes(so_str("hello world")));
        so_Result _res15 = os_File_WriteAt(&f, so_string_bytes(so_str("WORLD")), 6);
        err = _res15.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("WriteAt failed");
        }
        os_File_Close(&f);
        so_Result _res16 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res16.val.as_slice;
        err = _res16.err;
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
    {
        // WriteString.
        so_String name = so_str("test_writestr.txt");
        os_FileResult _res17 = os_Create(name);
        os_File f = _res17.val;
        so_Error err = _res17.err;
        if (err != NULL) {
            so_panic("Create failed");
        }
        so_Result _res18 = os_File_WriteString(&f, so_str("hello"));
        so_int n = _res18.val.as_int;
        err = _res18.err;
        if (err != NULL) {
            os_Remove(name);
            so_panic("WriteString failed");
        }
        if (n != 5) {
            os_Remove(name);
            so_panic("WriteString: wrong count");
        }
        os_File_Close(&f);
        so_Result _res19 = os_ReadFile((mem_Allocator){0}, name);
        so_Slice b = _res19.val.as_slice;
        err = _res19.err;
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
        // Getenv.
        so_String path = os_Getenv(so_str("PATH"));
        if (so_len(path) == 0) {
            so_panic("Getenv PATH: empty");
        }
    }
}
