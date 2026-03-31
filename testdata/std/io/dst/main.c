#include "main.h"

// -- Types --

typedef struct reader {
    so_Slice b;
} reader;

typedef struct writer {
    so_Slice b;
} writer;

// -- Forward declarations --
static so_R_int_err reader_Read(void* self, so_Slice p);
static so_R_int_err writer_Write(void* self, so_Slice p);

// -- Implementation --

static so_R_int_err reader_Read(void* self, so_Slice p) {
    reader* r = (reader*)self;
    if (so_len(r->b) == 0) {
        return (so_R_int_err){.val = 0, .err = io_EOF};
    }
    so_int n = so_copy(so_byte, p, r->b);
    r->b = so_slice(so_byte, r->b, n, r->b.len);
    return (so_R_int_err){.val = n, .err = NULL};
}

static so_R_int_err writer_Write(void* self, so_Slice p) {
    writer* w = (writer*)self;
    w->b = so_extend(so_byte, w->b, (p));
    return (so_R_int_err){.val = so_len(p), .err = NULL};
}

int main(void) {
    {
        // Copy.
        reader r = (reader){.b = so_string_bytes(so_str("hello world"))};
        writer w = (writer){.b = so_make_slice(so_byte, 0, 11)};
        {
            so_R_i64_err _res1 = io_Copy((io_Writer){.self = &w, .Write = writer_Write}, (io_Reader){.self = &r, .Read = reader_Read});
            so_Error err = _res1.err;
            if (err != NULL) {
                so_panic("Copy failed");
            }
        }
        if (so_string_ne(so_bytes_string(w.b), so_str("hello world"))) {
            so_panic("Copy failed");
        }
    }
    {
        // CopyN.
        reader r = (reader){.b = so_string_bytes(so_str("hello world"))};
        writer w = (writer){.b = so_make_slice(so_byte, 0, 5)};
        {
            so_R_i64_err _res2 = io_CopyN((io_Writer){.self = &w, .Write = writer_Write}, (io_Reader){.self = &r, .Read = reader_Read}, 5);
            so_Error err = _res2.err;
            if (err != NULL) {
                so_panic("CopyN failed");
            }
        }
        if (so_string_ne(so_bytes_string(w.b), so_str("hello"))) {
            so_panic("CopyN failed");
        }
    }
    {
        // ReadAll.
        reader r = (reader){.b = so_string_bytes(so_str("hello world"))};
        so_R_slice_err _res3 = io_ReadAll((mem_Allocator){0}, (io_Reader){.self = &r, .Read = reader_Read});
        so_Slice buf = _res3.val;
        so_Error err = _res3.err;
        if (err != NULL) {
            so_panic("ReadAll failed");
        }
        if (so_string_ne(so_bytes_string(buf), so_str("hello world"))) {
            so_panic("ReadAll failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, buf);
    }
    {
        // ReadFull.
        reader r = (reader){.b = so_string_bytes(so_str("hello world"))};
        so_Slice buf = so_make_slice(so_byte, 11, 11);
        {
            so_R_int_err _res4 = io_ReadFull((io_Reader){.self = &r, .Read = reader_Read}, buf);
            so_Error err = _res4.err;
            if (err != NULL) {
                so_panic("ReadFull failed");
            }
        }
        if (so_string_ne(so_bytes_string(buf), so_str("hello world"))) {
            so_panic("ReadFull failed");
        }
    }
    {
        // WriteString.
        writer w = (writer){.b = so_make_slice(so_byte, 0, 11)};
        so_R_int_err _res5 = io_WriteString((io_Writer){.self = &w, .Write = writer_Write}, so_str("hello world"));
        so_int n = _res5.val;
        so_Error err = _res5.err;
        if (err != NULL) {
            so_panic("WriteString failed");
        }
        if (n != 11 || so_string_ne(so_bytes_string(w.b), so_str("hello world"))) {
            so_panic("WriteString failed");
        }
    }
    {
        // LimitReader.
        reader r = (reader){.b = so_string_bytes(so_str("hello world"))};
        io_LimitedReader lr = io_LimitReader((io_Reader){.self = &r, .Read = reader_Read}, 5);
        so_Slice buf = so_make_slice(so_byte, 5, 5);
        {
            so_R_int_err _res6 = io_LimitedReader_Read(&lr, buf);
            so_Error err = _res6.err;
            if (err != NULL) {
                so_panic("LimitReader failed");
            }
        }
        if (so_string_ne(so_bytes_string(buf), so_str("hello"))) {
            so_panic("LimitReader failed");
        }
    }
}
