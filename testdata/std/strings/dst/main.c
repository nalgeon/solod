#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Clone.
        so_String s = so_str("hello");
        so_String c = strings_Clone((mem_Allocator){0}, s);
        if (so_string_ne(c, s)) {
            so_panic("Clone failed");
        }
        mem_FreeString((mem_Allocator){0}, c);
    }
    {
        // Compare.
        if (strings_Compare(so_str("abc"), so_str("abb")) <= 0) {
            so_panic("Compare failed");
        }
        if (strings_Compare(so_str("abc"), so_str("abd")) >= 0) {
            so_panic("Compare failed");
        }
        if (strings_Compare(so_str("abc"), so_str("abc")) != 0) {
            so_panic("Compare failed");
        }
    }
    {
        // Count.
        so_int n = strings_Count(so_str("hello world"), so_str("o"));
        if (n != 2) {
            so_panic("Count failed");
        }
        n = strings_Count(so_str("hello world"), so_str(""));
        if (n != 12) {
            so_panic("Count failed");
        }
    }
    {
        // Cut.
        so_R_str_str _res1 = strings_Cut(so_str("hello world"), so_str(" "));
        so_String before = _res1.val;
        so_String after = _res1.val2;
        if (so_string_ne(before, so_str("hello")) || so_string_ne(after, so_str("world"))) {
            so_panic("Cut failed");
        }
    }
    {
        // CutPrefix and CutSuffix.
        so_String src = so_str("hello world");
        {
            so_R_str_bool _res2 = strings_CutPrefix(src, so_str("hello"));
            so_String s = _res2.val;
            bool ok = _res2.val2;
            if (!ok || so_string_ne(s, so_str(" world"))) {
                so_panic("CutPrefix failed");
            }
        }
        {
            so_R_str_bool _res3 = strings_CutSuffix(src, so_str("world"));
            so_String s = _res3.val;
            bool ok = _res3.val2;
            if (!ok || so_string_ne(s, so_str("hello "))) {
                so_panic("CutSuffix failed");
            }
        }
    }
    {
        // Index and IndexAny.
        so_int idx = strings_Index(so_str("hello world"), so_str("o"));
        if (idx != 4) {
            so_panic("Index failed");
        }
        idx = strings_IndexAny(so_str("hello world"), so_str("ow"));
        if (idx != 4) {
            so_panic("IndexAny failed");
        }
    }
    {
        // Repeat.
        so_String r = strings_Repeat((mem_Allocator){0}, so_str("abc"), 3);
        if (so_string_ne(r, so_str("abcabcabc"))) {
            so_panic("Repeat failed");
        }
        mem_FreeString((mem_Allocator){0}, r);
    }
    {
        // Replace and ReplaceAll.
        so_String s = so_str("hello world");
        so_String r = strings_Replace((mem_Allocator){0}, s, so_str("o"), so_str("0"), 1);
        if (so_string_ne(r, so_str("hell0 world"))) {
            so_panic("Replace failed");
        }
        mem_FreeString((mem_Allocator){0}, r);
        r = strings_ReplaceAll((mem_Allocator){0}, s, so_str("o"), so_str("0"));
        if (so_string_ne(r, so_str("hell0 w0rld"))) {
            so_panic("ReplaceAll failed");
        }
        mem_FreeString((mem_Allocator){0}, r);
    }
    {
        // Split and Join.
        so_String s = so_str("a,b,c");
        so_Slice parts = strings_Split((mem_Allocator){0}, s, so_str(","));
        if (so_len(parts) != 3 || so_string_ne(so_at(so_String, parts, 0), so_str("a")) || so_string_ne(so_at(so_String, parts, 1), so_str("b")) || so_string_ne(so_at(so_String, parts, 2), so_str("c"))) {
            so_panic("Split failed");
        }
        so_String j = strings_Join((mem_Allocator){0}, parts, so_str(","));
        if (so_string_ne(j, s)) {
            so_panic("Join failed");
        }
        mem_FreeString((mem_Allocator){0}, j);
        mem_FreeSlice(so_String, (mem_Allocator){0}, parts);
    }
    {
        // ToUpper and ToLower.
        so_String s = so_str("Hello, 世界!");
        so_String u = strings_ToUpper((mem_Allocator){0}, s);
        if (so_string_ne(u, so_str("HELLO, 世界!"))) {
            so_panic("ToUpper failed");
        }
        mem_FreeString((mem_Allocator){0}, u);
        so_String l = strings_ToLower((mem_Allocator){0}, s);
        if (so_string_ne(l, so_str("hello, 世界!"))) {
            so_panic("ToLower failed");
        }
        mem_FreeString((mem_Allocator){0}, l);
    }
    {
        // Trim and TrimSpace.
        so_String s = so_str("  hello world  ");
        so_String t = strings_TrimSpace(s);
        if (so_string_ne(t, so_str("hello world"))) {
            so_panic("TrimSpace failed");
        }
        t = strings_Trim(s, so_str(" dh"));
        if (so_string_ne(t, so_str("ello worl"))) {
            so_panic("Trim failed");
        }
    }
    {
        // Builder.
        strings_Builder b = {0};
        strings_Builder_WriteString(&b, so_str("Hello"));
        strings_Builder_WriteByte(&b, ',');
        strings_Builder_WriteRune(&b, U' ');
        strings_Builder_WriteString(&b, so_str("world"));
        so_String s = strings_Builder_String(&b);
        if (so_string_ne(s, so_str("Hello, world"))) {
            so_panic("Builder failed");
        }
        strings_Builder_Free(&b);
    }
    {
        // Reader.
        strings_Reader r = strings_NewReader(so_str("hello world"));
        so_Slice buf = so_make_slice(so_byte, 5, 5);
        so_R_int_err _res4 = strings_Reader_Read(&r, buf);
        so_int n = _res4.val;
        so_Error err = _res4.err;
        if (err != NULL || n != 5 || so_string_ne(so_bytes_string(buf), so_str("hello"))) {
            so_panic("Reader failed");
        }
    }
}
