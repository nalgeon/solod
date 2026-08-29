package main

func helper() int {
	return 1
}

//so:inline
func Get() int {
	return helper()
}

func main() {
	println(Get())
}
