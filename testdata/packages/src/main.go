package main

// The same package under its own name and under an alias.
// Both usages emit the geom_ prefix.
import (
	"example/geom"
	g "example/geom"
)

func main() {
	a1 := geom.RectArea(5, 10)
	_ = a1

	_ = geom.Pi

	a2 := g.RectArea(5, 10)
	_ = a2

	_ = g.Pi

	var r g.Rect
	_ = r.Area()
}
