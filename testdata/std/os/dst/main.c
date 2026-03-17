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
}
