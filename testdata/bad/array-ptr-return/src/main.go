package main

func get(s []string) *[2]string {
	return (*[2]string)(s)
}

func main() {
	_ = get([]string{"a", "b"})
}
