package hex_test

import (
	"solod.dev/so/encoding/hex"
	"solod.dev/so/mem"
	"solod.dev/so/slices"
	"solod.dev/so/testing"
)

func TestEncode(t *testing.T) {
	src := []byte("Hello Gopher!")
	dst := slices.Make[byte](mem.System, hex.EncodedLen(len(src)))
	defer mem.FreeSlice(mem.System, dst)

	hex.Encode(dst, src)
	if string(dst) != "48656c6c6f20476f7068657221" {
		t.Error("unexpected Encode result")
	}
}

func TestEncodeToString(t *testing.T) {
	src := []byte("Hello Gopher!")
	encoded := hex.EncodeToString(mem.System, src)
	defer mem.FreeString(mem.System, encoded)

	if encoded != "48656c6c6f20476f7068657221" {
		t.Error("unexpected EncodeToString result")
	}
}

func TestDecode(t *testing.T) {
	src := []byte("48656c6c6f20476f7068657221")
	dst := slices.Make[byte](mem.System, hex.DecodedLen(len(src)))
	defer mem.FreeSlice(mem.System, dst)

	n, err := hex.Decode(dst, src)
	if err != nil {
		t.Fatal("Decode failed")
		return
	}
	if string(dst[:n]) != "Hello Gopher!" {
		t.Error("unexpected Decode result")
	}
}

func TestDecodeString(t *testing.T) {
	const s = "48656c6c6f20476f7068657221"
	decoded, err := hex.DecodeString(mem.System, s)
	if err != nil {
		t.Fatal("DecodeString failed")
		return
	}
	defer mem.FreeSlice(mem.System, decoded)

	if string(decoded) != "Hello Gopher!" {
		t.Error("unexpected DecodeString result")
	}
}

func TestDump(t *testing.T) {
	content := []byte("Go is an open source programming language.")
	dmp := hex.Dump(mem.System, content)
	defer mem.FreeString(mem.System, dmp)

	want := "00000000  47 6f 20 69 73 20 61 6e  20 6f 70 65 6e 20 73 6f  |Go is an open so|\n" +
		"00000010  75 72 63 65 20 70 72 6f  67 72 61 6d 6d 69 6e 67  |urce programming|\n" +
		"00000020  20 6c 61 6e 67 75 61 67  65 2e                    | language.|\n"
	if dmp != want {
		t.Error("unexpected Dump result")
	}
}
