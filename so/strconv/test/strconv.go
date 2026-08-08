package strconv_test

import (
	"solod.dev/so/strconv"
	"solod.dev/so/testing"
)

func TestAppendBool(t *testing.T) {
	buf := make([]byte, 0, strconv.MaxBoolLen)
	b := strconv.AppendBool(buf, true)
	if string(b) != "true" {
		t.Error("AppendBool")
	}
}

func TestAppendFloat(t *testing.T) {
	buf := make([]byte, 0, strconv.MaxFloat64Len)
	b := strconv.AppendFloat(buf, 3.1415926535, 'E', -1, 32)
	if string(b) != "3.1415927E+00" {
		t.Error("AppendFloat 32")
	}
	b = strconv.AppendFloat(buf, 3.1415926535, 'E', -1, 64)
	if string(b) != "3.1415926535E+00" {
		t.Error("AppendFloat 64")
	}
}

func TestAppendInt(t *testing.T) {
	buf := make([]byte, 0, strconv.MaxIntBase10Len)
	b := strconv.AppendInt(buf, -42, 10)
	if string(b) != "-42" {
		t.Error("AppendInt base 10")
	}
	b = strconv.AppendInt(buf, -42, 16)
	if string(b) != "-2a" {
		t.Error("AppendInt base 16")
	}
}

func TestAppendUint(t *testing.T) {
	buf := make([]byte, 0, strconv.MaxUintBase10Len)
	b := strconv.AppendUint(buf, 42, 10)
	if string(b) != "42" {
		t.Error("AppendUint base 10")
	}
	b = strconv.AppendUint(buf, 42, 16)
	if string(b) != "2a" {
		t.Error("AppendUint base 16")
	}
}

func TestAtof(t *testing.T) {
	f, err := strconv.ParseFloat("1844674407370955", 64)
	if err != nil {
		t.Fatal("Atof error")
		return
	}
	if f != float64(1844674407370955) {
		t.Error("Atof value")
	}
}

func TestAtoi(t *testing.T) {
	s, err := strconv.Atoi("10")
	if err != nil {
		t.Fatal("Atoi error")
		return
	}
	if s != 10 {
		t.Error("Atoi value")
	}
}

func TestFormatBool(t *testing.T) {
	s := strconv.FormatBool(true)
	if s != "true" {
		t.Error("FormatBool")
	}
}

func TestFormatFloat(t *testing.T) {
	buf := make([]byte, strconv.MaxFloat64Len)
	s := strconv.FormatFloat(buf, 3.1415926535, 'E', -1, 32)
	if s != "3.1415927E+00" {
		t.Error("FormatFloat 32")
	}
	s = strconv.FormatFloat(buf, 3.1415926535, 'E', -1, 64)
	if s != "3.1415926535E+00" {
		t.Error("FormatFloat 64")
	}
	s = strconv.FormatFloat(buf, 3.1415926535, 'g', -1, 64)
	if s != "3.1415926535" {
		t.Error("FormatFloat g")
	}
	s = strconv.FormatFloat(buf, 1844674407370955, 'f', -1, 64)
	if s != "1844674407370955" {
		t.Error("FormatFloat big")
	}
}

func TestFormatInt(t *testing.T) {
	buf := make([]byte, strconv.MaxIntBase10Len)
	s := strconv.FormatInt(buf, -42, 10)
	if s != "-42" {
		t.Error("FormatInt base 10")
	}
	s = strconv.FormatInt(buf, -42, 16)
	if s != "-2a" {
		t.Error("FormatInt base 16")
	}
	s = strconv.FormatInt(buf, int64(1<<31-1), 10)
	if s != "2147483647" {
		t.Error("FormatInt 31bit")
	}
	s = strconv.FormatInt(buf, int64(1<<56-1), 10)
	if s != "72057594037927935" {
		t.Error("FormatInt 56bit")
	}
	s = strconv.FormatInt(buf, int64(1<<62-1), 10)
	if s != "4611686018427387903" {
		t.Error("FormatInt 62bit")
	}
}

func TestFormatUint(t *testing.T) {
	buf := make([]byte, strconv.MaxUintBase10Len)
	s := strconv.FormatUint(buf, 42, 10)
	if s != "42" {
		t.Error("FormatUint base 10")
	}
	s = strconv.FormatUint(buf, 42, 16)
	if s != "2a" {
		t.Error("FormatUint base 16")
	}
}

