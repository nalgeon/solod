package main

// Point represents a 2D coordinate.
type Point struct {
	x int
	y int
}

// NewPoint creates a new Point.
func NewPoint(x int, y int) Point {
	return Point{x: x, y: y}
}

// Scale multiplies both coordinates by a factor.
func (p *Point) Scale(factor int) {
	p.x = p.x * factor
	p.y = p.y * factor
}

// offset is unexported.
func offset(p Point, dx int, dy int) Point {
	return Point{x: p.x + dx, y: p.y + dy}
}

// MaxCoord is the maximum coordinate value.
const MaxCoord int = 1000

func main() {
	// Create a point.
	p := NewPoint(1, 2)
	// Scale and offset.
	p.Scale(MaxCoord)
	p = offset(p, 1, 1)
	_ = p
}
