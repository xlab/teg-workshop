package tegview

import (
	"github.com/xlab/teg-workshop/geometry"
	"gopkg.in/qml.v0"
)

const (
	TransitionWidth    = 6.0
	TransitionHeight   = 30.0
	PlaceRadius        = 25.0
	ControlPointWidth  = 10.0
	ControlPointHeight = 10.0
)

type item interface {
	X() float64
	Y() float64
	Center() *geometry.Point
	Width() float64
	Height() float64
	Has(x, y float64) bool
	Move(x, y float64)
	Shift(dx, dy float64)
	Resize(w, h float64)
	SetLabel(s string)
	Label() string
}

type controlPoint struct {
	*geometry.Rect
	modified bool
}

func (c *controlPoint) Label() (s string) {
	return
}

func (c *controlPoint) SetLabel(s string) {
	return
}

type place struct {
	*geometry.Circle
	counter    int
	timer      int
	label      string
	in         *transition
	out        *transition
	inControl  *controlPoint
	outControl *controlPoint
}

type transition struct {
	*geometry.Rect
	in         []*place
	out        []*place
	label      string
	horizontal bool
}

func NewPlace(x float64, y float64) *place {
	return &place{
		Circle: geometry.NewCircle(x, y, PlaceRadius),
	}
}

func NewTransition(x float64, y float64) *transition {
	return &transition{
		Rect: geometry.NewRect(x, y, TransitionWidth, TransitionHeight),
	}
}

func (t *transition) Move(x, y float64) {
	t.Rect.Move(x-t.Width()/2, y-t.Height()/2)
}

func (c *controlPoint) Move(x, y float64) {
	c.Rect.Move(x-ControlPointWidth/2, y-ControlPointHeight/2)
}

func (t *transition) Rotate() {
	t.horizontal = !t.horizontal
	t.Rect.Rotate(t.horizontal)
}

func (t *transition) resetProperties() {
	t.label = ""
	if t.horizontal {
		t.Rotate()
	}
}

func (p *place) resetProperties() {
	p.counter = 0
	p.timer = 0
	p.label = ""
	if p.in != nil {
		p.resetControlPoint(true)
	}
	if p.out != nil {
		p.resetControlPoint(false)
	}
}

func (p *place) resetControlPoint(inbound bool) {
	if inbound {
		tcenter := p.in.Center()
		point := p.BorderPoint(tcenter[0], tcenter[1], 70.0)
		p.inControl = &controlPoint{
			geometry.NewRect(point[0], point[1], ControlPointWidth, ControlPointHeight),
			false,
		}
	} else {
		tcenter := p.out.Center()
		point := p.BorderPoint(tcenter[0], tcenter[1], 50.0)
		p.outControl = &controlPoint{
			geometry.NewRect(point[0], point[1], ControlPointWidth, ControlPointHeight),
			false,
		}
	}
}

func (p *place) Label() string {
	return p.label
}

func (t *transition) Label() string {
	return t.label
}

func (p *place) SetLabel(s string) {
	p.label = s
}

func (t *transition) SetLabel(s string) {
	t.label = s
}

type List struct {
	list   []interface{}
	Length int
}

func NewList(list []interface{}) *List {
	return &List{
		list:   list,
		Length: len(list),
	}
}

func (l *List) Value(i int) interface{} {
	if i >= 0 && i < l.Length {
		return l.list[i]
	}
	return nil
}

type TegModel struct {
	focus       item
	places      []*place
	transitions []*transition
	selected    map[item]bool

	PlacesLen      int
	TransitionsLen int
	Updated        bool // fake trigger
}

type ControlPointSpec struct {
	X, Y float64
}

type PlaceSpec struct {
	X, Y       float64
	Selected   bool
	Counter    int
	Timer      int
	Label      string
	InControl  *ControlPointSpec
	OutControl *ControlPointSpec
}

type TransitionSpec struct {
	X, Y       float64
	Label      string
	Selected   bool
	Horizontal bool
	In, Out    int
	ArcSpecs   *List
}

type ArcSpec struct {
	Place   *PlaceSpec
	Index   int
	Inbound bool
}

func (t *TegModel) updated() {
	qml.Changed(t, &t.Updated)
}

