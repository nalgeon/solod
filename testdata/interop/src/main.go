package main

import (
	"fmt"

	"example/interop/src/sub"
)

//so:include <stdint.h>
//so:include "person.ext.h"

//so:extern INT64_MAX
const maxInt64 = 1<<63 - 1

// MaxHalf references an extern constant, which comes from a C header
// and needs no declaration of its own.
const MaxHalf = maxInt64 / 2

//so:extern write_func_t
type WriteFunc func(a *Account, format string, args ...any)

//so:extern Account
type Account struct {
	name    string
	balance int64
	flags   []uint8
	write   WriteFunc
}

//so:extern
func (acc *Account) Log(format string, args ...any)

//so:extern Account
type Acc Account

func account_inc_balance(acc *Account, amount int64) int64

//so:extern nodecay
func account_set_name(acc *Account, name string)

//so:extern
func printf(format string, args ...any) int

//so:extern
func write_acc(acc *Account, format string, args ...any)

// measure sums the arguments the kinds string describes. It is nodecay, so
// every argument arrives as a Solod type.
//
//so:extern nodecay
func measure(kinds string, args ...any) int

//so:extern nodecay
func (acc *Account) Measure(kinds string, args ...any) int

// An extern function body is never emitted,
// so it can call the Go standard library.
//
//so:extern
func acc_name(acc *Account) string {
	return fmt.Sprintf("%s", acc.name)
}

//so:extern unsigned char
type uchar uint8

// An extern type uchar comes from C header, so an exported function
// can reference it even when it is unexported.
func FirstChar(buf *uchar) uchar {
	return *buf
}

func main() {
	{
		// Passing values between Solod and C and vice versa.
		acc := Account{
			name:    "Alice",
			balance: 100,
			flags:   []uint8{42},
		}

		balBefore := account_inc_balance(&acc, 50)

		println(
			"name =", acc.name,
			"balance =", balBefore, acc.balance,
			"flags[0] =", acc.flags[0],
		)
	}
	{
		// Calling variadic C functions from Solod.
		printf("One: %d\n", 1)
		printf("Two: %d, %d\n", 2, 3)
		printf("Three: %d, %d, %d\n", 4, 5, 6)
	}
	{
		// Extern nodecay functions.
		var acc Account
		name := "Alice"
		account_set_name(&acc, name)
		if acc.name != "Alice" {
			panic("Extern nodecay failed")
		}
	}
	{
		// Extern constants.
		if maxInt64 <= int64(1<<62) {
			panic("maxInt64 <= 1<<62")
		}
	}
	{
		// Extern variadic function.
		acc := Account{name: "Bob"}
		write_acc(&acc, "Hello %s!", "world")
	}
	{
		// Extern nodecay variadic function: the args go flat,
		// at their Solod types, and every scalar widens.
		name := "Alice"
		var i32 int32 = 7
		var u uint = 4
		var u8 uint8 = 3
		var f32 float32 = 1.5
		var acc Account
		got := measure("ssiiiiiuudp",
			name, "Bob", 10, -8, i32, true, 'A', u, u8, f32, &acc)
		want := len(name) + len("Bob") + 10 - 8 + int(i32) + 1 + int('A') +
			int(u) + int(u8) + int(f32) + 1
		if got != want {
			panic("measure failed")
		}
		if measure("") != 0 {
			panic("empty measure failed")
		}
	}
	{
		// Extern nodecay variadic method.
		acc := Account{balance: 20}
		if acc.Measure("is", 5, "abc") != 28 {
			panic("Measure failed")
		}
	}
	{
		// Extern variadic method.
		acc := Account{name: "Eve"}
		acc.Log("Balance: %d", 789)
	}
	{
		// Extern function pointer.
		acc := Account{name: "Charlie", write: write_acc}
		acc.write(&acc, "Balance: %d", 123)
	}
	{
		// Extern function pointer on a type alias.
		acc := Acc{write: write_acc}
		target := Account{name: "Diana"}
		acc.write(&target, "Balance: %d", 456)
	}
	{
		// Extern function pointer from a different package.
		var s sub.Stream
		s.Write = sub.Discard
		s.Write("Hello, %s!", "world")
	}
	{
		// An untyped constant argument takes the C type of the parameter.
		// The type comes from a package the main package does not import.
		const factor = 21
		if sub.Scale(factor) != 42 {
			panic("Scale failed")
		}
	}
	{
		// Multi-word type names.
		var b byte = 'a'
		var ch uchar = uchar(b)
		if byte(ch) != b {
			panic("unexpected uchar value")
		}
		if FirstChar(&ch) != ch {
			panic("unexpected FirstChar value")
		}
	}
}
