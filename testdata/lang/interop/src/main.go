package main

//so:include <stdint.h>
//so:include <stdio.h>
//so:include "person.ext.h"

//so:extern INT64_MAX
const maxInt64 = 1<<63 - 1

//so:extern Account
type Account struct {
	name    string
	balance int64
	flags   []uint8
}

func account_inc_balance(acc *Account, amount int64) int64

//so:extern nodecay
func account_set_name(acc *Account, name string)

//so:extern
func printf(format string, args ...any) int

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
		if maxInt64 <= 1<<62 {
			panic("maxInt64 <= 1<<62")
		}
	}
}
