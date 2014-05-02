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
	GroupHeaderHeight  = 20.0
	GroupMargin        = 10.0
	PlaceRadius        = 25.0
	ControlPointWidth  = 10.0
	ControlPointHeight = 10.0
	GridDefaultGap     = 16
)

type drawable interface {
	X() float64
	Y() float64
	Width() float64
	Height() float64
}

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
	Align() (float64, float64)
	Copy() item
}

type controlPoint struct {
	*geometry.Rect
	modified bool
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

type group struct {
	*geometry.Rect
	model     *teg
	label     string
	collapsed bool
}

type MagicStroke struct {
	X0, Y0, X1, Y1 float64
}

func newPlace(x, y float64) *place {
	return &place{
		Circle: geometry.NewCircle(x, y, PlaceRadius),
	}
}

func newTransition(x, y float64) *transition {
	return &transition{
		Rect: geometry.NewRect(x-TransitionWidth/2, y-TransitionHeight/2,
			TransitionWidth, TransitionHeight),
	}
}

func newGroup() *group {
	return &group{
		model: newTeg(),
	}
}

func newControlPoint(x float64, y float64, modified bool) *controlPoint {
	return &controlPoint{
		geometry.NewRect(
			x-ControlPointWidth/2,
			y-ControlPointHeight/2,
			ControlPointWidth, ControlPointHeight),
		modified,
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

func (p *place) shiftControls(dx, dy float64) {
	if p.in != nil {
		if p.inControl.modified {
			p.inControl.Shift(dx, dy)
		} else {
			p.resetControlPoint(true)
		}
	}

	if p.out != nil {
		if p.outControl.modified {
			p.outControl.Shift(dx, dy)
		} else {
			p.resetControlPoint(false)
		}
	}
}

func (t *teg) shiftItem(it item, dx float64, dy float64) {
	it.Shift(dx, dy)

	if place, ok := it.(*place); ok {
		place.shiftControls(dx, dy)
	}
}

func (p *place) Copy() item {
	place := newPlace(p.Center().X(), p.Center().Y())
	place.label = p.label
	place.counter = p.counter
	place.timer = p.timer
	return item(place)
}

func (t *transition) Copy() item {
	transition := newTransition(t.Center().X(), t.Center().Y())
	transition.label = t.label
	if t.horizontal {
		transition.Rotate()
	}
	return item(transition)
}

func (g *group) Copy() item {
	group := newGroup()
	/*
		items := make(map[item]bool, len(g.model.places)+len(g.model.transitions))
		for _, p := range g.model.places {
			items[p] = true
		}
		for _, t := range g.model.transitions {
			items[t] = true
		}
		group.model.cloneItems(items)
	*/
	group.model = g.model
	group.label = g.label
	group.collapsed = g.collapsed
	group.updateBounds()
	return item(group)
}

func (t *transition) Align() (float64, float64) {
	x, y := t.Center().X(), t.Center().Y()
	shiftX, shiftY := geometry.Align(int(x), int(y), GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	t.Move(x, y)
	t.Shift(shiftX, shiftY)
	return shiftX, shiftY
}

func (p *place) Align() (float64, float64) {
	x, y := p.Center().X(), p.Center().Y()
	shiftX, shiftY := geometry.Align(int(x), int(y), GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	p.Move(x, y)
	p.Shift(shiftX, shiftY)
	p.shiftControls(shiftX, shiftY)
	return shiftX, shiftY
}

func (g *group) Align() (float64, float64) {
	x, y := g.Center().X(), g.Center().Y()
	shiftX, shiftY := geometry.Align(int(x), int(y), GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	g.Move(x, y)
	g.Shift(shiftX, shiftY)
	return shiftX, shiftY
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

func (t *transition) Has(x, y float64) bool {
	var radius float64
	if t.horizontal {
		radius = (t.Width() + 6.0) / 2
	} else {
		radius = (t.Height() + 6.0) / 2
	}
	return math.Pow(x-t.Center()[0], 2)+math.Pow(y-t.Center()[1], 2) < math.Pow(radius, 2)
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
		point := p.BorderPoint(centerT[0], centerT[1], 15.0)
		p.inControl = newControlPoint(point[0], point[1], false)
	} else {
		centerT := p.out.Center()
		point := p.BorderPoint(centerT[0], centerT[1], 15.0)
		p.outControl = newControlPoint(point[0], point[1], false)
	}
}

func (p *place) Label() string {
	return p.label
}

func (t *transition) Label() string {
	return t.label
}

func (g *group) Label() string {
	return g.label
}

func (p *place) SetLabel(s string) {
	p.label = s
}

func (t *transition) SetLabel(s string) {
	t.label = s
}

func (g *group) SetLabel(s string) {
	g.label = s
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

type teg struct {
	places      []*place
	transitions []*transition
	groups      []*group
	inputs      []*transition
	outputs     []*transition
	selected    map[item]bool

	PlacesLen       int
	GroupsLen       int
	TransitionsLen  int
	Updated         bool // fake trigger
	MagicStroke     *MagicStroke
	MagicStrokeUsed bool
	MagicRectUsed   bool
}

type ControlPointSpec struct {
	X, Y float64
}

type GroupSpec struct {
	X, Y      float64
	Width     float64
	Height    float64
	Label     string
	Selected  bool
	Collapsed bool
	Model     *teg
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

func detectBounds(items map[item]bool) (x0, y0, x1, y1 float64) {
	x0, y0 = math.MaxFloat64, math.MaxFloat64
	x1, y1 = -math.MaxFloat64, -math.MaxFloat64
	if len(items) < 1 {
		return 0, 0, 0, 0
	}
	for it := range items {
		x0, y0, x1, y1 = bound(it, x0, y0, x1, y1)
		/*
			if p, ok := it.(*place); ok {
				if p.inControl != nil {
					x0, y0, x1, y1 = bound(p.inControl, x0, y0, x1, y1)
				}
				if p.outControl != nil {
					x0, y0, x1, y1 = bound(p.outControl, x0, y0, x1, y1)
				}
			}
		*/
	}
	return x0, y0, x1, y1
}

func bound(it drawable,
	x0 float64, y0 float64, x1 float64, y1 float64) (mx0, my0, mx1, my1 float64) {
	mx0, my0, mx1, my1 = x0, y0, x1, y1
	if it.X() < x0 {
		mx0 = it.X()
	}
	if it.Y() < y0 {
		my0 = it.Y()
	}
	xw := it.X() + it.Width()
	if xw > x1 {
		mx1 = xw
	}
	yh := it.Y() + it.Height()
	if yh > y1 {
		my1 = yh
	}
	return
}

func (tg *teg) Items() map[item]bool {
	items := make(map[item]bool,
		len(tg.places)+len(tg.transitions)+len(tg.groups))
	for _, p := range tg.places {
		items[p] = true
	}
	for _, t := range tg.transitions {
		items[t] = true

	}
	for _, g := range tg.groups {
		items[g] = true
	}
	return items
}

func (t *teg) updated() {
	qml.Changed(t, &t.Updated)
}

func (t *teg) updatedMagicStroke() {
	qml.Changed(t, &t.MagicStrokeUsed)
	qml.Changed(t, &t.MagicRectUsed)
}

func (tg *teg) GetPlaceSpec(i int) *PlaceSpec {
	return tg.newPlaceSpec(tg.places[i])
}

func (tg *teg) GetGroupSpec(i int) *GroupSpec {
	return tg.newGroupSpec(tg.groups[i])
}

func (tg *teg) GetTransitionSpec(i int) *TransitionSpec {
	return tg.newTransitionSpec(tg.transitions[i])
}

func (tg *teg) newGroupSpec(g *group) *GroupSpec {
	spec := &GroupSpec{
		X: g.X(), Y: g.Y(), Selected: tg.isSelected(g),
		Collapsed: g.collapsed, Model: g.model, Label: g.Label(),
		Width: g.Width(), Height: g.Height(),
	}
	return spec
}

func (tg *teg) newPlaceSpec(p *place) *PlaceSpec {
	spec := &PlaceSpec{
		X: p.X(), Y: p.Y(), Selected: tg.isSelected(p),
		Counter: p.counter, Timer: p.timer, Label: p.Label(),
	}
	if p.in != nil {
		spec.InControl = &ControlPointSpec{
			X: p.inControl.Center().X(),
			Y: p.inControl.Center().Y(),
		}
	}
	if p.out != nil {
		spec.OutControl = &ControlPointSpec{
			X: p.outControl.Center().X(),
			Y: p.outControl.Center().Y(),
		}
	}
	return spec
}

func (tg *teg) newTransitionSpec(t *transition) *TransitionSpec {
	spec := &TransitionSpec{
		X: t.X(), Y: t.Y(), Selected: tg.isSelected(t),
		In: len(t.in), Out: len(t.out), Label: t.Label(),
		Horizontal: t.horizontal,
	}

	list := make([]interface{}, 0, spec.Out+spec.In)
	for i, p := range t.in {
		list = append(list, &ArcSpec{
			Place:   tg.newPlaceSpec(p),
			Inbound: true, Index: i,
		})
	}
	for i, p := range t.out {
		list = append(list, &ArcSpec{
			Place:   tg.newPlaceSpec(p),
			Inbound: false, Index: i,
		})
	}
	spec.ArcSpecs = NewList(list)
	return spec
}

func (g *group) updateBounds() {
	x0, y0, x1, y1 := detectBounds(g.model.Items())
	g.Rect = geometry.NewRect(x0-GroupMargin, y0-GroupMargin-GroupHeaderHeight,
		(x1-x0)+2*GroupMargin, (y1-y0)+2*GroupMargin+GroupHeaderHeight)
}

func (t *teg) addPlace(x, y float64) *place {
	place := newPlace(x, y)
	t.places = append(t.places, place)
	t.PlacesLen = len(t.places)
	return place
}

func (t *teg) addGroup(items map[item]bool) *group {
	group := newGroup()
	group.model.transferItems(t, items)
	group.updateBounds()
	t.groups = append(t.groups, group)
	t.GroupsLen = len(t.groups)
	return group
}

func (t *teg) removePlace(p *place) {
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

func (t *teg) addTransition(x, y float64) *transition {
	transition := newTransition(x, y)
	t.transitions = append(t.transitions, transition)
	t.TransitionsLen = len(t.transitions)
	return transition
}

func (tg *teg) removeTransition(t *transition) {
	for _, place := range t.in {
		place.out = nil
	}
	for _, place := range t.out {
		place.in = nil
	}
	for i, transition := range tg.transitions {
		if transition == t {
			tg.transitions = append(tg.transitions[:i], tg.transitions[i+1:]...)
		}
	}
	tg.TransitionsLen = len(tg.transitions)
}

func (t *teg) fakeData() {
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

	for _, place := range t.places {
		place.Align()
	}

	for _, transition := range t.transitions {
		transition.Align()
	}

	t.updated()
}

func (t *teg) deselectAll() {
	for k := range t.selected {
		delete(t.selected, k)
	}
}

func (t *teg) deselectItem(it item) {
	delete(t.selected, it)
}

func (t *teg) selectItem(it item) {
	if it != nil {
		t.selected[it] = true
	}
}

func (t *teg) isSelected(it item) bool {
	_, ok := t.selected[it]
	return ok
}

func (tg *teg) areConnected(t *transition, p *place, inbound bool) bool {
	if inbound {
		return p.out == t
	} else {
		return p.in == t
	}
}

func (tg *teg) connectItems(t *transition, p *place, inbound bool) {
	if tg.areConnected(t, p, inbound) {
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

func (tg *teg) disconnectItems(t *transition, p *place, inbound bool) {
	if !tg.areConnected(t, p, inbound) {
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

func (tg *teg) transferItems(tg2 *teg, items map[item]bool) {
	// copy refs to tg
	for it := range items {
		if p, ok := it.(*place); ok {
			tg.places = append(tg.places, p)
		} else if t, ok := it.(*transition); ok {
			tg.transitions = append(tg.transitions, t)
		} else if g, ok := it.(*group); ok {
			tg.groups = append(tg.groups, g)
		}
	}
	tg.PlacesLen = len(tg.places)
	tg.TransitionsLen = len(tg.transitions)
	tg.GroupsLen = len(tg.groups)

	// cleanup refs in tg2
	if tg.PlacesLen > 0 {
		places := make([]*place, 0, 256)
		for _, p := range tg2.places {
			if !items[p] {
				places = append(places, p)
			}
		}
		tg2.places = places
		tg2.PlacesLen = len(places)
	}
	if tg.TransitionsLen > 0 {
		transitions := make([]*transition, 0, 256)
		for _, t := range tg2.transitions {
			if !items[t] {
				transitions = append(transitions, t)
			}
		}
		tg2.transitions = transitions
		tg2.TransitionsLen = len(transitions)
	}
	if tg.GroupsLen > 0 {
		groups := make([]*group, 0, 32)
		for _, g := range tg2.groups {
			if !items[g] {
				groups = append(groups, g)
			}
		}
		tg2.groups = groups
		tg2.GroupsLen = len(groups)
	}
}

func (tg *teg) cloneItems(items map[item]bool) (clones map[item]item) {
	clones = make(map[item]item, len(items))
	for it := range items {
		if t, ok := it.(*transition); ok {
			tNew := t.Copy().(*transition)
			clones[t] = tNew
			tg.transitions = append(tg.transitions, tNew)
			tg.TransitionsLen = len(tg.transitions)
			for _, p := range t.in {
				if _, ok := items[p]; ok {
					var pNew *place
					if pNew, ok = clones[p].(*place); !ok {
						pNew = p.Copy().(*place)
						clones[p] = pNew
						tg.places = append(tg.places, pNew)
						tg.PlacesLen = len(tg.places)
					}
					tg.connectItems(tNew, pNew, true)
					if p.outControl.modified {
						pNew.outControl = newControlPoint(p.outControl.X(),
							p.outControl.Y(), true)
					}
				}
			}
			for _, p := range t.out {
				if _, ok := items[p]; ok {
					var pNew *place
					if pNew, ok = clones[p].(*place); !ok {
						pNew = p.Copy().(*place)
						clones[p] = pNew
						tg.places = append(tg.places, pNew)
						tg.PlacesLen = len(tg.places)
					}
					tg.connectItems(tNew, pNew, false)
					if p.inControl.modified {
						pNew.inControl = newControlPoint(p.inControl.X(),
							p.inControl.Y(), true)
					}
				}
			}
		}
	}
	for k := range clones {
		delete(items, k)
	}
	for it := range items {
		if p, ok := it.(*place); ok {
			var pNew *place
			if pNew, ok = clones[p].(*place); !ok {
				pNew = p.Copy().(*place)
				clones[p] = pNew
				tg.places = append(tg.places, pNew)
				tg.PlacesLen = len(tg.places)
			}
		}
	}
	return
}

func (t *teg) findDrawable(x float64, y float64) (interface{}, bool) {
	for _, it := range t.places {
		switch {
		case it.Has(x, y):
			return item(it), true
		case t.isSelected(it):
			if it.in != nil && it.inControl.Has(x, y) {
				return it.inControl, true
			} else if it.out != nil && it.outControl.Has(x, y) {
				return it.outControl, true
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

func newTeg() *teg {
	return &teg{
		places:      make([]*place, 0, 256),
		transitions: make([]*transition, 0, 256),
		groups:      make([]*group, 0, 32),
		selected:    make(map[item]bool, 256),
		inputs:      make([]*transition, 0, 8),
		outputs:     make([]*transition, 0, 8),

		PlacesLen:      0,
		TransitionsLen: 0,
		GroupsLen:      0,
		Updated:        false,
		MagicStroke:    &MagicStroke{},
	}
}
