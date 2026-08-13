#include "main.h"

// -- Embeds --

//go:build ignore

#ifndef so_build_hosted

// This case stands in for the board. A freestanding environment has no entropy
// source and no clock, so it defines the target hooks that crypto/crand and
// time need. A real target reads its hardware here.

// so_crand_read counts up, so the test can check the bytes it gets back.
so_int so_crand_read(uint8_t* buf, so_int size) {
    for (so_int i = 0; i < size; i++) {
        buf[i] = (uint8_t)(i + 1);
    }
    return size;
}

// fakeMono counts the nanoseconds that so_time_mono and so_time_sleep report.
static int64_t fakeMono = 1000000;

// so_time_wall reports a fixed 2009-11-10T23:00:00Z, the Go playground date.
so_R_i64_i32 so_time_wall(void) {
    return (so_R_i64_i32){.val = 1257894000, .val2 = 0};
}

// so_time_mono adds a millisecond on every call, so elapsed time always grows.
int64_t so_time_mono(void) {
    fakeMono += 1000000;
    return fakeMono;
}

// so_time_sleep adds the full duration to the monotonic count.
void so_time_sleep(int64_t ns) {
    fakeMono += ns;
}

#endif  // !so_build_hosted

// -- Types --

typedef struct hexWord hexWord;

// hexWord is a byte with a hexadecimal text form.
typedef struct hexWord {
    so_byte Value;
} hexWord;

// -- Forward declarations --
static so_R_slice_err hexWord_AppendText(void* self, so_Slice b);
static void check(bool ok, so_String msg);

// -- Variables and constants --
so_Error main_ErrCheck = errors_New("check failed");

// -- Implementation --

// AppendText appends the hexadecimal form of the byte to b.
static so_R_slice_err hexWord_AppendText(void* self, so_Slice b) {
    hexWord* w = self;
    so_int n = so_len(b);
    b = so_append(so_byte, b, 2, 0, 0);
    hex_Encode(so_slice(so_byte, b, n, b.len), (so_Slice){(so_byte[1]){w->Value}, 1, 1});
    return (so_R_slice_err){.val = b, .err = (so_Error){}};
}

static void check(bool ok, so_String msg) {
    if (!ok) {
        so_panic(so_cstr(msg));
    }
}

