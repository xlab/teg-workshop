package tegview

import (
	"math"
	"sort"

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

type MagicStroke struct {
	X0, Y0, X1, Y1 float64
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

type places []*place
type placesByX struct{ places }
type placesByY struct{ places }

func (p places) Len() int      { return len(p) }
func (p places) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p placesByX) Less(i, j int) bool {
	return p.places[i].Center()[0] < p.places[j].Center()[0]
}
func (p placesByY) Less(i, j int) bool {
	return p.places[i].Center()[1] < p.places[j].Center()[1]
}

func (t *transition) OrderArcs(inbound bool) {
	if inbound {
		if t.horizontal {
			sort.Sort(placesByX{t.in})
		} else {
			sort.Sort(placesByY{t.in})
		}
	} else {
		if t.horizontal {
			sort.Sort(placesByX{t.out})
		} else {
			sort.Sort(placesByY{t.out})
		}
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
	t.OrderArcs(true)
	t.OrderArcs(false)
}

func (t *transition) Bound() *geometry.Rect {
	return t.Rect
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
		centerT := p.in.Center()
		point := p.BorderPoint(centerT[0], centerT[1], 25.0)
		p.inControl = &controlPoint{
			geometry.NewRect(point[0], point[1], ControlPointWidth, ControlPointHeight),
			false,
		}
	} else {
		centerT := p.out.Center()
		point := p.BorderPoint(centerT[0], centerT[1], 10.0)
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

	PlacesLen       int
	TransitionsLen  int
	Updated         bool // fake trigger
	MagicStroke     *MagicStroke
	MagicStrokeUsed bool
	MagicRectUsed   bool
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

func (t *TegModel) updatedMagicStroke() {
	qml.Changed(t, &t.MagicStrokeUsed)
	qml.Changed(t, &t.MagicRectUsed)
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

func (t *TegModel) addPlace(x, y float64) *place {
	place := NewPlace(x, y)
	t.places = append(t.places, place)
	t.PlacesLen = len(t.places)
	return place
}

func (t *TegModel) removePlace(p *place) {
	if p.in != nil {
		t.disconnectItems(p.in, p, false)
	}
	if p.out != nil {
		t.disconnectItems(p.out, p, true)
	}
	for i, place := range t.places {
		if place == p {
			t.places = append(t.places[:i], t.places[i+1:]...)
		}
	}
	t.PlacesLen = len(t.places)
}

func (t *TegModel) addTransition(x, y float64) *transition {
	transition := NewTransition(x, y)
	t.transitions = append(t.transitions, transition)
	t.TransitionsLen = len(t.transitions)
	return transition
}

func (tm *TegModel) removeTransition(t *transition) {
	for _, place := range t.in {
		place.out = nil
	}
	for _, place := range t.out {
		place.in = nil
	}
	for i, transition := range tm.transitions {
		if transition == t {
			tm.transitions = append(tm.transitions[:i], tm.transitions[i+1:]...)
		}
	}
	tm.TransitionsLen = len(tm.transitions)
}

func (t *TegModel) fakeData() {
	p1 := t.addPlace(-138.863281, -96.941406)
	p1.timer, p1.counter = 8, 0
	p2 := t.addPlace(-138.468750, 2.511719)
	p2.timer, p2.counter = 3, 0
	p3 := t.addPlace(67.714844, -56.714844)
	p3.timer, p3.counter = 5, 3
	p4 := t.addPlace(-61.046875, -56.257812)
	p4.timer, p4.counter = 3, 4
	p5 := t.addPlace(4.554688, 149.593750)
	p5.timer, p5.counter = 2, 2
	p5.label = "delayed\ncycle"
	p6 := t.addPlace(155.371094, -63.152344)
	p7 := t.addPlace(118.007812, 8.843750)

	t1 := t.addTransition(-265.976562, -46.679688)
	t1.label = "input sink"
	t2 := t.addTransition(6.859375, 29.503906)
	t2.label = "x1"
	t3 := t.addTransition(7.425781, -145.574219)
	t3.label = "x2"
	t4 := t.addTransition(253.656250, -35.816406)
	t4.label = "output sink"

	t.connectItems(t1, p1, false)
	t.connectItems(t1, p2, false)
	t.connectItems(t2, p2, true)
	t.connectItems(t2, p3, true)
	t.connectItems(t2, p4, false)
	t.connectItems(t2, p5, true)
	t.connectItems(t2, p5, false)
	t.connectItems(t2, p7, false)
	t.connectItems(t3, p1, true)
	t.connectItems(t3, p3, false)
	t.connectItems(t3, p4, true)
	t.connectItems(t3, p6, false)
	t.connectItems(t4, p6, true)
	t.connectItems(t4, p7, true)

	p5.inControl.Move(68.4609375, 106.078125)
	p5.inControl.modified = true
	p5.outControl.Move(-55.18359375, 102.73046875)
	p5.outControl.modified = true
	p3.inControl.Move(73.6484375, -94.69921875)
	p3.inControl.modified = true
	p3.outControl.Move(-21.8046875, -32.6015625)
	p3.outControl.modified = true
	p4.inControl.Move(18.88671875, -50.78515625)
	p4.inControl.modified = true
	p4.outControl.Move(-57.30078125, -109.2734375)
	p4.outControl.modified = true
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
		return
	}
	var changed bool
	if inbound && p.out == nil {
		p.out = t
		t.in = append(t.in, p)
		p.resetControlPoint(false)
		t.OrderArcs(true)
		changed = true
	} else if !inbound && p.in == nil {
		p.in = t
		t.out = append(t.out, p)
		t.OrderArcs(false)
		p.resetControlPoint(true)
		changed = true
	}

	if changed {
		if t.horizontal {
			w := calcTransitionHeight(len(t.in), len(t.out))
			t.Resize(w, TransitionWidth) // rotated
		} else {
			h := calcTransitionHeight(len(t.in), len(t.out))
			t.Resize(TransitionWidth, h) // normal
		}
	}
}

func (tm *TegModel) disconnectItems(t *transition, p *place, inbound bool) {
	if !tm.areConnected(t, p, inbound) {
		return
	}
	var changed bool
	if inbound && p.out != nil {
		p.out = nil
		for i, it := range t.in {
			if it == p {
				t.in = append(t.in[:i], t.in[i+1:]...)
			}
		}
		p.outControl = nil
		changed = true
	} else if !inbound && p.in != nil {
		p.in = nil
		for i, it := range t.out {
			if it == p {
				t.out = append(t.out[:i], t.out[i+1:]...)
			}
		}
		p.inControl = nil
		changed = true
	}

	if changed {
		if t.horizontal {
			w := calcTransitionHeight(len(t.in), len(t.out))
			t.Resize(w, TransitionWidth) // rotated
		} else {
			h := calcTransitionHeight(len(t.in), len(t.out))
			t.Resize(TransitionWidth, h) // normal
		}
	}
}

func calcTransitionHeight(in, out int) float64 {
	return math.Max(math.Max(float64(in)*TransitionHeight/2.0,
		float64(out)*TransitionHeight/2.0), TransitionHeight)
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
		MagicStroke:    &MagicStroke{},
	}
}
