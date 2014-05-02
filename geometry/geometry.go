package geometry

import "math"

type Point [2]float64
type Line [2]*Point

type Rect struct {
	corner *Point
	w, h   float64
}
type Circle struct {
	center *Point
	r      float64
}

func (p *Point) X() float64 {
	return p[0]
}

func (p *Point) Y() float64 {
	return p[1]
}

func (r *Rect) X() float64 {
	return r.corner[0]
}

func (r *Rect) Y() float64 {
	return r.corner[1]
}

func (r *Rect) Rotate(horizontal bool) {
	r.corner[0] -= (r.h - r.w) / 2
	r.corner[1] += (r.h - r.w) / 2
	r.w, r.h = r.h, r.w
}

func (c *Circle) X() float64 {
	return c.center[0] - c.r
}

func (c *Circle) Y() float64 {
	return c.center[1] - c.r
}

func Align(x, y, gap int) (shiftX, shiftY float64) {
	dx, dy := math.Abs(float64(x%gap)), math.Abs(float64(y%gap))
	if dx == 0 && dy == 0 {
		return 0, 0
	}
	sx, sy := -1.0, -1.0
	if dx > float64(gap)/2.0 {
		sx = 1.0
		dx = float64(gap) - dx
	}
	if dy > float64(gap)/2.0 {
		sy = 1.0
		dy = float64(gap) - dy
	}
	if x < 0 {
		dx = -dx
	}
	if y < 0 {
		dy = -dy
	}
	return sx * dx, sy * dy
}

func (c *Circle) BorderPoint(x, y, distance float64) *Point {
	angle := math.Atan2(x-c.center[0], y-c.center[1])
	dX := (c.r + distance) * math.Sin(angle)
	dY := (c.r + distance) * math.Cos(angle)
	return &Point{c.center[0] + dX, c.center[1] + dY}
}

func (p *Point) Move(x, y float64) {
	p[0], p[1] = x, y
}

func (p *Point) Shift(dx, dy float64) {
	p[0] += dx
	p[1] += dy
}

func (r *Rect) Move(x, y float64) {
	r.corner[0] = x
	r.corner[1] = y
}

func (r *Rect) Shift(dx, dy float64) {
	r.corner[0] += dx
	r.corner[1] += dy
}

func (c *Circle) Move(x, y float64) {
	c.center[0] = x
	c.center[1] = y
}

func (c *Circle) Shift(dx, dy float64) {
	c.center[0] += dx
	c.center[1] += dy
}

func (c *Circle) Resize(w, h float64) {
	c.r = w
}

func (r *Rect) Resize(w, h float64) {
	r.w, r.h = w, h
}

func (r *Rect) Width() float64 {
	return r.w
}

func (r *Rect) Height() float64 {
	return r.h
}

func (r *Rect) Center() *Point {
	return &Point{
		r.corner[0] + r.w/2,
		r.corner[1] + r.h/2,
	}
}

func (c *Circle) Height() float64 {
	return c.r * 2
}

func (c *Circle) Width() float64 {
	return c.r * 2
}

func (c *Circle) Center() *Point {
	return c.center
}

func (c *Circle) Bound() *Rect {
	return NewRect(
		c.center[0]-c.r,
		c.center[1]-c.r,
		c.r*2, c.r*2)
}

func (c *Circle) Has(x, y float64) bool {
	return math.Pow(x-c.center[0], 2)+math.Pow(y-c.center[1], 2) < math.Pow(c.r, 2)
}

func (r *Rect) Intersect(r2 *Rect) bool {
	// http://stackoverflow.com/questions/306316/determine-if-two-rectangles-overlap-each-other
	if r.X() < r2.X()+r2.Width() && r.X()+r.Width() > r2.X() &&
		r.Y() < r2.Y()+r2.Height() && r.Y()+r.Height() > r2.Y() {
		return true
	}
	return false
}

func (r *Rect) Has(x, y float64) bool {
	if x >= r.corner[0] && x <= r.corner[0]+r.w && y >= r.corner[1] && y <= r.corner[1]+r.h {
		return true
	}
	return false
}

func CheckSegmentsCrossing(p0, p1, p2, p3 *Point) bool {
	x0, y0, x1, y1 := p0[0], p0[1], p1[0], p1[1]
	x2, y2, x3, y3 := p2[0], p2[1], p3[0], p3[1]
	sx1 := x1 - x0
	sy1 := y1 - y0
	sx2 := x3 - x2
	sy2 := y3 - y2

	s := (-sy1*(x0-x2) + sx1*(y0-y2)) / (-sx2*sy1 + sx1*sy2)
	t := (sx2*(y0-y2) - sy2*(x0-x2)) / (-sx2*sy1 + sx1*sy2)

	if s >= 0 && s <= 1 && t >= 0 && t <= 1 {
		return true
	}
	return false
}

func NewCircle(x, y, r float64) *Circle {
	return &Circle{center: &Point{x, y}, r: r}
}

func NewPoint(x, y float64) *Point {
	return &Point{x, y}
}

func NewRect(x, y, width, height float64) *Rect {
	return &Rect{
		corner: &Point{x, y},
		w:      width,
		h:      height,
	}
}

func (l *Line) Start() *Point {
	return l[0]
}

func (l *Line) End() *Point {
	return l[1]
}

func NewLine(x0, y0, x1, y1 float64) *Line {
	return &Line{&Point{x0, y0}, &Point{x1, y1}}
}
