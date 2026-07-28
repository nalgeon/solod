package main

func main() {
	// long is mangled to long_, which is already taken.
	long_ := 1
	long := 2
	_ = long_
	_ = long
}
