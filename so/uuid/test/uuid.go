package uuid_test

import (
	"solod.dev/so/encoding/binary"
	"solod.dev/so/testing"
	"solod.dev/so/time"
	"solod.dev/so/uuid"
)

const (
	nilStr = "00000000-0000-0000-0000-000000000000"
	maxStr = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	ustr   = "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
)

// uval is the value that ustr represents.
var uval = [16]byte{
	0xf8, 0x1d, 0x4f, 0xae,
	0x7d, 0xec,
	0x11, 0xd0,
	0xa7, 0x65,
	0x00, 0xa0, 0xc9, 0x1e, 0x6b, 0xf6,
}

func TestNew(t *testing.T) {
	checkNew(t, "New", uuid.New(), 4)
	checkNew(t, "NewV4", uuid.NewV4(), 4)
	checkNew(t, "NewV7", uuid.NewV7(), 7)
}

func TestNewV7Millis(t *testing.T) {
	u := uuid.NewV7()
	got := binary.BigEndian.Uint64(u.Value[:8]) >> 16
	want := uint64(time.Now().UnixMilli())
	// The clock can pass a millisecond between the two calls.
	if got != want && got+1 != want {
		t.Errorf("NewV7() millis = %d, want %d", got, want)
	}
}

func TestString(t *testing.T) {
	u := uid()
	buf := make([]byte, uuid.UUIDLen)
	if got := u.String(buf); got != ustr {
		t.Errorf("String() = %s, want %s", got, ustr)
	}
	if got := uuid.Nil().String(buf); got != nilStr {
		t.Errorf("Nil().String() = %s, want %s", got, nilStr)
	}
	if got := uuid.Max().String(buf); got != maxStr {
		t.Errorf("Max().String() = %s, want %s", got, maxStr)
	}
}

func TestMarshalText(t *testing.T) {
	u := uid()
	buf := make([]byte, uuid.UUIDLen)
	got, err := u.MarshalText(buf)
	if err != nil {
		t.Fatal("MarshalText failed")
		return
	}
	if string(got) != ustr {
		t.Errorf("MarshalText() = %s, want %s", string(got), ustr)
	}
}

func TestAppendText(t *testing.T) {
	const prefix = "urn:uuid:"
	u := uid()
	buf := make([]byte, 0, len(prefix)+uuid.UUIDLen)
	buf = append(buf, prefix...)
	got, err := u.AppendText(buf)
	if err != nil {
		t.Fatal("AppendText failed")
		return
	}
	if string(got) != prefix+ustr {
		t.Errorf("AppendText() = %s, want %s", string(got), prefix+ustr)
	}
}

func TestUnmarshalText(t *testing.T) {
	var got uuid.UUID
	if err := got.UnmarshalText([]byte(ustr)); err != nil {
		t.Fatal("UnmarshalText failed")
		return
	}
	if !got.Equal(uid()) {
		t.Error("UnmarshalText mismatch")
	}
}

func TestParse(t *testing.T) {
	want := uid()
	// Every accepted form of the same UUID.
	cases := []string{
		"f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
		"F81D4FAE-7DEC-11D0-A765-00A0C91E6BF6",
		"f81d4fae7dec11d0a76500a0c91e6bf6",
		"{f81d4fae-7dec-11d0-a765-00a0c91e6bf6}",
		"urn:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
	}
	for _, s := range cases {
		u, err := uuid.Parse(s)
		if err != nil {
			t.Errorf("Parse(%s) failed", s)
			continue
		}
		if !u.Equal(want) {
			t.Errorf("Parse(%s) mismatch", s)
		}
	}

	unil, err := uuid.Parse(nilStr)
	if err != nil || !unil.Equal(uuid.Nil()) {
		t.Error("Parse(nilStr) != Nil()")
	}
	umax, err := uuid.Parse(maxStr)
	if err != nil || !umax.Equal(uuid.Max()) {
		t.Error("Parse(maxStr) != Max()")
	}
}

func TestParseErrors(t *testing.T) {
	cases := []string{
		"",
		"0000000000000-0000-0000-000000000000",
		"00000000-000000000-0000-000000000000",
		"00000000-0000-000000000-000000000000",
		"00000000-0000-0000-00000000000000000",
		"00000000-0000-0000-0000-00000000000",
		"x0000000-0000-0000-0000-000000000000",
		"00000000-x000-0000-0000-000000000000",
		"00000000-0000-x000-0000-000000000000",
		"00000000-0000-0000-x000-000000000000",
		"00000000-0000-0000-0000-x00000000000",
		"{x0000000-0000-0000-0000-000000000000}",
		"urn:uuid:x000000-0000-0000-0000-000000000000",
		"x0000000000000000000000000000000",
		// Some parsers permit hyphens in non-standard locations,
		// but Solod does not.
		"0000-0000-0000-0000-0000-0000-0000-0000",
		// Combinations of variant encodings that Solod could parse,
		// but does not.
		"{00000000000000000000000000000000}",
		"{urn:uuid:00000000-0000-0000-0000-000000000000}",
		"urn:uuid:00000000000000000000000000000000",
	}
	for _, s := range cases {
		if _, err := uuid.Parse(s); err == nil {
			t.Errorf("Parse(%s) succeeded, want error", s)
		}
	}
}

func TestEqual(t *testing.T) {
	u := uuid.MustParse(ustr)
	if !u.Equal(uid()) {
		t.Error("Equal: parsed != uval")
	}
	if u.Equal(uuid.Nil()) {
		t.Error("Equal: parsed == Nil()")
	}
}

func TestCompare(t *testing.T) {
	// The UUIDs are in ascending order.
	uuids := []uuid.UUID{uuid.Nil(), uuid.MustParse(ustr), uuid.Max()}
	for i, u := range uuids {
		if u.Compare(u) != 0 {
			t.Errorf("Compare: uuids[%d] != itself", i)
		}
		if i == 0 {
			continue
		}
		prev := uuids[i-1]
		if u.Compare(prev) != 1 {
			t.Errorf("Compare: uuids[%d] <= uuids[%d]", i, i-1)
		}
		if prev.Compare(u) != -1 {
			t.Errorf("Compare: uuids[%d] >= uuids[%d]", i-1, i)
		}
	}
}

// checkNew reports a constructor that returns a wrong version or variant.
func checkNew(t *testing.T, name string, u uuid.UUID, version int) {
	if u.Version() != version {
		t.Errorf("%s: version = %d, want %d", name, u.Version(), version)
	}
	if variant(u) != 0b10 {
		t.Errorf("%s: variant = %d, want 2", name, variant(u))
	}
}

// uid returns the UUID that ustr represents.
// Solod cannot use an array value in a composite literal,
// so the test assigns the field separately.
func uid() uuid.UUID {
	var u uuid.UUID
	u.Value = uval
	return u
}

// variant returns the variant field of u.
func variant(u uuid.UUID) byte {
	return u.Value[8] >> 6
}
