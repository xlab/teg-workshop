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

func (c *Circle) X() float64 {
	return c.center[0] - c.r
}

func (c *Circle) Y() float64 {
	return c.center[1] - c.r
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

func (r *Rect) Move(x, y float64) {
	r.corner[0] = x
	r.corner[1] = y
}

func (c *Circle) Move(x, y float64) {
	c.center[0] = x
	c.center[1] = y
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

func (r *Rect) Has(x, y float64) bool {
	if x >= r.corner[0] && x <= r.corner[0]+r.w && y >= r.corner[1] && y <= r.corner[1]+r.h {
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
