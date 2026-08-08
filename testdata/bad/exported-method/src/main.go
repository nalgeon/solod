package main

type link struct {
	val int
}

type Store struct {
	n int
}

func (s *Store) Take(p *link) int {
	return s.n + p.val
}

func main() {
	s := Store{}
	_ = s.Take(&link{})
}
