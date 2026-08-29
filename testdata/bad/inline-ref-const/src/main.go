package main

const secret = 7

//so:inline
func Get() int {
	return secret
}

func main() {
	println(Get())
}
