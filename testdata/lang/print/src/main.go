package main

type person struct {
	name string
}

func main() {
	var vInt int = 42
	var vFloat float64 = 3.14
	var vBool bool = true
	var vByte byte = 'x'
	var vRune rune = '本'
	var vString string = "hello"
	alice := person{name: "alice"}
	var vPtr = &alice
	println(vInt, vFloat, vBool, vByte, vRune, vString, vPtr)

	print("a")
	print()
	print("b")
	println()

	// Complex types are not supported.
	// arr := [3]int{1, 2, 3}
	// println(arr)
	// alice := person{name: "alice"}
	// println(alice)
}
