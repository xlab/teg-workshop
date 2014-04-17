package tegview

import (
	"log"

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
	Width() float64
	Height() float64
	Has(x, y float64) bool
	Move(x, y float64)
	Resize(w, h float64)
	Label() string
}

type controlPoint struct {
	*geometry.Rect
}

func (c *controlPoint) Label() (s string) {
	return
}

type place struct {
	*geometry.Circle
	in      *transition
	out     *transition
	counter int
	timer   int
	label   string
	control *controlPoint
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
		control: &controlPoint{
			geometry.NewRect(x+50.0, y+50.0, ControlPointWidth, ControlPointHeight),
		},
	}
}

func NewTransition(x float64, y float64) *transition {
	return &transition{
		Rect: geometry.NewRect(x, y, TransitionWidth, TransitionHeight),
	}
}

func (t *transition) Move(x, y float64) {
	t.Rect.Move(x-TransitionWidth/2, y-t.Height()/2)
}

func (c *controlPoint) Move(x, y float64) {
	c.Rect.Move(x-ControlPointWidth/2, y-ControlPointHeight/2)
}

func (p *place) Label() string {
	return p.label
}

func (t *transition) Label() string {
	return t.label
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
	updateChan  chan *updateRequest

	Updated         bool // fake trigger
	PlaceSpecs      *List
	TransitionSpecs *List
	ArcSpecs        *List
}

type updateRequest struct {
	places, transitions, arcs bool
}

type ControlPointSpec struct {
	X, Y float64
}

type PlaceSpec struct {
	X, Y     float64
	Selected bool
	Counter  int
	Timer    int
	Label    string
	Control  *ControlPointSpec
}

type TransitionSpec struct {
	X, Y       float64
	Label      string
	Selected   bool
	Horizontal bool
	In, Out    int
}

type ArcSpec struct {
	Transition *TransitionSpec
	Place      *PlaceSpec
	Index      int
	Inbound    bool
}

func (t *TegModel) handleUpdates() {
	go func() {
		for {
			update := <-t.updateChan
			if update.places {
				t.updatePlaceSpecs()
				qml.Changed(t, &t.PlaceSpecs)
			}
			if update.transitions {
				t.updateTransitionSpecs()
				qml.Changed(t, &t.TransitionSpecs)
			}
			if update.arcs {
				t.updateArcSpecs()
				qml.Changed(t, &t.ArcSpecs)
			}
			if update.places || update.transitions || update.arcs {
				t.updated()
			}
		}
	}()
}

func (t *TegModel) queueUpdate(places, transitions, arcs bool) {
	t.updateChan <- &updateRequest{places, transitions, arcs}
}

func (t *TegModel) updated() {
	qml.Changed(t, &t.Updated)
}

func (t *TegModel) updatePlaceSpecs() {
	specs := make([]interface{}, len(t.places))
	for k, v := range t.places {
		specs[k] = t.newPlaceSpec(v)
	}
	t.PlaceSpecs = NewList(specs)
}

func (t *TegModel) updateTransitionSpecs() {
	specs := make([]interface{}, len(t.transitions))
	for k, v := range t.transitions {
		specs[k] = t.newTransitionSpec(v)
	}
	t.TransitionSpecs = NewList(specs)
}

func (tm *TegModel) newPlaceSpec(p *place) *PlaceSpec {
	return &PlaceSpec{
		X: p.X(), Y: p.Y(), Selected: tm.isSelected(p),
		Counter: p.counter, Timer: p.timer, Label: p.Label(),
		Control: &ControlPointSpec{X: p.control.X(), Y: p.control.Y()},
	}
}

func (tm *TegModel) newTransitionSpec(t *transition) *TransitionSpec {
	return &TransitionSpec{
		X: t.X(), Y: t.Y(), Selected: tm.isSelected(t),
		In: len(t.in), Out: len(t.out), Label: t.Label(),
		Horizontal: t.horizontal,
	}
}

