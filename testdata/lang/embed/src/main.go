package main

import (
	_ "embed"
)

//so:embed main.h
var main_h string

//so:embed main.c
var main_c string

var GoSecret int64 = 42

func getCSecret() int64

func main() {
	cSecret := getCSecret()
	if cSecret != GoSecret {
		panic("secret mismatch")
	}
}
