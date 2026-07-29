package main

func main() {
	var a any = 1
	p := &a
	// An empty interface is a void*, so it compares with a pointer.
	if a == p {
		println("same")
	}
	// It does not compare with a value: a holds a pointer to the value,
	// so the comparison would compare an address with a value.
	if a == 1 {
		println("one")
	}
}
