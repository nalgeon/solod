#include "main.h"

// -- Forward declarations (functions and methods) --
static bool isLatinLower(so_rune r);
static bool isPunct(so_rune r);
static so_rune toDot(so_rune r);

// -- Implementation --

static bool isLatinLower(so_rune r) {
    return U'a' <= (uint32_t)(r) && (uint32_t)(r) <= U'z';
}

static bool isPunct(so_rune r) {
    return (uint32_t)(r) == U',' || (uint32_t)(r) == U';';
}

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
        // ContainsAny.
        so_Slice b = so_string_bytes(so_str("I like seafood."));
        if (!bytes_ContainsAny(b, so_str("aei"))) {
            so_panic("ContainsAny failed");
        }
        if (bytes_ContainsAny(b, so_str("xyz"))) {
            so_panic("ContainsAny failed");
        }
    }
    {
        // ContainsRune.
        so_Slice b = so_string_bytes(so_str("I like seafood."));
        if (!bytes_ContainsRune(b, U'f')) {
            so_panic("ContainsRune failed");
        }
        if (bytes_ContainsRune(b, U'x')) {
            so_panic("ContainsRune failed");
        }
    }
    {
        // ContainsFunc.
        if (bytes_ContainsFunc(so_string_bytes(so_str("HELLO")), isLatinLower)) {
            so_panic("ContainsFunc failed");
        }
        if (!bytes_ContainsFunc(so_string_bytes(so_str("World")), isLatinLower)) {
            so_panic("ContainsFunc failed");
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
        // CutPrefix.
        so_Slice b = so_string_bytes(so_str("hello"));
        so_Result _res1 = bytes_CutPrefix(b, so_string_bytes(so_str("hel")));
        so_Slice after = _res1.val.as_slice;
        bool found = _res1.val2.as_bool;
        if (so_string_ne(so_bytes_string(after), so_str("lo")) || !found) {
            so_panic("CutPrefix failed");
        }
    }
    {
        // CutSuffix.
        so_Slice b = so_string_bytes(so_str("hello"));
        so_Result _res2 = bytes_CutSuffix(b, so_string_bytes(so_str("lo")));
        so_Slice before = _res2.val.as_slice;
        bool found = _res2.val2.as_bool;
        if (so_string_ne(so_bytes_string(before), so_str("hel")) || !found) {
            so_panic("CutSuffix failed");
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
        // Fields.
        so_Slice b = so_string_bytes(so_str("go is fun"));
        so_Slice fields = bytes_Fields((mem_Allocator){0}, b);
        if (so_len(fields) != 3) {
            so_panic("Fields failed");
        }
        if (so_string_ne(so_bytes_string(so_at(so_Slice, fields, 0)), so_str("go")) || so_string_ne(so_bytes_string(so_at(so_Slice, fields, 1)), so_str("is")) || so_string_ne(so_bytes_string(so_at(so_Slice, fields, 2)), so_str("fun"))) {
            so_panic("Fields failed");
        }
        mem_FreeSlice(so_Slice, (mem_Allocator){0}, fields);
    }
    {
        // FieldsFunc.
        so_Slice b = so_string_bytes(so_str("go,is;fun"));
        so_Slice fields = bytes_FieldsFunc((mem_Allocator){0}, b, isPunct);
        if (so_len(fields) != 3) {
            so_panic("FieldsFunc failed");
        }
        if (so_string_ne(so_bytes_string(so_at(so_Slice, fields, 0)), so_str("go")) || so_string_ne(so_bytes_string(so_at(so_Slice, fields, 1)), so_str("is")) || so_string_ne(so_bytes_string(so_at(so_Slice, fields, 2)), so_str("fun"))) {
            so_panic("FieldsFunc failed");
        }
        mem_FreeSlice(so_Slice, (mem_Allocator){0}, fields);
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
        // Index, IndexByte, IndexAny, IndexRune.
        so_Slice b = so_string_bytes(so_str("hello"));
        if (bytes_Index(b, so_string_bytes(so_str("l"))) != 2) {
            so_panic("Index failed");
        }
        if (bytes_IndexByte(b, 'e') != 1) {
            so_panic("Index failed");
        }
        if (bytes_IndexAny(b, so_str("aeiou")) != 1) {
            so_panic("IndexAny failed");
        }
        if (bytes_IndexRune(b, U'o') != 4) {
            so_panic("IndexRune failed");
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
        // LastIndex, LastIndexByte, LastIndexAny.
        so_Slice b = so_string_bytes(so_str("hello"));
        if (bytes_LastIndex(b, so_string_bytes(so_str("l"))) != 3) {
            so_panic("LastIndex failed");
        }
        if (bytes_LastIndexByte(b, 'l') != 3) {
            so_panic("LastIndexByte failed");
        }
        if (bytes_LastIndexAny(b, so_str("al")) != 3) {
            so_panic("LastIndexAny failed");
        }
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
        // Repeat.
        so_Slice b = so_string_bytes(so_str("go"));
        so_Slice repeated = bytes_Repeat((mem_Allocator){0}, b, 3);
        if (so_string_ne(so_bytes_string(repeated), so_str("gogogo"))) {
            so_panic("Repeat failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, repeated);
    }
    {
        // Replace and ReplaceAll.
        so_Slice b = so_string_bytes(so_str("hello"));
        so_Slice r1 = bytes_Replace((mem_Allocator){0}, b, so_string_bytes(so_str("l")), so_string_bytes(so_str("x")), 1);
        if (so_string_ne(so_bytes_string(r1), so_str("hexlo"))) {
            so_panic("Replace failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, r1);
        so_Slice r2 = bytes_ReplaceAll((mem_Allocator){0}, b, so_string_bytes(so_str("l")), so_string_bytes(so_str("x")));
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
        // ToTitle.
        so_Slice b = so_string_bytes(so_str("hello"));
        so_Slice titled = bytes_ToTitle((mem_Allocator){0}, b);
        if (so_string_ne(so_bytes_string(titled), so_str("HELLO"))) {
            so_panic("ToTitle failed");
        }
        mem_FreeSlice(so_byte, (mem_Allocator){0}, titled);
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
}
