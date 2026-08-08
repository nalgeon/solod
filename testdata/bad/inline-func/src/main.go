package main

type cfg struct {
	size int
}

//so:inline
func inl(c cfg) int {
	return c.size
}

func main() {
	_ = inl(cfg{})
}
