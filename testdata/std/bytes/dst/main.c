#include "main.h"

// -- Forward declarations --
static so_rune toDot(so_rune r);

// -- Implementation --

static so_rune toDot(so_rune r) {
    (void)r;
    return U'.';
}

int main(void) {
    {
        // Clone.
        so_Slice b = so_string_bytes(so_str("abc"));
        so_Slice clone = bytes_Clone((mem_Allocator){0}, b);
        if (so_string_ne(so_bytes_string(clone), so_str("abc"))) {
            so_panic("Clone failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, clone);
    }
    {
        // Compare and Equal.
        so_Slice a = so_string_bytes(so_str("abc"));
        so_Slice b = so_string_bytes(so_str("abc"));
        so_Slice c = so_string_bytes(so_str("xyz"));
        if (bytes_Compare(a, b) != 0) {
            so_panic("Compare failed: a != b");
        }
        if (bytes_Compare(a, c) >= 0) {
            so_panic("Compare failed: a >= c");
        }
        if (bytes_Compare(c, a) <= 0) {
            so_panic("Compare failed: c <= a");
        }
        if (!bytes_Equal(a, b)) {
            so_panic("Equal failed: a != b");
        }
        if (bytes_Equal(a, c)) {
            so_panic("Equal failed: a == c");
        }
    }
    {
        // Contains.
        so_Slice b = so_string_bytes(so_str("seafood"));
        if (!bytes_Contains(b, so_string_bytes(so_str("foo")))) {
            so_panic("Contains failed");
        }
        if (bytes_Contains(b, so_string_bytes(so_str("bar")))) {
            so_panic("Contains failed");
        }
    }
    {
        // Count.
        so_Slice b = so_string_bytes(so_str("cheese"));
        if (bytes_Count(b, so_string_bytes(so_str("e"))) != 3) {
            so_panic("Count failed");
        }
        if (bytes_Count(b, so_string_bytes(so_str("x"))) != 0) {
            so_panic("Count failed");
        }
    }
    {
        // Cut.
        so_Slice b = so_string_bytes(so_str("go is fun"));
        bytes_CutResult res = bytes_Cut(b, so_string_bytes(so_str(" is ")));
        if (so_string_ne(so_bytes_string(res.Before), so_str("go")) || so_string_ne(so_bytes_string(res.After), so_str("fun")) || !res.Found) {
            so_panic("Cut failed");
        }
    }
    {
        // Equal.
        so_Slice b = so_string_bytes(so_str("hello"));
        if (!bytes_Equal(b, so_string_bytes(so_str("hello")))) {
            so_panic("Equal failed");
        }
        if (bytes_Equal(b, so_string_bytes(so_str("world")))) {
            so_panic("Equal failed");
        }
    }
    {
        // HasPrefix and HasSuffix.
        so_Slice b = so_string_bytes(so_str("hello"));
        if (!bytes_HasPrefix(b, so_string_bytes(so_str("he")))) {
            so_panic("HasPrefix failed");
        }
        if (bytes_HasPrefix(b, so_string_bytes(so_str("lo")))) {
            so_panic("HasPrefix failed");
        }
        if (!bytes_HasSuffix(b, so_string_bytes(so_str("lo")))) {
            so_panic("HasSuffix failed");
        }
        if (bytes_HasSuffix(b, so_string_bytes(so_str("he")))) {
            so_panic("HasSuffix failed");
        }
    }
    {
        // Index, IndexByte.
        so_Slice b = so_string_bytes(so_str("hello"));
        if (bytes_Index(b, so_string_bytes(so_str("l"))) != 2) {
            so_panic("Index failed");
        }
        if (bytes_IndexByte(b, 'e') != 1) {
            so_panic("Index failed");
        }
    }
    {
        // Join.
        so_Slice b1 = so_string_bytes(so_str("go"));
        so_Slice b2 = so_string_bytes(so_str("is"));
        so_Slice b3 = so_string_bytes(so_str("fun"));
        so_Slice joined = bytes_Join((mem_Allocator){0}, (so_Slice){(so_Slice[3]){b1, b2, b3}, 3, 3}, so_string_bytes(so_str(" ")));
        if (so_string_ne(so_bytes_string(joined), so_str("go is fun"))) {
            so_panic("Join failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, joined);
    }
    {
        // Map.
        so_Slice b = so_string_bytes(so_str("hello"));
        so_Slice mapped = bytes_Map((mem_Allocator){0}, toDot, b);
        if (so_string_ne(so_bytes_string(mapped), so_str("....."))) {
            so_panic("Map failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, mapped);
    }
    {
        // Replace.
        so_Slice b = so_string_bytes(so_str("hello"));
        so_Slice r1 = bytes_Replace((mem_Allocator){0}, b, so_string_bytes(so_str("l")), so_string_bytes(so_str("x")), 1);
        if (so_string_ne(so_bytes_string(r1), so_str("hexlo"))) {
            so_panic("Replace failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, r1);
        so_Slice r2 = bytes_Replace((mem_Allocator){0}, b, so_string_bytes(so_str("l")), so_string_bytes(so_str("x")), -1);
        if (so_string_ne(so_bytes_string(r2), so_str("hexxo"))) {
            so_panic("ReplaceAll failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, r2);
    }
    {
        // Runes.
        so_Slice b = so_string_bytes(so_str("fun"));
        so_Slice runes = bytes_Runes((mem_Allocator){0}, b);
        if (so_len(runes) != 3) {
            so_panic("Runes failed");
        }
        if (so_at(so_rune, runes, 0) != U'f' || so_at(so_rune, runes, 1) != U'u' || so_at(so_rune, runes, 2) != U'n') {
            so_panic("Runes failed");
        }
        mem_FreeSlice(so_rune, (mem_Allocator){0}, runes);
    }
    {
        // Split and SplitN.
        so_Slice b = so_string_bytes(so_str("go is fun"));
        so_Slice s1 = bytes_Split((mem_Allocator){0}, b, so_string_bytes(so_str(" ")));
        if (so_len(s1) != 3) {
            so_panic("Split failed");
        }
        if (so_string_ne(so_bytes_string(so_at(so_Slice, s1, 0)), so_str("go")) || so_string_ne(so_bytes_string(so_at(so_Slice, s1, 1)), so_str("is")) || so_string_ne(so_bytes_string(so_at(so_Slice, s1, 2)), so_str("fun"))) {
            so_panic("Split failed");
        }
        mem_FreeSlice(so_Slice, (mem_Allocator){0}, s1);
        so_Slice s2 = bytes_SplitN((mem_Allocator){0}, b, so_string_bytes(so_str(" ")), 2);
        if (so_len(s2) != 2) {
            so_panic("SplitN failed");
        }
        if (so_string_ne(so_bytes_string(so_at(so_Slice, s2, 0)), so_str("go")) || so_string_ne(so_bytes_string(so_at(so_Slice, s2, 1)), so_str("is fun"))) {
            so_panic("SplitN failed");
        }
        mem_FreeSlice(so_Slice, (mem_Allocator){0}, s2);
    }
    {
        // Trim, TrimLeft, TrimRight.
        so_Slice b = so_string_bytes(so_str("  hello  "));
        if (so_string_ne(so_bytes_string(bytes_Trim(b, so_str(" "))), so_str("hello"))) {
            so_panic("Trim failed");
        }
        if (so_string_ne(so_bytes_string(bytes_TrimLeft(b, so_str(" "))), so_str("hello  "))) {
            so_panic("TrimLeft failed");
        }
        if (so_string_ne(so_bytes_string(bytes_TrimRight(b, so_str(" "))), so_str("  hello"))) {
            so_panic("TrimRight failed");
        }
    }
    {
        // TrimPrefix and TrimSuffix.
        so_Slice b = so_string_bytes(so_str("hello"));
        if (so_string_ne(so_bytes_string(bytes_TrimPrefix(b, so_string_bytes(so_str("he")))), so_str("llo"))) {
            so_panic("TrimPrefix failed");
        }
        if (so_string_ne(so_bytes_string(bytes_TrimSuffix(b, so_string_bytes(so_str("lo")))), so_str("hel"))) {
            so_panic("TrimSuffix failed");
        }
    }
    {
        // ToLower and ToUpper.
        so_Slice b = so_string_bytes(so_str("Hello"));
        so_Slice lowered = bytes_ToLower((mem_Allocator){0}, b);
        if (so_string_ne(so_bytes_string(lowered), so_str("hello"))) {
            so_panic("ToLower failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, lowered);
        so_Slice uppered = bytes_ToUpper((mem_Allocator){0}, b);
        if (so_string_ne(so_bytes_string(uppered), so_str("HELLO"))) {
            so_panic("ToUpper failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, uppered);
    }
    {
        // Buffer.
        bytes_Buffer buf = bytes_NewBuffer((mem_Allocator){0}, so_string_bytes(so_str("hello")));
        bytes_Buffer_Write(&buf, so_string_bytes(so_str(" world")));
        if (so_string_ne(bytes_Buffer_String(&buf), so_str("hello world"))) {
            so_panic("Buffer Write failed");
        }
        bytes_Buffer_Grow(&buf, 64);
        if (bytes_Buffer_Cap(&buf) < 64) {
            so_panic("Buffer Grow failed");
        }
        so_Slice rdbuf = so_make_slice(so_byte, 5, 5);
        so_R_int_err _res1 = bytes_Buffer_Read(&buf, rdbuf);
        so_int n = _res1.val;
        so_Error err = _res1.err;
        if (n != 5 || so_string_ne(so_bytes_string(rdbuf), so_str("hello")) || err != NULL) {
            so_panic("Buffer Read failed");
        }
        if (so_string_ne(bytes_Buffer_String(&buf), so_str(" world"))) {
            so_panic("Buffer Read did not advance the buffer");
        }
        bytes_Buffer_Free(&buf);
    }
    {
        // Reader.
        so_String s = so_str("hello world");
        bytes_Reader r = bytes_NewReader(so_string_bytes(s));
        if (bytes_Reader_Len(&r) != so_len(s)) {
            so_panic("Reader Len failed");
        }
        so_R_slice_err _res2 = io_ReadAll((mem_Allocator){0}, (io_Reader){.self = &r, .Read = bytes_Reader_Read});
        so_Slice b = _res2.val;
        so_Error err = _res2.err;
        if (err != NULL) {
            so_panic(err->msg);
        }
        if (so_string_ne(so_bytes_string(b), s)) {
            so_panic("Reader Read failed");
        }
        if (bytes_Reader_Len(&r) != 0) {
            so_panic("Reader Len failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, b);
    }
}
