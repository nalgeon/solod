package main

import "example/interop/src/sub"

//so:include <stdint.h>
//so:include "person.ext.h"

//so:extern INT64_MAX
const maxInt64 = 1<<63 - 1

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

//so:extern unsigned char
type uchar uint8

// An extern type uchar comes from C header, so an exported function
// can name it even when it is unexported.
func FirstChar(buf *uchar) uchar {
	return *buf
}

func main() {
	{
		// Passing values between So and C and vice versa.
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
		// Calling variadic C functions from So.
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
