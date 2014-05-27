package planeview

import (
	"math"

	"github.com/xlab/teg-workshop/dioid"
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

const MaxGenerated int = 50

type vertex struct {
	X, Y   float64
	parent *Plane
	temp   bool
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

func (v *vertex) transfer() {
	v.temp = true
	v.parent.dioid = v.parent.dioid.RemoveGd(v.G(), v.D())
	for i, v2 := range v.parent.defined {
		if v2 == v {
			v.parent.defined = append(v.parent.defined[:i], v.parent.defined[i+1:]...)
			break
		}
	}
	v.parent.temporary = append(v.parent.temporary, v)
}

func (v *vertex) shift(dx, dy float64) {
	if !v.temp {
		v.transfer()
	}
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
	dioid     dioid.Serie
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
	p.temporary = append(p.temporary, v...)
}

func (p *Plane) placeV(x, y float64) (v *vertex) {
	v = &vertex{X: x, Y: y, parent: p}
	p.addV(v)
	return
}

func (p *Plane) removeV(v *vertex) {
	if !v.temp {
		v.transfer()
	}
	for i, v2 := range p.temporary {
		if v2 == v {
			p.temporary = append(p.temporary[:i], p.temporary[i+1:]...)
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
	p.mergeTemporary()
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

func (p *Plane) Label() string {
	return p.ioLabel
}

func (p *Plane) SetLabel(label string) {
	p.ioLabel = label
}

func (p *Plane) star(set ...*vertex) {
	var poly dioid.Poly
	for _, v := range set {
		p.dioid = p.dioid.RemoveGd(v.G(), v.D())
		poly = append(poly, dioid.Gd{G: v.G(), D: v.D()})
	}
	p.SetDioid(dioid.SerieOplus(p.dioid, dioid.PolyStar(poly)))
}

func (p *Plane) mergeTemporary() {
	var poly dioid.Poly
	for _, v := range p.temporary {
		poly = append(poly, dioid.Gd{G: v.G(), D: v.D()})
	}
	p.SetDioid(dioid.SerieOplus(p.dioid, dioid.Serie{P: poly}))
}

func (p *Plane) Dioid() (serie dioid.Serie) {
	return p.dioid
}

func (p *Plane) SetDioid(serie dioid.Serie) {
	p.dioid = dioid.SerieCanonize(serie)
	p.temporary = make([]*vertex, 0, 16)
	p.defined = make([]*vertex, 0, 64)
	p.generated = make([]*vertex, 0, MaxGenerated)
	if !serie.P.IsEps() {
		for _, m := range serie.P {
			p.defined = append(p.defined, p.newV(m.G, m.D))
		}
	}
	var gen int
	var g, d int
	if !serie.R.IsE() {
		for gen < MaxGenerated {
			for _, m := range serie.Q {
				if g == 0 && d == 0 {
					p.defined = append(p.defined, p.newV(m.G, m.D))
				} else {
					p.generated = append(p.generated, p.newV(m.G+g, m.D+d))
				}
				gen++
			}
			g, d = g+serie.R.G, d+serie.R.D
		}
	} else {
		for _, m := range serie.Q {
			p.defined = append(p.defined, p.newV(m.G, m.D))
		}
	}
}

func (p *Plane) SetColor(color string) {
	p.color = color
}

func NewPlane(ioId, ioLabel string, input bool) *Plane {
	return &Plane{
		ioId:      ioId,
		ioLabel:   ioLabel,
		input:     input,
		util:      &utility{},
		color:     ColorDefault,
		dioid:     dioid.Serie{},
		defined:   make([]*vertex, 0, 64),
		temporary: make([]*vertex, 0, 16),
		generated: make([]*vertex, 0, MaxGenerated),
		selected:  make(map[*vertex]bool, 256),
	}
}
