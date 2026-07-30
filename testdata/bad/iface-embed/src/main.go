package main

type reader interface {
	read() int
}

type readWriter interface {
	reader
	write(v int)
}

type file struct{ n int }

func (f *file) read() int   { return f.n }
func (f *file) write(v int) { f.n = v }

func main() {
	var rw readWriter = &file{}
	rw.write(5)
}
