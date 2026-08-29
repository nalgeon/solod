package geom

const Pi = 3.14159

type Rect struct {
	W, H float64
}

func (r Rect) Area() float64 {
	return rectArea(r.W, r.H)
}

func rectArea(width, height float64) float64 {
	return width * height
}

func RectArea(width, height float64) float64 {
	return rectArea(width, height)
}