int main(void) {
    so_Slice heap = so_make_slice(so_byte, 4096, 4096);
    mem_Arena arena = mem_NewArena(heap);
    mem_Allocator alloc = (mem_Allocator){.self = &arena, .Alloc = mem_Arena_Alloc, .Free = mem_Arena_Free, .Realloc = mem_Arena_Realloc};
    {
        // bufio, bytes, io
        bytes_Reader src = bytes_NewReader(so_string_bytes(so_str("hello")));
        bufio_Reader br = bufio_NewReaderSize(alloc, (io_Reader){.self = &src, .Read = bytes_Reader_Read}, 64);
        so_R_byte_err _res1 = bufio_Reader_ReadByte(&br);
        so_byte b = _res1.val;
        so_Error err = _res1.err;
        check(err.self == NULL, so_str("bufio: read failed"));
        check(b == 'h', so_str("bufio: wrong byte"));
        bufio_Reader_Free(&br);
        bytes_Buffer w = bytes_NewBuffer(alloc, (so_Slice){});
        so_R_int_err _res2 = io_WriteString((io_Writer){.self = &w, .Write = bytes_Buffer_Write}, so_str("hello"));
        so_int n = _res2.val;
        err = _res2.err;
        check(err.self == NULL, so_str("io: write failed"));
        check(n == 5, so_str("io: wrong count"));
        check(so_string_eq(so_bytes_string(bytes_Buffer_Bytes(&w)), so_str("hello")), so_str("io: wrong text"));
        bytes_Buffer_Free(&w);
        check(bytes_Equal(so_string_bytes(so_str("ab")), so_string_bytes(so_str("ab"))), so_str("bytes: not equal"));
    }
    {
        // bytealg, strings, unicode, unicode/utf8
        check(bytealg_IndexByteString(so_str("hello"), 'e') == 1, so_str("bytealg: wrong index"));
        check(strings_Contains(so_str("hello"), so_str("ell")), so_str("strings: no substring"));
        check(unicode_IsDigit('5'), so_str("unicode: not alloc digit"));
        check(utf8_RuneCountInString(so_str("héllo")) == 5, so_str("utf8: wrong rune count"));
    }
    {
        // c, unsafe
        check(c_Sizeof(int32_t) == 4, so_str("c: wrong size"));
        check(unsafe_Sizeof((int32_t)(0)) == 4, so_str("unsafe: wrong size"));
    }
    {
        // cmp, slices
        check(cmp_Compare(so_int, (1), (2)) < 0, so_str("cmp: wrong order"));
        so_Slice nums = (so_Slice){(so_int[3]){3, 1, 2}, 3, 3};
        slices_Sort(so_int, (nums));
        check(so_at(so_int, nums, 0) == 1, so_str("slices: not sorted"));
        check(slices_Contains(so_int, (nums), (3)), so_str("slices: no value"));
    }
    {
        // crypto/crand. A freestanding host has no entropy source of its own,
        // so the target defines so_crand_read (see main.c).
        so_Slice buf = so_make_slice(so_byte, 4, 4);
        so_R_int_err _res3 = crand_Read(buf);
        so_int n = _res3.val;
        so_Error err = _res3.err;
        check(err.self == NULL, so_str("crand: read failed"));
        check(n == 4, so_str("crand: wrong count"));
        if (HOOKED) {
            check(so_at(so_byte, buf, 0) == 1 && so_at(so_byte, buf, 3) == 4, so_str("crand: wrong bytes"));
        }
    }
    {
        // encoding
        encoding_TextAppender app = (encoding_TextAppender){.self = &(hexWord){0x0a}, .AppendText = hexWord_AppendText};
        so_R_slice_err _res4 = app.AppendText(app.self, so_make_slice(so_byte, 0, 2));
        so_Slice out = _res4.val;
        so_Error err = _res4.err;
        check(err.self == NULL, so_str("encoding: append failed"));
        check(so_string_eq(so_bytes_string(out), so_str("0a")), so_str("encoding: wrong text"));
    }
    {
        // encoding/binary, encoding/hex
        binary_LE le = {};
        so_Slice word = (so_Slice){(so_byte[4]){0, 0, 0, 0}, 4, 4};
        binary_LE_PutUint32(le, word, 0x01020304);
        check(binary_LE_Uint32(le, word) == 0x01020304, so_str("binary: wrong value"));
        so_Slice dst = so_make_slice(so_byte, hex_EncodedLen(so_len(word)), hex_EncodedLen(so_len(word)));
        hex_Encode(dst, word);
        check(so_string_eq(so_bytes_string(dst), so_str("04030201")), so_str("hex: wrong encoding"));
    }
    {
        // encoding/json
        so_Slice out = so_make_slice(so_byte, 64, 64);
        strings_Builder sb = strings_FixedBuilder(out);
        json_Encoder enc = json_NewEncoder((io_Writer){.self = &sb, .Write = strings_Builder_Write});
        json_Encoder_BeginObject(&enc);
        json_Encoder_Str(&enc, so_str("pi"));
        json_Encoder_Float(&enc, 3.5);
        json_Encoder_EndObject(&enc);
        json_Encoder_Flush(&enc);
        check(json_Encoder_Err(&enc).self == NULL, so_str("json: encode failed"));
        check(so_string_eq(strings_Builder_String(&sb), so_str("{\"pi\":3.5}")), so_str("json: wrong text"));
    }
    {
        // errors
        so_Error err = main_ErrCheck;
        check(so_string_eq(err.Error(err.self), so_str("check failed")), so_str("errors: wrong text"));
    }
    {
        // fmt. The print family works. Print, Println and Printf drop the
        // bytes, because a freestanding host has no standard output. The scan
        // family panics.
        so_Slice buf = so_make_slice(so_byte, 32, 32);
        so_String text = fmt_Sprintf(buf, so_str("n=%d"), (so_int)(42));
        check(so_string_eq(text, so_str("n=42")), so_str("fmt: wrong text"));
        so_Slice out = so_make_slice(so_byte, 32, 32);
        strings_Builder sb = strings_FixedBuilder(out);
        so_R_int_err _res5 = fmt_Fprintf((io_Writer){.self = &sb, .Write = strings_Builder_Write}, so_str("%s=%d"), so_str("n"), (so_int)(42));
        so_int cnt = _res5.val;
        so_Error err = _res5.err;
        check(err.self == NULL, so_str("fmt: write failed"));
        check(cnt == 4, so_str("fmt: wrong write count"));
        check(so_string_eq(strings_Builder_String(&sb), so_str("n=42")), so_str("fmt: wrong output"));
        so_R_int_err _res6 = fmt_Printf(so_str("%d\n"), (so_int)(42));
        cnt = _res6.val;
        err = _res6.err;
        check(err.self == NULL, so_str("fmt: print failed"));
        check(cnt == 3, so_str("fmt: wrong print count"));
    }
    {
        // maps
        maps_Map m = maps_New(so_String, so_int, (alloc), (8));
        maps_Map_Set(so_String, so_int, (&m), (so_str("one")), (1));
        check(maps_Map_Len(so_String, so_int, (&m)) == 1, so_str("maps: wrong length"));
        check(maps_Map_Get(so_String, so_int, (&m), (so_str("one"))) == 1, so_str("maps: wrong value"));
        maps_Map_Free(so_String, so_int, (&m));
    }
    {
        // math/bits, math/rand, runtime
        check(bits_Len(255) == 8, so_str("bits: wrong length"));
        check(rand_Uint64() != rand_Uint64(), so_str("rand: repeated value"));
        check(runtime_NumCPU() >= 1, so_str("runtime: no CPU"));
        check(runtime_Seed() != 0, so_str("runtime: zero seed"));
        if (HOOKED) {
            // Seed draws from so_crand_read, which counts up.
            check(runtime_Seed() == 0x0807060504030201, so_str("runtime: wrong seed"));
        }
    }
    {
        // mem
        so_int* p = mem_Alloc(so_int, (alloc));
        *p = 42;
        check(*p == 42, so_str("mem: wrong value"));
        mem_Free(so_int, (alloc), (p));
    }
    {
        // net/netip. A numeric zone works in freestanding mode,
        // but an interface name resolves to no zone.
        netip_AddrResult _res7 = netip_ParseAddr(so_str("fe80::1%2"));
        netip_Addr ip = _res7.val;
        so_Error err = _res7.err;
        check(err.self == NULL, so_str("netip: parse failed"));
        so_Slice buf = so_make_slice(so_byte, netip_MaxZoneLen, netip_MaxZoneLen);
        check(so_string_eq(netip_Addr_Zone(ip, buf), so_str("2")), so_str("netip: wrong zone"));
        netip_Addr named = netip_Addr_WithZone(ip, so_str("eth0"));
        check(so_string_eq(netip_Addr_Zone(named, buf), so_str("")), so_str("netip: name resolved to a zone"));
    }
    {
        // path
        check(so_string_eq(path_Base(so_str("/dir/file.txt")), so_str("file.txt")), so_str("path: wrong base"));
        check(so_string_eq(path_Ext(so_str("/dir/file.txt")), so_str(".txt")), so_str("path: wrong extension"));
    }
    {
        // strconv
        so_Slice buf = so_make_slice(so_byte, 32, 32);
        check(so_string_eq(strconv_Itoa(buf, -42), so_str("-42")), so_str("strconv: wrong text"));
        so_R_int_err _res8 = strconv_Atoi(so_str("42"));
        so_int n = _res8.val;
        so_Error err = _res8.err;
        check(err.self == NULL, so_str("strconv: parse failed"));
        check(n == 42, so_str("strconv: wrong number"));
    }
    {
        // sync/atomic
        atomic_Int64 cnt = {};
        atomic_Int64_Store(&cnt, 1);
        check(atomic_Int64_Add(&cnt, 2) == 3, so_str("atomic: wrong sum"));
        check(atomic_Int64_CompareAndSwap(&cnt, 3, 7), so_str("atomic: swap failed"));
        check(atomic_Int64_Load(&cnt) == 7, so_str("atomic: wrong value"));
    }
    {
        // time. Now, Since, Until and Sleep draw from the target hooks
        // so_time_wall, so_time_mono and so_time_sleep (see main.c).
        if (HOOKED) {
            // so_time_wall reports the Go playground date.
            time_Time now = time_Now();
            so_Slice buf = so_make_slice(so_byte, time_RFC3339Len + 1, time_RFC3339Len + 1);
            so_String got = time_Time_Format(now, buf, time_RFC3339, time_UTC);
            check(so_string_eq(got, so_str("2009-11-10T23:00:00Z")), so_str("time: wrong time from the hook"));
            time_Time start = time_Now();
            time_Sleep(10 * time_Millisecond);
            time_Duration elapsed = time_Since(start);
            check(elapsed >= 10 * time_Millisecond, so_str("time: sleep did not elapse"));
            check(time_Until(start) <= 0, so_str("time: start is not in the past"));
        }
    }
    {
        // time. Format only supports named layouts in freestanding mode.
        time_Time ts = time_Date(2026, time_August, 11, 12, 0, 0, 0, time_UTC);
        check(time_Time_Year(ts) == 2026, so_str("time: wrong year"));
        so_Slice buf = so_make_slice(so_byte, time_RFC3339Len + 1, time_RFC3339Len + 1);
        so_String got = time_Time_Format(ts, buf, time_RFC3339, time_UTC);
        check(so_string_eq(got, so_str("2026-08-11T12:00:00Z")), so_str("time: wrong text"));
    }
    {
        // uuid. New and NewV4 draw from crypto/crand, so they need
        // so_crand_read. NewV7 also reads the clock, so it needs
        // so_time_wall as well.
        uuid_UUID u = uuid_New();
        so_Slice buf = so_make_slice(so_byte, uuid_UUIDLen, uuid_UUIDLen);
        so_R_slice_err _res9 = uuid_UUID_MarshalText(u, buf);
        so_Slice text = _res9.val;
        so_Error err = _res9.err;
        check(err.self == NULL, so_str("uuid: marshal failed"));
        check(so_len(text) == uuid_UUIDLen, so_str("uuid: wrong length"));
        uuid_UUIDResult _res10 = uuid_Parse(so_bytes_string(text));
        uuid_UUID parsed = _res10.val;
        err = _res10.err;
        check(err.self == NULL, so_str("uuid: parse failed"));
        check(uuid_UUID_Equal(parsed, u), so_str("uuid: round trip changed the value"));
        check(uuid_UUID_Version(u) == 4, so_str("uuid: wrong version"));
        uuid_UUID v7 = uuid_NewV7();
        check(uuid_UUID_Version(v7) == 7, so_str("uuid: wrong v7 version"));
        check(!uuid_UUID_Equal(v7, uuid_Nil()), so_str("uuid: v7 is the nil value"));
    }
    so_println("%s", "ok");
    return 0;
}