func TestItoa(t *testing.T) {
	buf := make([]byte, strconv.MaxIntBase10Len)
	s := strconv.Itoa(buf, 10)
	if s != "10" {
		t.Error("Itoa")
	}
}

func TestParseBool(t *testing.T) {
	s, err := strconv.ParseBool("true")
	if err != nil {
		t.Fatal("ParseBool error")
		return
	}
	if !s {
		t.Error("ParseBool value")
	}
}

func TestParseFloat(t *testing.T) {
	buf := make([]byte, strconv.MaxFloat64Len)
	s, err := strconv.ParseFloat("3.1415926535", 32)
	if err != nil {
		t.Fatal("ParseFloat 32 error")
		return
	}
	r := strconv.FormatFloat(buf, s, 'E', -1, 32)
	if r != "3.1415927E+00" {
		t.Error("ParseFloat 32 value")
	}
	s, err = strconv.ParseFloat("3.1415926535", 64)
	if err != nil {
		t.Fatal("ParseFloat 64 error")
		return
	}
	if s != 3.1415926535 {
		t.Error("ParseFloat 64 value")
	}
	// NaN.
	s, err = strconv.ParseFloat("NaN", 32)
	if err != nil {
		t.Fatal("ParseFloat NaN error")
		return
	}
	if s == s {
		t.Error("ParseFloat NaN value")
	}
	// Case insensitive.
	s, err = strconv.ParseFloat("nan", 32)
	if err != nil {
		t.Fatal("ParseFloat nan error")
		return
	}
	if s == s {
		t.Error("ParseFloat nan value")
	}
	// inf.
	s, err = strconv.ParseFloat("inf", 32)
	if err != nil {
		t.Fatal("ParseFloat inf error")
		return
	}
	r = strconv.FormatFloat(buf, s, 'g', -1, 64)
	if r != "+Inf" {
		t.Error("ParseFloat inf value")
	}
	// +Inf.
	s, err = strconv.ParseFloat("+Inf", 32)
	if err != nil {
		t.Fatal("ParseFloat +Inf error")
		return
	}
	r = strconv.FormatFloat(buf, s, 'g', -1, 64)
	if r != "+Inf" {
		t.Error("ParseFloat +Inf value")
	}
	// -Inf.
	s, err = strconv.ParseFloat("-Inf", 32)
	if err != nil {
		t.Fatal("ParseFloat -Inf error")
		return
	}
	r = strconv.FormatFloat(buf, s, 'g', -1, 64)
	if r != "-Inf" {
		t.Error("ParseFloat -Inf value")
	}
	// -0.
	s, err = strconv.ParseFloat("-0", 32)
	if err != nil {
		t.Fatal("ParseFloat -0 error")
		return
	}
	r = strconv.FormatFloat(buf, s, 'g', -1, 64)
	if r != "-0" {
		t.Error("ParseFloat -0 value")
	}
	// +0.
	s, err = strconv.ParseFloat("+0", 32)
	if err != nil {
		t.Fatal("ParseFloat +0 error")
		return
	}
	if s != 0 {
		t.Error("ParseFloat +0 value")
	}
}

func TestParseInt(t *testing.T) {
	s, err := strconv.ParseInt("-354634382", 10, 32)
	if err != nil {
		t.Fatal("ParseInt 32 error")
		return
	}
	if s != -354634382 {
		t.Error("ParseInt 32 value")
	}
	s, err = strconv.ParseInt("-3546343826724305832", 10, 64)
	if err != nil {
		t.Fatal("ParseInt 64 error")
		return
	}
	if s != -3546343826724305832 {
		t.Error("ParseInt 64 value")
	}
}

func TestParseUint(t *testing.T) {
	s, err := strconv.ParseUint("42", 10, 32)
	if err != nil {
		t.Fatal("ParseUint 32 error")
		return
	}
	if s != 42 {
		t.Error("ParseUint 32 value")
	}
	s, err = strconv.ParseUint("42", 10, 64)
	if err != nil {
		t.Fatal("ParseUint 64 error")
		return
	}
	if s != 42 {
		t.Error("ParseUint 64 value")
	}
}