func (tm *TegModel) GetPlaceSpec(i int) *PlaceSpec {
	return tm.newPlaceSpec(tm.places[i])
}

func (tm *TegModel) GetTransitionSpec(i int) *TransitionSpec {
	return tm.newTransitionSpec(tm.transitions[i])
}

func (tm *TegModel) newPlaceSpec(p *place) *PlaceSpec {
	spec := &PlaceSpec{
		X: p.X(), Y: p.Y(), Selected: tm.isSelected(p),
		Counter: p.counter, Timer: p.timer, Label: p.Label(),
	}
	if p.in != nil {
		spec.InControl = &ControlPointSpec{
			X: p.inControl.X(),
			Y: p.inControl.Y(),
		}
	}
	if p.out != nil {
		spec.OutControl = &ControlPointSpec{
			X: p.outControl.X(),
			Y: p.outControl.Y(),
		}
	}
	return spec
}

func (tm *TegModel) newTransitionSpec(t *transition) *TransitionSpec {
	spec := &TransitionSpec{
		X: t.X(), Y: t.Y(), Selected: tm.isSelected(t),
		In: len(t.in), Out: len(t.out), Label: t.Label(),
		Horizontal: t.horizontal,
	}

	list := make([]interface{}, 0, spec.Out+spec.In)
	for i, p := range t.in {
		list = append(list, &ArcSpec{
			Place:   tm.newPlaceSpec(p),
			Inbound: true, Index: i,
		})
	}
	for i, p := range t.out {
		list = append(list, &ArcSpec{
			Place:   tm.newPlaceSpec(p),
			Inbound: false, Index: i,
		})
	}
	spec.ArcSpecs = NewList(list)
	return spec
}

func (t *TegModel) fakeData() {
	t.places = []*place{
		NewPlace(0, 0),
		NewPlace(0, 120),
		NewPlace(0, 180),
	}

	t.transitions = []*transition{
		NewTransition(0-200, 0),
		NewTransition(0-200, 120),
	}
	t.PlacesLen = len(t.places)
	t.TransitionsLen = len(t.transitions)

	t.connectItems(t.transitions[1], t.places[0], true)
	t.connectItems(t.transitions[1], t.places[0], false)
	t.connectItems(t.transitions[1], t.places[1], true)
}

func (t *TegModel) deselectAll() {
	for k := range t.selected {
		delete(t.selected, k)
	}
}

func (t *TegModel) deselectItem(it item) {
	delete(t.selected, it)
}

func (t *TegModel) selectItem(it item) {
	if it != nil {
		t.selected[it] = true
	}
}

func (t *TegModel) focusedIsPlace() bool {
	_, ok := t.focus.(*place)
	return ok
}

func (t *TegModel) focusedIsTransition() bool {
	_, ok := t.focus.(*place)
	return ok
}

func (t *TegModel) isSelected(it item) bool {
	_, ok := t.selected[it]
	return ok
}

func (tm *TegModel) areConnected(t *transition, p *place, inbound bool) bool {
	if inbound {
		return p.out == t
	} else {
		return p.in == t
	}
}

func (tm *TegModel) connectItems(t *transition, p *place, inbound bool) {
	if tm.areConnected(t, p, inbound) {
		panic("already connected")
	}
	if inbound {
		p.out = t
		t.in = append(t.in, p)
	} else {
		p.in = t
		t.out = append(t.out, p)
	}
	p.resetControlPoint(!inbound)
}

func (t *TegModel) findDrawable(x float64, y float64) (item, bool) {
	for _, it := range t.places {
		switch {
		case it.Has(x, y):
			return item(it), true
		case t.isSelected(it):
			if it.in != nil && it.inControl.Has(x, y) {
				return item(it.inControl), true
			} else if it.out != nil && it.outControl.Has(x, y) {
				return item(it.outControl), true
			}
		}
	}
	for _, it := range t.transitions {
		if it.Has(x, y) {
			return it, true
		}
	}
	return nil, false
}

func NewModel() *TegModel {
	return &TegModel{
		focus:       nil,
		places:      make([]*place, 0, 256),
		transitions: make([]*transition, 0, 256),
		selected:    make(map[item]bool, 256),

		PlacesLen:      0,
		TransitionsLen: 0,
		Updated:        false,
	}
}
