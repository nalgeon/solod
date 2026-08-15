package uuid_test

import (
	"solod.dev/so/runtime"
	"solod.dev/so/testing"
	"solod.dev/so/uuid"
)

const ustr = "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"

func TestNew(t *testing.T) {
	if !runtime.Hosted {
		t.Skip("needs a hosted environment")
		return
	}
	u4 := uuid.NewV4()
	if u4.Version() != 4 {
		t.Error("NewV4() version != 4")
	}
	u7 := uuid.NewV7()
	if u7.Version() != 7 {
		t.Error("NewV7() version != 7")
	}
}

func TestStringParse(t *testing.T) {
	u1 := uuid.MustParse(ustr)
	buf := make([]byte, uuid.UUIDLen)
	s := u1.String(buf)
	if s != ustr {
		t.Error("String() mismatch")
	}
	u2, err := uuid.Parse(s)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !u1.Equal(u2) {
		t.Error("Parse/String mismatch")
	}
}

func TestCompare(t *testing.T) {
	unil := uuid.Nil()
	uid := uuid.MustParse(ustr)
	umax := uuid.Max()
	if uid.Compare(unil) <= 0 {
		t.Error("Compare: uid <= unil")
	}
	if uid.Compare(umax) >= 0 {
		t.Error("Compare: uid >= umax")
	}
	if uid.Compare(uid) != 0 {
		t.Error("Compare: uid != uid")
	}
}
