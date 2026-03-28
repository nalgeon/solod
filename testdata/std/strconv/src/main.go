package main

import "solod.dev/so/strconv"

func main() {
	buf := make([]byte, 64)
	{
		// AppendBool.
		b := strconv.AppendBool(buf[:0], true)
		if string(b) != "true" {
			panic("AppendBool")
		}
	}
	{
		// AppendFloat.
		b := strconv.AppendFloat(buf[:0], 3.1415926535, 'E', -1, 32)
		if string(b) != "3.1415927E+00" {
			panic("AppendFloat 32")
		}
		b = strconv.AppendFloat(buf[:0], 3.1415926535, 'E', -1, 64)
		if string(b) != "3.1415926535E+00" {
			panic("AppendFloat 64")
		}
	}
	{
		// AppendInt.
		b := strconv.AppendInt(buf[:0], -42, 10)
		if string(b) != "-42" {
			panic("AppendInt base 10")
		}
		b = strconv.AppendInt(buf[:0], -42, 16)
		if string(b) != "-2a" {
			panic("AppendInt base 16")
		}
	}
	{
		// AppendUint.
		b := strconv.AppendUint(buf[:0], 42, 10)
		if string(b) != "42" {
			panic("AppendUint base 10")
		}
		b = strconv.AppendUint(buf[:0], 42, 16)
		if string(b) != "2a" {
			panic("AppendUint base 16")
		}
	}
	{
		// Atoi.
		s, err := strconv.Atoi("10")
		if err != nil {
			panic("Atoi error")
		}
		if s != 10 {
			panic("Atoi value")
		}
	}
	{
		// FormatBool.
		s := strconv.FormatBool(true)
		if s != "true" {
			panic("FormatBool")
		}
	}
	{
		// FormatFloat.
		s := strconv.FormatFloat(buf, 3.1415926535, 'E', -1, 32)
		if s != "3.1415927E+00" {
			panic("FormatFloat 32")
		}
		s = strconv.FormatFloat(buf, 3.1415926535, 'E', -1, 64)
		if s != "3.1415926535E+00" {
			panic("FormatFloat 64")
		}
		s = strconv.FormatFloat(buf, 3.1415926535, 'g', -1, 64)
		if s != "3.1415926535" {
			panic("FormatFloat g")
		}
	}
	{
		// FormatInt.
		s := strconv.FormatInt(buf, -42, 10)
		if s != "-42" {
			panic("FormatInt base 10")
		}
		s = strconv.FormatInt(buf, -42, 16)
		if s != "-2a" {
			panic("FormatInt base 16")
		}
	}
	{
		// FormatUint.
		s := strconv.FormatUint(buf, 42, 10)
		if s != "42" {
			panic("FormatUint base 10")
		}
		s = strconv.FormatUint(buf, 42, 16)
		if s != "2a" {
			panic("FormatUint base 16")
		}
	}
	{
		// Itoa.
		s := strconv.Itoa(buf, 10)
		if s != "10" {
			panic("Itoa")
		}
	}
	{
		// ParseBool.
		s, err := strconv.ParseBool("true")
		if err != nil {
			panic("ParseBool error")
		}
		if !s {
			panic("ParseBool value")
		}
	}
	{
		// ParseFloat.
		s, err := strconv.ParseFloat("3.1415926535", 32)
		if err != nil {
			panic("ParseFloat 32 error")
		}
		r := strconv.FormatFloat(buf, s, 'E', -1, 32)
		if r != "3.1415927E+00" {
			panic("ParseFloat 32 value")
		}
		s, err = strconv.ParseFloat("3.1415926535", 64)
		if err != nil {
			panic("ParseFloat 64 error")
		}
		if s != 3.1415926535 {
			panic("ParseFloat 64 value")
		}
		// NaN.
		s, err = strconv.ParseFloat("NaN", 32)
		if err != nil {
			panic("ParseFloat NaN error")
		}
		if s == s {
			panic("ParseFloat NaN value")
		}
		// Case insensitive.
		s, err = strconv.ParseFloat("nan", 32)
		if err != nil {
			panic("ParseFloat nan error")
		}
		if s == s {
			panic("ParseFloat nan value")
		}
		// inf.
		s, err = strconv.ParseFloat("inf", 32)
		if err != nil {
			panic("ParseFloat inf error")
		}
		r = strconv.FormatFloat(buf, s, 'g', -1, 64)
		if r != "+Inf" {
			panic("ParseFloat inf value")
		}
		// +Inf.
		s, err = strconv.ParseFloat("+Inf", 32)
		if err != nil {
			panic("ParseFloat +Inf error")
		}
		r = strconv.FormatFloat(buf, s, 'g', -1, 64)
		if r != "+Inf" {
			panic("ParseFloat +Inf value")
		}
		// -Inf.
		s, err = strconv.ParseFloat("-Inf", 32)
		if err != nil {
			panic("ParseFloat -Inf error")
		}
		r = strconv.FormatFloat(buf, s, 'g', -1, 64)
		if r != "-Inf" {
			panic("ParseFloat -Inf value")
		}
		// -0.
		s, err = strconv.ParseFloat("-0", 32)
		if err != nil {
			panic("ParseFloat -0 error")
		}
		r = strconv.FormatFloat(buf, s, 'g', -1, 64)
		if r != "-0" {
			panic("ParseFloat -0 value")
		}
		// +0.
		s, err = strconv.ParseFloat("+0", 32)
		if err != nil {
			panic("ParseFloat +0 error")
		}
		if s != 0 {
			panic("ParseFloat +0 value")
		}
	}
	{
		// ParseInt.
		s, err := strconv.ParseInt("-354634382", 10, 32)
		if err != nil {
			panic("ParseInt 32 error")
		}
		if s != -354634382 {
			panic("ParseInt 32 value")
		}
		s, err = strconv.ParseInt("-3546343826724305832", 10, 64)
		if err != nil {
			panic("ParseInt 64 error")
		}
		if s != -3546343826724305832 {
			panic("ParseInt 64 value")
		}
	}
	{
		// ParseUint.
		s, err := strconv.ParseUint("42", 10, 32)
		if err != nil {
			panic("ParseUint 32 error")
		}
		if s != 42 {
			panic("ParseUint 32 value")
		}
		s, err = strconv.ParseUint("42", 10, 64)
		if err != nil {
			panic("ParseUint 64 error")
		}
		if s != 42 {
			panic("ParseUint 64 value")
		}
	}
}
