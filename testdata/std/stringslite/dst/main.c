#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Clone.
        so_String s = so_str("hello");
        so_String c = stringslite_Clone((mem_Allocator){0}, s);
        if (so_string_ne(c, s)) {
            so_panic("Clone failed");
        }
        mem_FreeString((mem_Allocator){0}, c);
    }
    {
        // Cut.
        so_R_str_str _res1 = stringslite_Cut(so_str("hello world"), so_str(" "));
        so_String before = _res1.val;
        so_String after = _res1.val2;
        if (so_string_ne(before, so_str("hello")) || so_string_ne(after, so_str("world"))) {
            so_panic("Cut failed");
        }
    }
    {
        // CutPrefix.
        so_R_str_bool _res2 = stringslite_CutPrefix(so_str("hello world"), so_str("hello "));
        so_String after = _res2.val;
        bool found = _res2.val2;
        if (so_string_ne(after, so_str("world")) || !found) {
            so_panic("CutPrefix failed");
        }
    }
    {
        // CutSuffix.
        so_R_str_bool _res3 = stringslite_CutSuffix(so_str("hello world"), so_str(" world"));
        so_String before = _res3.val;
        bool found = _res3.val2;
        if (so_string_ne(before, so_str("hello")) || !found) {
            so_panic("CutSuffix failed");
        }
    }
    {
        // HasPrefix.
        if (!stringslite_HasPrefix(so_str("hello world"), so_str("hello"))) {
            so_panic("HasPrefix failed");
        }
        if (stringslite_HasPrefix(so_str("hello world"), so_str("world"))) {
            so_panic("HasPrefix failed");
        }
    }
    {
        // HasSuffix.
        if (!stringslite_HasSuffix(so_str("hello world"), so_str("world"))) {
            so_panic("HasSuffix failed");
        }
        if (stringslite_HasSuffix(so_str("hello world"), so_str("hello"))) {
            so_panic("HasSuffix failed");
        }
    }
    {
        // Index.
        so_int idx = stringslite_Index(so_str("hello world"), so_str("world"));
        if (idx != 6) {
            so_panic("Index failed");
        }
    }
    {
        // IndexByte.
        so_int idx = stringslite_IndexByte(so_str("hello world"), 'o');
        if (idx != 4) {
            so_panic("IndexByte failed");
        }
    }
    {
        // TrimPrefix.
        so_String s = stringslite_TrimPrefix(so_str("hello world"), so_str("hello "));
        if (so_string_ne(s, so_str("world"))) {
            so_panic("TrimPrefix failed");
        }
    }
    {
        // TrimSuffix.
        so_String s = stringslite_TrimSuffix(so_str("hello world"), so_str(" world"));
        if (so_string_ne(s, so_str("hello"))) {
            so_panic("TrimSuffix failed");
        }
    }
}
