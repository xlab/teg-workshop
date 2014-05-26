package planeview

import (
	"math"

	"github.com/xlab/teg-workshop/geometry"
)

const (
	PointRadius = 3.5
	PadOffset   = 4.0
	GridGap     = 16.0
)

const (
	UtilNone = iota
	UtilRect
)

type vertex struct {
	X, Y   float64
	parent *Plane
}

func (v *vertex) G() int {
	rat := v.X / GridGap
	if rat-math.Floor(rat) < 0.5 {
		return int(rat)
	} else {
		return int(math.Ceil(rat))
	}
}

func (v *vertex) D() int {
	rat := -v.Y / GridGap
	if rat-math.Floor(rat) < 0.5 {
		return int(rat)
	} else {
		return int(math.Ceil(rat))
	}
}

func (v *vertex) align() {
	v.X = float64(v.G()) * GridGap
	v.Y = -float64(v.D()) * GridGap
}

func (v *vertex) isSelected() bool {
	return v.parent.selected[v]
}

func (v *vertex) move(x, y float64) {
	v.X, v.Y = x, y
}

func (v *vertex) shift(dx, dy float64) {
	v.X, v.Y = v.X+dx, v.Y+dy
}

func (v *vertex) has(x, y float64) bool {
	return math.Pow(x-v.X, 2)+math.Pow(y-v.Y, 2) < math.Pow(PointRadius+PadOffset, 2)
}

func (v *vertex) bound() *geometry.Rect {
	return geometry.NewRect(v.X-PointRadius, v.Y-PointRadius, 2*PointRadius, 2*PointRadius)
}

type utility struct {
	min, max *geometry.Point
	kind     int
}

type Plane struct {
	id        int
	input     bool
	ioId      string
	ioLabel   string
	util      *utility
	color     string
	defined   []*vertex
	temporary []*vertex
	generated []*vertex
	selected  map[*vertex]bool
	updated   chan int
}

func (p *Plane) IsInput() bool {
	return p.input
}

func (p *Plane) update() {
	p.updated <- p.id
}

func (p *Plane) addV(v ...*vertex) {
	p.defined = append(p.defined, v...)
}

func (p *Plane) placeV(x, y float64) (v *vertex) {
	v = &vertex{X: x, Y: y, parent: p}
	p.addV(v)
	return
}

func (p *Plane) removeV(v *vertex) {
	for i, v2 := range p.defined {
		if v2 == v {
			p.defined = append(p.defined[:i], p.defined[i+1:]...)
			return
		}
	}
}

func (p *Plane) newV(g, d int) *vertex {
	return &vertex{
		X: float64(g) * GridGap, Y: -float64(d) * GridGap,
		parent: p,
	}
}

func (p *Plane) FakeData() {
	p.addV(p.newV(3, 2), p.newV(4, 4), p.newV(5, 5))
}

func (p *Plane) deselectAll() {
	for v := range p.selected {
		delete(p.selected, v)
	}
}

func (p *Plane) deselectV(v ...*vertex) {
	for _, it := range v {
		delete(p.selected, it)
	}
}

func (p *Plane) selectV(v ...*vertex) {
	for _, it := range v {
		p.selected[it] = true
	}
}

func (p *Plane) findV(x, y float64) (*vertex, bool) {
	for _, v := range p.defined {
		if v.has(x, y) {
			return v, true
		}
	}
	for _, v := range p.temporary {
		if v.has(x, y) {
			return v, true
		}
	}
	return nil, false
}

func NewPlane(ioId, ioLabel string, input bool) *Plane {
	return &Plane{
		ioId:      ioId,
		ioLabel:   ioLabel,
		input:     input,
		util:      &utility{},
		color:     ColorDefault,
		defined:   make([]*vertex, 0, 256),
		temporary: make([]*vertex, 0, 16),
		generated: make([]*vertex, 0, 256),
		selected:  make(map[*vertex]bool, 256),
	}
}
