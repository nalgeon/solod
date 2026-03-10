package main

func xopen(x *int) {
	println("open", *x)
}

func xclose(a any) {
	x := a.(*int)
	println("close", *x)
}

func main() {
	x := 42
	xopen(&x)
	defer xclose(&x)
	println("working...")
}
