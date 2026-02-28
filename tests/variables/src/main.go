package main

type person struct {
	age int
}

func main() {
	{
		// Definition with var and explicit type.
		var vInt int = 42
		_ = vInt
		var vFloat float64 = 3.14
		_ = vFloat
		var vBool bool = true
		_ = vBool
		var vByte byte = 'x'
		_ = vByte
		var vRune rune = '本'
		_ = vRune
		var vString string = "hello"
		_ = vString
		var vSlice []int = []int{1, 2, 3}
		_ = vSlice
		var vStruct = person{age: 42}
		var vPtr = &vStruct
		_ = vPtr
		var vAnyVal any = 42
		_ = vAnyVal
		var vAnyPtr any = vPtr
		_ = vAnyPtr
	}

	{
		// Definition with var and type inference.
		var vInt = 42
		_ = vInt
		var vFloat = 3.14
		_ = vFloat
		var vBool = true
		_ = vBool
		var vByte = 'x'
		_ = vByte
		var vRune = '本'
		_ = vRune
		var vString = "hello"
		_ = vString
		var vSlice = []int{1, 2, 3}
		_ = vSlice
		var vStruct = person{age: 42}
		var vPtr = &vStruct
		_ = vPtr
		var vAnyVal = any(42)
		_ = vAnyVal
		var vAnyPtr = any(vPtr)
		_ = vAnyPtr
	}

	{
		// Definition with short variable declaration.
		vInt := 42
		_ = vInt
		vFloat := 3.14
		_ = vFloat
		vBool := true
		_ = vBool
		vByte := 'x'
		_ = vByte
		vRune := '本'
		_ = vRune
		vString := "hello"
		_ = vString
		vSlice := []int{1, 2, 3}
		_ = vSlice
		vStruct := person{age: 42}
		vPtr := &vStruct
		_ = vPtr
		vAnyVal := any(42)
		_ = vAnyVal
		vAnyPtr := any(vPtr)
		_ = vAnyPtr
	}

	{
		// Multiple typed variable declaration.
		var a, b, c int = 11, 22, 33
		_ = a
		_ = b
		_ = c
		var b1, b2 byte = 'a', 'b'
		_ = b1
		_ = b2
		var s1, s2 string = "foo", "bar"
		_ = s1
		_ = s2
		var a1, a2 []int = []int{1, 2}, []int{3, 4}
		_ = a1
		_ = a2
		var p1, p2 person = person{age: 42}, person{age: 43}
		_ = p1
		_ = p2
	}

	{
		// Multiple untyped variable declaration.
		var vInt, vFloat, vBool = 42, 3.14, true
		_ = vInt
		_ = vFloat
		_ = vBool
		var vByte, vRune, vString = 'x', '本', "hello"
		_ = vByte
		_ = vRune
		_ = vString
		var vSlice, vStruct = []int{1, 2, 3}, person{age: 42}
		_ = vSlice
		_ = vStruct
	}

	{
		// Multiple variable declaration with short variable declaration.
		vInt, vFloat, vBool := 42, 3.14, true
		_ = vInt
		_ = vFloat
		_ = vBool
		vByte, vRune, vString := 'x', '本', "hello"
		_ = vByte
		_ = vRune
		_ = vString
		vSlice, vStruct := []int{1, 2, 3}, person{age: 42}
		_ = vSlice
		_ = vStruct
	}
}
