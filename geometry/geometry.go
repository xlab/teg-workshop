package geometry

import "math"

type Point struct {
	X, Y float64
}

func (p *Point) Rotate(o *Point, angle float64) *Point {
	return &Point{
		o.X + (p.X-o.X)*math.Cos(angle) - (p.Y-o.Y)*math.Sin(angle),
		o.Y + (p.X-o.X)*math.Sin(angle) + (p.Y-o.Y)*math.Cos(angle),
	}
}

type Rect struct {
	corner *Point
	w, h   float64
}
type Circle struct {
	center *Point
	r      float64
}

func (r *Rect) X() float64 {
	return r.corner.X
}

func (r *Rect) Y() float64 {
	return r.corner.Y
}

func (r *Rect) Corner() *Point {
	return r.corner
}

func (r *Rect) Rotate(horizontal bool) {
	r.corner.X -= (r.h - r.w) / 2
	r.corner.Y += (r.h - r.w) / 2
	r.w, r.h = r.h, r.w
}

func (c *Circle) X() float64 {
	return c.center.X - c.r
}

func (c *Circle) Y() float64 {
	return c.center.Y - c.r
}

func (c *Circle) Corner() *Point {
	return &Point{c.center.X - c.r, c.center.Y - c.r}
}

func Align(x, y, gap float64) (shiftX, shiftY float64) {
	dx := math.Abs(float64(int(x) % int(gap)))
	dy := math.Abs(float64(int(y) % int(gap)))
	if dx == 0 && dy == 0 {
		return 0, 0
	}
	sx, sy := -1.0, -1.0
	if dx > gap/2 {
		sx = 1.0
		dx = gap - dx
	}
	if dy > gap/2 {
		sy = 1.0
		dy = gap - dy
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
	angle := math.Atan2(x-c.center.X, y-c.center.Y)
	dX := (c.r + distance) * math.Sin(angle)
	dY := (c.r + distance) * math.Cos(angle)
	return &Point{c.center.X + dX, c.center.Y + dY}
}

func (p *Point) Move(x, y float64) {
	p.X, p.Y = x, y
}

func (p *Point) Shift(dx, dy float64) {
	p.X += dx
	p.Y += dy
}

func (r *Rect) Move(x, y float64) {
	r.corner.X = x
	r.corner.Y = y
}

func (r *Rect) Shift(dx, dy float64) {
	r.corner.X += dx
	r.corner.Y += dy
}

func (c *Circle) Move(x, y float64) {
	c.center.X = x
	c.center.Y = y
}

func (c *Circle) Shift(dx, dy float64) {
	c.center.X += dx
	c.center.Y += dy
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
		r.corner.X + r.w/2,
		r.corner.Y + r.h/2,
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
		c.center.X-c.r,
		c.center.Y-c.r,
		c.r*2, c.r*2)
}

func (c *Circle) Has(x, y float64) bool {
	return math.Pow(x-c.center.X, 2)+math.Pow(y-c.center.Y, 2) < math.Pow(c.r, 2)
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
	if x >= r.corner.X && x <= r.corner.X+r.w && y >= r.corner.Y && y <= r.corner.Y+r.h {
		return true
	}
	return false
}

func CheckSegmentsCrossing(p0, p1, p2, p3 *Point) bool {
	x0, y0, x1, y1 := p0.X, p0.Y, p1.X, p1.Y
	x2, y2, x3, y3 := p2.X, p2.Y, p3.X, p3.Y
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

func NewRect(x, y, width, height float64) *Rect {
	return &Rect{
		corner: &Point{x, y},
		w:      width,
		h:      height,
	}
}