func (t *TegModel) updateArcSpecs() {
	specs := make([]interface{}, 0, len(t.places)*2)
	for _, v := range t.transitions {
		for i, p := range v.in {
			specs = append(specs, &ArcSpec{
				Transition: t.newTransitionSpec(v),
				Place:      t.newPlaceSpec(p),
				Inbound:    true, Index: i,
			})
		}
		for i, p := range v.out {
			specs = append(specs, &ArcSpec{
				Transition: t.newTransitionSpec(v),
				Place:      t.newPlaceSpec(p),
				Inbound:    false, Index: i,
			})
		}
	}
	t.ArcSpecs = NewList(specs)
}

func (t *TegModel) fakeData() {
	t.places = []*place{
		NewPlace(0, 0),
		NewPlace(0, 120),
		NewPlace(0, 180),
		NewPlace(0, 240),
		NewPlace(0, 300),
		NewPlace(0, 360),
		NewPlace(0, 420),
		NewPlace(0, 480),
		NewPlace(0, 540),
		NewPlace(0, 600),
		NewPlace(0, 660),
	}
	t.transitions = []*transition{NewTransition(0-200, 0), NewTransition(0-200, 120)}
	t.connectItems(t.transitions[1], t.places[0], true)
	t.connectItems(t.transitions[1], t.places[1], false)
	t.connectItems(t.transitions[1], t.places[2], true)
	t.connectItems(t.transitions[1], t.places[3], false)
	t.connectItems(t.transitions[1], t.places[4], true)
	t.connectItems(t.transitions[1], t.places[5], false)
	t.connectItems(t.transitions[1], t.places[6], true)
	t.connectItems(t.transitions[1], t.places[7], false)
	t.connectItems(t.transitions[1], t.places[8], true)
	t.connectItems(t.transitions[1], t.places[9], false)
	t.connectItems(t.transitions[1], t.places[10], true)
}

func (t *TegModel) unfocusItem() {
	t.focus = nil
}

func (t *TegModel) focusItemAt(x float64, y float64, additive bool) {
	it, ok := t.findDrawable(x, y)

	if !ok && !additive {
		t.deselectAll()
	} else if ok {
		if _, ok := it.(*controlPoint); ok {
			t.focus = it
		} else {
			t.focus = it
			if !additive {
				t.deselectAll()
			}
			t.selectItem(it)
		}
	}

	if ok {
		log.Printf("Focused item at (%f,%f): %v", x, y, it)
	}
}

func (t *TegModel) moveFocusedItem(x, y float64) {
	if t.focus != nil {
		t.focus.Move(x, y)
		if place, ok := t.focus.(*place); ok {
			if place.in != nil {
				transition := place.in
				point := place.BorderPoint(transition.X(), transition.Y(), 50.0)
				place.control.Move(point.X(), point.Y())
			} else {
				place.control.Move(x-50.0, y+50.0)
			}
		}
		t.queueUpdateByItem(t.focus)
	}
}

func (t *TegModel) changeTimerOfFocusedItem(delta int) {
	p, _ := t.focus.(*place) // panic is ok
	p.timer += delta
}

func (t *TegModel) changeCounterOfFocusedItem(delta int) {
	p, _ := t.focus.(*place) // panic is ok
	p.counter += delta
}

func (t *TegModel) queueUpdateByItem(it item) {
	var places, transitions bool
	switch it.(type) {
	case *place:
		places = true
	case *transition:
		transitions = true
	}
	t.queueUpdate(places, transitions, true)
}

func (t *TegModel) deselectAll() {
	for k := range t.selected {
		delete(t.selected, k)
	}
	t.queueUpdate(true, true, true)
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

func (t *TegModel) selectItem(it item) {
	if it != nil {
		t.selected[it] = true
		t.queueUpdateByItem(it)
	}
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
	tm.queueUpdate(true, true, true)
}

func (t *TegModel) findDrawable(x float64, y float64) (item, bool) {
	for _, it := range t.places {
		switch {
		case it.Has(x, y):
			return item(it), true
		case t.isSelected(it) && it.control.Has(x, y):
			return item(it.control), true
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
		updateChan:  make(chan *updateRequest, 1024),

		Updated:         false,
		PlaceSpecs:      &List{},
		TransitionSpecs: &List{},
		ArcSpecs:        &List{},
	}
}
