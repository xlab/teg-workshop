package tegview

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/xlab/teg-workshop/geometry"
	"github.com/xlab/teg-workshop/planeview"
	"github.com/xlab/teg-workshop/util"
)

const (
	TransitionWidth    = 6.0
	TransitionHeight   = 30.0
	GroupMargin        = 10.0
	GroupIOSpacing     = 20.0
	GroupMinSize       = 150.0
	ControlPointWidth  = 10.0
	ControlPointHeight = 10.0
	PlaceRadius        = 25.0
	GridDefaultGap     = 16.0
)

var (
	PlaneColors = []string{
		"#d35400", "#2980b9", "#27ae60",
		"#2c3e50", "#16a085", "#e67e22",
		"#f1c40f", "#2ecc71", "#9b59b6",
	}
)

const (
	TransitionInternal = iota
	TransitionInput
	TransitionOutput
	TransitionExposed
)

const (
	PlaceInternal = iota
	PlaceExposed
)

const (
	GroupInternal = iota
	GroupExposed
)

const (
	UtilNone = iota
	UtilStroke
	UtilRect
)

type drawable interface {
	X() float64
	Y() float64
	Width() float64
	Height() float64
}

type item interface {
	Id() string
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
	IsSelected() bool
}

type controlPoint struct {
	*geometry.Rect
	modified bool
}

type place struct {
	*geometry.Circle
	id         string
	counter    int
	timer      int
	label      string
	in         *transition
	out        *transition
	inControl  *controlPoint
	outControl *controlPoint
	parent     *teg
}

type transition struct {
	*geometry.Rect
	id         string
	in         []*place
	out        []*place
	proxy      *transition
	group      *group
	label      string
	horizontal bool
	parent     *teg
	kind       int
}

type group struct {
	*geometry.Rect
	id      string
	iostate map[*transition]*transition
	inputs  []*transition
	outputs []*transition
	label   string
	folded  bool
	model   *teg
	parent  *teg
}

type utility struct {
	min, max *geometry.Point
	kind     int
}

func newPlace(x, y float64) *place {
	return &place{
		Circle: geometry.NewCircle(x, y, PlaceRadius),
		id:     util.GenUUID(),
	}
}

func newTransition(x, y float64) *transition {
	return &transition{
		Rect: geometry.NewRect(x-TransitionWidth/2, y-TransitionHeight/2,
			TransitionWidth, TransitionHeight),
		id: util.GenUUID(),
	}
}

func newGroup() *group {
	return &group{
		model:   newTeg(),
		inputs:  make([]*transition, 0, 8),
		outputs: make([]*transition, 0, 8),
		id:      util.GenUUID(),
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

type ControlPoint struct {
	X, Y     float64
	Modified bool
}

type Place struct {
	Id         string
	X, Y       float64
	Counter    int
	Timer      int
	Label      string
	InControl  *ControlPoint
	OutControl *ControlPoint
}

type Transition struct {
	Id         string
	X, Y       float64
	In, Out    []*Place
	ProxyId    string
	Label      string
	Horizontal bool
	Kind       int
}

type Group struct {
	Id      string
	X, Y    float64
	Inputs  []*Transition
	Outputs []*Transition
	Iostate map[string]string
	Label   string
	Folded  bool
	Model   *Teg
}

type Teg struct {
	Id          string
	Places      []*Place
	Transitions []*Transition
	Groups      []*Group
}

func (cp *controlPoint) Model() *ControlPoint {
	return &ControlPoint{cp.X(), cp.Y(), cp.modified}
}

func (p *place) Model() *Place {
	model := &Place{
		Id: p.id,
		X:  p.Center().X, Y: p.Center().Y,

		Counter: p.counter,
		Timer:   p.timer,
		Label:   p.label,
	}
	if p.in != nil {
		model.InControl = p.inControl.Model()
	}
	if p.out != nil {
		model.OutControl = p.outControl.Model()
	}
	return model
}

func (t *transition) Model() *Transition {
	model := &Transition{
		Id: t.id,
		X:  t.X(), Y: t.Y(),

		Label:      t.label,
		Kind:       t.kind,
		Horizontal: t.horizontal,

		In:  make([]*Place, len(t.in)),
		Out: make([]*Place, len(t.out)),
	}
	if t.proxy != nil {
		model.ProxyId = t.proxy.id
	}
	for i, p := range t.in {
		model.In[i] = p.Model()
	}
	for i, p := range t.out {
		model.Out[i] = p.Model()
	}
	return model
}

func (g *group) Model(copy bool) *Group {
	model := &Group{
		Id: g.id,
		X:  g.X(), Y: g.Y(),

		Label:  g.label,
		Folded: g.folded,
		Model:  g.model.Model(copy),

		Inputs:  make([]*Transition, len(g.inputs)),
		Outputs: make([]*Transition, len(g.outputs)),
		Iostate: make(map[string]string, len(g.inputs)+len(g.outputs)),
	}
	if copy {
		model.Id = util.GenUUID()
	}
	for i, t := range g.inputs {
		model.Inputs[i] = t.Model()
	}
	for i, t := range g.outputs {
		model.Outputs[i] = t.Model()
	}
	for proxy, io := range g.iostate {
		model.Iostate[io.id] = proxy.id
	}
	return model
}

func (tg *teg) Model(copy bool) *Teg {
	model := &Teg{
		Id: tg.id,

		Places:      make([]*Place, 0, len(tg.places)),
		Transitions: make([]*Transition, len(tg.transitions)),
		Groups:      make([]*Group, len(tg.groups)),
	}
	if copy {
		model.Id = util.GenUUID()
	}
	ignore := make(map[*place]bool, len(tg.places))
	for i, t := range tg.transitions {
		model.Transitions[i] = t.Model()
		for _, p := range t.in {
			ignore[p] = true
		}
		for _, p := range t.out {
			ignore[p] = true
		}
	}
	for _, p := range tg.places {
		if !ignore[p] {
			model.Places = append(model.Places, p.Model())
		}
	}
	for i, g := range tg.groups {
		model.Groups[i] = g.Model(copy)
	}
	return model
}

func (tg *teg) ModelItems(items map[item]bool) *Teg {
	sub := newTeg()
	for it := range items {
		if p, ok := it.(*place); ok {
			sub.places = append(sub.places, p)
		} else if t, ok := it.(*transition); ok {
			sub.transitions = append(sub.transitions, t)
		} else if g, ok := it.(*group); ok {
			sub.groups = append(sub.groups, g)
		}
	}
	return sub.Model(true)
}

func (tg *teg) ConstructItems(model *Teg) (items map[item]bool) {
	sub := newTeg()
	sub.Construct(model)
	items = sub.Items()
	tg.transferItems(sub, items)
	return
}

func (cp *controlPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(cp.Model())
}
func (p *place) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Model())
}
func (t *transition) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Model())
}
func (g *group) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Model(false))
}
func (tg *teg) MarshalJSON() ([]byte, error) {
	return json.Marshal(tg.Model(false))
}

func constructTransition(model *Transition) *transition {
	t := &transition{
		Rect:       geometry.NewRect(model.X, model.Y, TransitionWidth, TransitionHeight),
		id:         model.Id,
		horizontal: model.Horizontal,
		label:      model.Label,
		kind:       model.Kind,
	}
	if t.horizontal {
		t.Rect.Rotate(t.horizontal)
	}
	return t
}

func constructPlace(model *Place) *place {
	p := &place{
		Circle:  geometry.NewCircle(model.X, model.Y, PlaceRadius),
		id:      model.Id,
		timer:   model.Timer,
		counter: model.Counter,
		label:   model.Label,
	}
	return p
}

func (tg *teg) findById(id string) item {
	for it := range tg.Items() {
		if it.Id() == id {
			return it
		}
	}
	return nil
}

func (tg *teg) Construct(model *Teg) {
	tg.id = model.Id
	submodels := make(map[string]*teg, len(model.Groups))
	madePlaces := make(map[string]*place, 256)
	for _, g := range model.Groups {
		gNew := newGroup()
		gNew.id = g.Id
		gNew.label = g.Label
		gNew.folded = g.Folded
		gNew.parent = tg

		sub, ok := submodels[g.Model.Id]
		if !ok {
			sub = newTeg()
			sub.Construct(g.Model)
			sub.parent = tg
			submodels[g.Model.Id] = sub
		}

		gNew.model = sub
		gNew.Rect = geometry.NewRect(g.X, g.Y, 0, 0)
		for _, t := range g.Inputs {
			tNew := constructTransition(t)
			proxId, ok := g.Iostate[t.Id]
			if !ok {
				panic(errors.New("group constructing: iostate broken"))
			}
			tNew.group = gNew
			tNew.proxy = gNew.model.findById(proxId).(*transition)
			tNew.parent = tg
			for _, p := range t.In {
				pNew, ok := madePlaces[p.Id]
				if !ok {
					pNew = constructPlace(p)
					pNew.parent = tg
					madePlaces[p.Id] = pNew
					tg.places = append(tg.places, pNew)
				}
				pNew.out = tNew
				tNew.in = append(tNew.in, pNew)
				if p.OutControl != nil {
					pNew.outControl = newControlPoint(p.OutControl.X, p.OutControl.Y, p.OutControl.Modified)
				}
				pNew.refineControls()
			}
			for _, p := range t.Out {
				pInner := gNew.model.findById(p.Id).(*place)
				tNew.out = append(tNew.out, pInner)
			}
			tNew.refineSize()
			gNew.inputs = append(gNew.inputs, tNew)
		}
		for _, t := range g.Outputs {
			tNew := constructTransition(t)
			proxId := g.Iostate[t.Id]
			tNew.group = gNew
			tNew.proxy = gNew.model.findById(proxId).(*transition)
			tNew.parent = tg
			for _, p := range t.Out {
				pNew, ok := madePlaces[p.Id]
				if !ok {
					pNew = constructPlace(p)
					pNew.parent = tg
					madePlaces[p.Id] = pNew
					tg.places = append(tg.places, pNew)
				}
				pNew.in = tNew
				tNew.out = append(tNew.out, pNew)
				if p.InControl != nil {
					pNew.inControl = newControlPoint(p.InControl.X, p.InControl.Y, p.InControl.Modified)
				}
				pNew.refineControls()
			}
			for _, p := range t.In {
				pInner := gNew.model.findById(p.Id).(*place)
				tNew.in = append(tNew.in, pInner)
			}
			tNew.refineSize()
			gNew.outputs = append(gNew.outputs, tNew)
		}

		gNew.iostate = make(map[*transition]*transition, len(gNew.inputs)+len(gNew.outputs))
		for _, t := range gNew.inputs {
			if t.proxy != nil {
				gNew.iostate[t.proxy] = t
			}
		}
		for _, t := range gNew.outputs {
			if t.proxy != nil {
				gNew.iostate[t.proxy] = t
			}
		}

		tg.groups = append(tg.groups, gNew)
		gNew.updateBounds(false)
		gNew.updateIO()
		gNew.adjustIO()
		gNew.Align()
	}
	for _, t := range model.Transitions {
		tNew := constructTransition(t)
		tNew.parent = tg
		for _, p := range t.In {
			pNew, ok := madePlaces[p.Id]
			if !ok {
				pNew = constructPlace(p)
				pNew.parent = tg
				madePlaces[p.Id] = pNew
				tg.places = append(tg.places, pNew)
			}
			pNew.out = tNew
			tNew.in = append(tNew.in, pNew)
			if p.OutControl != nil {
				pNew.outControl = newControlPoint(p.OutControl.X, p.OutControl.Y, p.OutControl.Modified)
			}
			pNew.refineControls()
		}
		for _, p := range t.Out {
			pNew, ok := madePlaces[p.Id]
			if !ok {
				pNew = constructPlace(p)
				pNew.parent = tg
				madePlaces[p.Id] = pNew
				tg.places = append(tg.places, pNew)
			}
			pNew.in = tNew
			tNew.out = append(tNew.out, pNew)
			if p.InControl != nil {
				pNew.inControl = newControlPoint(p.InControl.X, p.InControl.Y, p.InControl.Modified)
			}
			pNew.refineControls()
		}
		tNew.refineSize()
		tg.transitions = append(tg.transitions, tNew)
	}
	for _, p := range model.Places {
		_, ok := madePlaces[p.Id]
		if !ok {
			pNew := constructPlace(p)
			pNew.parent = tg
			tg.places = append(tg.places, pNew)
		}
	}
}

func (tg *teg) UnmarshalJSON(data []byte) (err error) {
	m := &Teg{}
	if err = json.Unmarshal(data, m); err != nil {
		return
	}
	tg.Construct(m)
	return
}

type places []*place
type placesByX struct{ places }
type placesByY struct{ places }

func (p places) Len() int      { return len(p) }
func (p places) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p placesByX) Less(i, j int) bool {
	diff := p.places[i].Center().X - p.places[j].Center().X
	if math.Abs(diff) > 1 {
		return math.Signbit(diff)
	} else {
		return p.places[i].Center().Y > p.places[j].Center().Y
	}
}
func (p placesByY) Less(i, j int) bool {
	diff := p.places[i].Center().Y - p.places[j].Center().Y
	if math.Abs(diff) > 1 {
		return math.Signbit(diff)
	} else {
		return p.places[i].Center().X > p.places[j].Center().X
	}
}

type transitions []*transition
type transitionsByMedian struct {
	transitions
	inputs bool
}

func (t transitions) Len() int      { return len(t) }
func (t transitions) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t transitionsByMedian) Less(i, j int) bool {
	if t.inputs {
		h1, h2 := t.transitions[i].horizontal, t.transitions[j].horizontal
		i1, i2 := t.transitions[i].proxy.out, t.transitions[j].proxy.out
		diff := places(i2).calcMedian(h2) - places(i1).calcMedian(h1)
		if math.Abs(diff) > 1 {
			return !math.Signbit(diff)
		} else {
			return places(i2).calcMedian(!h2) < places(i1).calcMedian(!h1)
		}
	} else {
		h1, h2 := t.transitions[i].horizontal, t.transitions[j].horizontal
		o1, o2 := t.transitions[i].proxy.in, t.transitions[j].proxy.in
		diff := places(o2).calcMedian(h2) - places(o1).calcMedian(h1)
		if math.Abs(diff) > 1 {
			return math.Signbit(diff)
		} else {
			return places(o2).calcMedian(!h2) < places(o1).calcMedian(!h1)
		}
	}
}

func (p places) calcMedian(horizontal bool) (median float64) {
	if horizontal {
		for i := range p {
			median += p[i].X()
		}
	} else {
		for i := range p {
			median += p[i].Y()
		}
	}
	return median / float64(len(p))
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

func (p *place) refineControls() {
	if p.in != nil && p.inControl != nil && !p.inControl.modified {
		p.resetControlPoint(true)
	}
	if p.out != nil && p.outControl != nil && !p.outControl.modified {
		p.resetControlPoint(false)
	}
}

func (p *place) Shift(dx, dy float64) {
	p.Circle.Shift(dx, dy)
	p.shiftControls(dx, dy)
}

func (t *transition) Shift(dx, dy float64) {
	t.Rect.Shift(dx, dy)
	for _, p := range t.in {
		p.refineControls()
	}
	for _, p := range t.out {
		p.refineControls()
	}
}

func (g *group) Shift(dx, dy float64) {
	g.Rect.Shift(dx, dy)
	for _, t := range g.inputs {
		t.Shift(dx, dy)
	}
	for _, t := range g.outputs {
		t.Shift(dx, dy)
	}
}

func (p *place) Copy() item {
	clonemap := make(map[item]item, 1)
	pNew := newPlace(p.Center().X, p.Center().Y)
	pNew.label = p.label
	pNew.parent = p.parent
	pNew.counter = p.counter
	pNew.timer = p.timer
	clonemap[p] = pNew
	return item(pNew)
}

func (t *transition) Copy() item {
	clonemap := make(map[item]item, 1)
	tNew := newTransition(t.Center().X, t.Center().Y)
	tNew.label = t.label
	tNew.proxy = t.proxy
	tNew.group = t.group
	tNew.kind = t.kind
	if t.horizontal {
		tNew.rotate()
	}
	clonemap[t] = tNew
	return item(tNew)
}

func (g *group) Copy() item {
	gNew := newGroup()
	/*
		items := make(map[item]bool, len(g.model.places)+
			len(g.model.transitions)+len(g.model.groups))
		for _, p := range g.model.places {
			items[p] = true
		}
		for _, t := range g.model.transitions {
			items[t] = true
		}
		for _, g := range g.model.groups {
			items[g] = true
		}
		clonemap := group.model.cloneItems(items)
	*/
	gNew.Rect = geometry.NewRect(g.X(), g.Y(), 0, 0)
	gNew.model = g.model
	gNew.label = g.label
	gNew.parent = g.parent
	gNew.folded = g.folded
	return item(gNew)
}

func (t *transition) BorderPoint(inbound bool, index int) *geometry.Point {
	var count int
	if inbound {
		count = len(t.in)
	} else {
		count = len(t.out)
	}
	end := calcBorderPointTransition(t, pt(0, 0), inbound, count, index)
	if inbound {
		return pt(end.xTip, end.yTip)
	}
	return pt(end.x, end.y)
}

func (t *transition) Align() (float64, float64) {
	if t.proxy != nil {
		return 0, 0 //hack
	}
	x, y := math.Floor(t.Center().X), math.Floor(t.Center().Y)
	shiftX, shiftY := geometry.Align(x, y, GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	t.Move(x, y)
	t.Shift(shiftX, shiftY)
	for _, p := range t.in {
		p.refineControls()
	}
	for _, p := range t.out {
		p.refineControls()
	}
	return shiftX, shiftY
}

func (p *place) Align() (float64, float64) {
	x, y := math.Floor(p.Center().X), math.Floor(p.Center().Y)
	shiftX, shiftY := geometry.Align(x, y, GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	p.Move(x, y)
	p.Shift(shiftX, shiftY)
	return shiftX, shiftY
}

func (g *group) Align() (float64, float64) {
	x, y := math.Floor(g.Center().X), math.Floor(g.Center().Y)
	shiftX, shiftY := geometry.Align(x, y, GridDefaultGap)
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

func (g *group) Move(x, y float64) {
	g.Rect.Move(x-g.Width()/2, y-g.Height()/2)
}

func (c *controlPoint) Move(x, y float64) {
	c.Rect.Move(x-ControlPointWidth/2, y-ControlPointHeight/2)
}

func (t *transition) rotate() {
	t.horizontal = !t.horizontal
	t.Rect.Rotate(t.horizontal)
	t.OrderArcs(true)
	t.OrderArcs(false)
}

func (tg *teg) findProxy(t *transition) (proxy *transition, ok bool) {
	for _, g := range tg.groups {
		for _, t2 := range g.inputs {
			if t2.proxy == t {
				return t2, true
			}
		}
		for _, t2 := range g.outputs {
			if t2.proxy == t {
				return t2, true
			}
		}
	}
	if tg.parent != nil {
		return tg.parent.findProxy(t)
	}
	return nil, false
}

func (p *place) KindInGroup(items map[item]bool) int {
	inFound, outFound := false, false
	if p.in == nil {
		inFound = true
	} else if p.in.parent.parent != nil {
		if proxy, ok := p.in.parent.parent.findProxy(p.in); ok {
			if items[proxy.group] {
				inFound = true
			}
		}
	}
	if p.out == nil {
		outFound = true
	} else if p.out.parent.parent != nil {
		if proxy, ok := p.out.parent.parent.findProxy(p.out); ok {
			if items[proxy.group] {
				outFound = true
			}
		}
	}
	if !inFound || !outFound {
		for it := range items {
			if t, ok := it.(*transition); ok {
				if t == p.in {
					inFound = true
				}
				if t == p.out {
					outFound = true
				}
			}
		}
	}
	if !inFound || !outFound {
		return PlaceExposed
	}
	return PlaceInternal
}

func (g *group) KindInGroup(items map[item]bool) int {
	inFound, outFound := false, false
	if len(g.outputs) < 1 {
		outFound = true
	}
	if len(g.inputs) < 1 {
		inFound = true
	}
	if !outFound || !inFound {
		for it := range items {
			if p, ok := it.(*place); ok {
				for _, t := range g.inputs {
					if t == p.out || t.proxy == p.out {
						inFound = true
					}
				}
				for _, t := range g.outputs {
					if t == p.in || t.proxy == p.in {
						outFound = true
					}
				}
			}
		}
	}
	if !outFound || !inFound {
		return GroupExposed
	}
	return GroupInternal
}

func (t *transition) KindInGroup(items map[item]bool) int {
	if t.proxy != nil {
		return TransitionInternal
	}
	var interlinkIn, interlinkOut, outlinkIn, outlinkOut bool
	for _, p := range t.in {
		found := false
		for it := range items {
			if p2, ok := it.(*place); ok {
				if p2 == p {
					found = true
					interlinkIn = true
				}
			}
		}
		outlinkIn = outlinkIn || !found
	}
	for _, p := range t.out {
		found := false
		for it := range items {
			if p2, ok := it.(*place); ok {
				if p2 == p {
					found = true
					interlinkOut = true
				}
			}
		}
		outlinkOut = outlinkOut || !found
	}
	switch {
	case !interlinkIn && (outlinkIn || interlinkOut) && !outlinkOut:
		return TransitionInput
	case !interlinkOut && (outlinkOut || interlinkIn) && !outlinkIn:
		return TransitionOutput
	case !(outlinkIn || outlinkOut):
		return TransitionInternal
	}
	return TransitionExposed
}

func (t *transition) Has(x, y float64) bool {
	var radius float64
	if t.horizontal {
		radius = (t.Width() + 6.0) / 2
	} else {
		radius = (t.Height() + 6.0) / 2
	}
	return math.Pow(x-t.Center().X, 2)+math.Pow(y-t.Center().Y, 2) < math.Pow(radius, 2)
}

func (t *transition) Bound() *geometry.Rect {
	return t.Rect
}

func (g *group) Bound() *geometry.Rect {
	return g.Rect
}

func (t *transition) resetProperties() {
	t.label = ""
	if t.horizontal {
		t.rotate()
	}
}

func (g *group) resetProperties() {
	g.label = ""
	for _, t := range g.inputs {
		if t.horizontal {
			t.rotate()
		}
	}
	for _, t := range g.outputs {
		if t.horizontal {
			t.rotate()
		}
	}
	g.updateBounds(false)
	g.adjustIO()
}

func (p *place) resetProperties() {
	p.label = ""
	if p.in != nil {
		p.resetControlPoint(true)
	}
	if p.out != nil {
		p.resetControlPoint(false)
	}
}

func (t *transition) newControlPoint(center *geometry.Point, inbound bool) *controlPoint {
	circle := geometry.NewCircle(center.X, center.Y, PlaceRadius)
	if inbound {
		centerT := t.Center()
		point := circle.BorderPoint(centerT.X, centerT.Y, 15.0)
		return newControlPoint(point.X, point.Y, false)
	} else {
		centerT := t.Center()
		point := circle.BorderPoint(centerT.X, centerT.Y, 15.0)
		return newControlPoint(point.X, point.Y, false)
	}
}

func (p *place) resetControlPoint(inbound bool) {
	if inbound {
		centerT := p.in.Center()
		point := p.BorderPoint(centerT.X, centerT.Y, 15.0)
		p.inControl = newControlPoint(point.X, point.Y, false)
	} else {
		centerT := p.out.Center()
		point := p.BorderPoint(centerT.X, centerT.Y, 15.0)
		p.outControl = newControlPoint(point.X, point.Y, false)
	}
}

func (p *place) Id() string {
	return p.id
}

func (t *transition) Id() string {
	return t.id
}

func (g *group) Id() string {
	return g.id
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

type teg struct {
	parent      *teg
	util        *utility
	places      []*place
	transitions []*transition
	groups      []*group
	selected    map[item]bool
	infos       map[string]*planeview.Plane
	updated     chan interface{}
	updatedInfo chan interface{}
	id          string
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

func calcItemsShift(center *geometry.Point, items map[item]bool) *geometry.Point {
	x0, y0, x1, y1 := detectBounds(items)
	itemsCenter := pt(x0+(x1-x0)/2, y0+(y1-y0)/2)
	return pt(center.X-itemsCenter.X, center.Y-itemsCenter.Y)
}

func bound(it drawable,
	x0 float64, y0 float64, x1 float64, y1 float64) (mx0, my0, mx1, my1 float64) {
	mx0, my0, mx1, my1 = x0, y0, x1, y1
	x, y, w, h := it.X(), it.Y(), it.Width(), it.Height()
	if x < x0 {
		mx0 = x
	}
	if y < y0 {
		my0 = y
	}
	if x+w > x1 {
		mx1 = x + w
	}
	if y+h > y1 {
		my1 = y + h
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
		if t.proxy != nil {
			continue
		}
		items[t] = true
	}
	for _, g := range tg.groups {
		items[g] = true
	}
	return items
}

func (tg *teg) update() {
	tg.updated <- nil
	tg.updateInfos()
}

func (t *teg) updateParentGroups() {
	if t.parent != nil {
		for _, g := range t.parent.groups {
			folded := g.folded
			if !folded {
				t.parent.foldGroup(g)
			}
			g.updateBounds(false)
			g.updateIO()
			g.adjustIO()
			if !folded {
				t.parent.unfoldGroup(g)
			}
		}
		t.parent.update()
		t.parent.updateParentGroups()
	}
}

func (g *group) updateIO() {
	if g.model == nil {
		return
	}

	g.inputs = make([]*transition, 0, len(g.model.transitions))
	g.outputs = make([]*transition, 0, len(g.model.transitions))

	for _, t := range g.model.transitions {
		kind := t.kind
		if kind == TransitionInternal {
			// try to guess
			kind = t.KindInGroup(g.model.Items())
			if kind == TransitionInternal {
				continue
			}
		}
		c := t.Copy().(*transition)
		c.proxy = t
		c.group = g
		c.parent = g.parent
		if old, ok := g.iostate[t]; ok {
			switch kind {
			case TransitionInput:
				for _, p := range old.in {
					c.in = append(c.in, p)
					p.out = c
				}
				for _, p := range t.out {
					c.out = append(c.out, p)
				}
			case TransitionOutput:
				for _, p := range old.out {
					c.out = append(c.out, p)
					p.in = c
				}
				for _, p := range t.in {
					c.in = append(c.in, p)
				}
			}
		} else {
			in := make(map[*place]bool, len(t.in))
			out := make(map[*place]bool, len(t.out))
			for _, p := range t.in {
				c.in = append(c.in, p)
				in[p] = true
			}
			for _, p := range t.out {
				c.out = append(c.out, p)
				out[p] = true
			}
			switch kind {
			case TransitionInput:
				for p := range in {
					t.unlink(p, true, true)
					p.out = c
					p.resetControlPoint(false)
				}
			case TransitionOutput:
				for p := range out {
					t.unlink(p, false, true)
					p.in = c
					p.resetControlPoint(true)
				}
			}
		}
		switch kind {
		case TransitionInput:
			t.kind = TransitionInput
			c.kind = TransitionOutput // for parent tg
			g.inputs = append(g.inputs, c)
		case TransitionOutput:
			t.kind = TransitionOutput
			c.kind = TransitionInput // for parent tg
			g.outputs = append(g.outputs, c)
		}
		if old, ok := g.iostate[t]; ok {
			if old.horizontal != c.horizontal {
				c.rotate()
			}
			if len(old.label) > 0 {
				c.label = old.label
			}
		}
	}

	g.iostate = make(map[*transition]*transition, len(g.inputs)+len(g.outputs))

	for _, t := range g.inputs {
		if t.proxy != nil {
			g.iostate[t.proxy] = t
		}
	}
	for _, t := range g.outputs {
		if t.proxy != nil {
			g.iostate[t.proxy] = t
		}
	}
}

func (g *group) adjustIO() {
	if g.model == nil {
		return
	}
	sort.Sort(transitionsByMedian{g.inputs, true})
	sort.Sort(transitionsByMedian{g.outputs, false})
	var offi, offj float64
	for _, t := range g.inputs {
		t.refineSize()
		if t.horizontal {
			base := g.X() + GroupMargin
			t.Move(base+offi+t.Width()/2, g.Y())
			offi += t.Width() + GroupIOSpacing
		} else {
			base := g.Y() + GroupMargin
			t.Move(g.X(), base+offj+t.Height()/2)
			offj += t.Height() + GroupIOSpacing
		}
		for _, p := range t.in {
			p.refineControls()
		}
	}
	offi, offj = 0.0, 0.0
	for _, t := range g.outputs {
		t.refineSize()
		if t.horizontal {
			base := g.X() + g.Width() - GroupMargin
			t.Move(base-offi-t.Width()/2, g.Y()+g.Height())
			offi += t.Width() + GroupIOSpacing
		} else {
			base := g.Y() + g.Height() - GroupMargin
			t.Move(g.X()+g.Width(), base-offj-t.Height()/2)
			offj += t.Height() + GroupIOSpacing
		}
		for _, p := range t.out {
			p.refineControls()
		}
	}
}

func (g *group) getFoldedSize() (w, h float64) {
	var inWidth, outWidth, inHeight, outHeight float64
	for _, t := range g.inputs {
		t.refineSize()
		if t.horizontal {
			inWidth += t.Width() + GroupIOSpacing
		} else {
			inHeight += t.Height() + GroupIOSpacing
		}
	}
	for _, t := range g.outputs {
		t.refineSize()
		if t.horizontal {
			outWidth += t.Width() + GroupIOSpacing
		} else {
			outHeight += t.Height() + GroupIOSpacing
		}
	}
	inWidth -= GroupIOSpacing
	inHeight -= GroupIOSpacing
	outWidth -= GroupIOSpacing
	outHeight -= GroupIOSpacing
	w, h = math.Max(math.Max(inWidth, outWidth), GroupMinSize)+GroupMargin*2,
		math.Max(math.Max(inHeight, outHeight), GroupMinSize)+GroupMargin*2
	return
}

func (g *group) updateBounds(move bool) {
	if g.model == nil {
		return
	}
	var w, h float64
	x0, y0, x1, y1 := detectBounds(g.model.Items())
	if !g.folded {
		w, h = (x1-x0)+2*GroupMargin, (y1-y0)+2*GroupMargin
		k := math.Ceil(w / GridDefaultGap)
		w = k * GridDefaultGap
		k = math.Ceil(h / GridDefaultGap)
		h = k * GridDefaultGap
	}
	fw, fh := g.getFoldedSize()
	if w < fw {
		w = fw
	}
	if h < fh {
		h = fh
	}
	if move {
		x, y := x0-GroupMargin, y0-GroupMargin
		g.Rect = geometry.NewRect(x, y, w, h)
	} else {
		g.Rect.Resize(w, h)
	}
}

func (tg *teg) addPlace(x, y float64) *place {
	place := newPlace(x, y)
	place.parent = tg
	tg.places = append(tg.places, place)
	return place
}

func (tg *teg) addGroup(items map[item]bool) *group {
	group := newGroup()
	group.parent = tg
	group.model.parent = tg
	group.model.transferItems(tg, items)
	group.updateBounds(true)
	shift := calcItemsShift(pt(0, 0), items)
	for it := range items {
		it.Shift(shift.X, shift.Y)
		it.Align()
	}
	tg.groups = append(tg.groups, group)
	return group
}

type linkPair struct {
	t  *transition
	p  *place
	in bool
}

func (tg *teg) flatGroup(g *group) map[item]bool {
	items := make(map[item]bool, len(g.model.Items()))
	clones := tg.cloneItems(g.model.Items())
	for _, it := range clones {
		items[it] = true
		if t, ok := it.(*transition); ok {
			t.kind = TransitionInternal
		}
	}
	toClean := make(map[*transition]bool)
	toLink := make([]linkPair, 0, 32)
	for _, t := range g.inputs {
		if t.proxy != nil {
			for _, p := range t.in {
				p.out = nil
				p.outControl = nil
				inner := clones[t.proxy].(*transition)
				toClean[inner] = true
				toLink = append(toLink, linkPair{t: inner, p: p, in: true})
			}
			t.in = nil
		}
	}
	for _, t := range g.outputs {
		if t.proxy != nil {
			for _, p := range t.out {
				p.in = nil
				p.inControl = nil
				inner := clones[t.proxy].(*transition)
				toClean[inner] = false
				toLink = append(toLink, linkPair{t: inner, p: p, in: false})
			}
			t.out = nil
		}
	}
	for t, inbound := range toClean {
		if inbound {
			t.in = make([]*place, 0, 8)
		} else {
			t.out = make([]*place, 0, 8)
		}
	}
	for _, pair := range toLink {
		pair.t.link(pair.p, pair.in)
	}

	shift := calcItemsShift(g.Center(), items)
	for it := range items {
		it.Shift(shift.X, shift.Y)
	}
	tg.removeGroup(g)
	return items
}

func (tg *teg) foldGroup(g *group) {
	if g.folded {
		return
	}
	w1, h1 := g.Width(), g.Height()
	g.folded = true
	g.updateBounds(false)
	g.adjustIO()
	w2, h2 := g.Width(), g.Height()
	items := tg.Items()
	delete(items, g)
	for it := range items {
		if (it.Y()+it.Height() >= g.Y()) && (it.Y() <= g.Y()+h1) && (it.X() >= g.X()+w1) {
			it.Shift(w2-w1, 0)
		}
		if (it.X()+it.Width() >= g.X()) && (it.X() <= g.X()+w1) && (it.Y() >= g.Y()+h1) {
			it.Shift(0, h2-h1)
		}
	}
}

func (tg *teg) unfoldGroup(g *group) {
	if !g.folded {
		return
	}
	w1, h1 := g.Width(), g.Height()
	g.folded = false
	g.updateBounds(false)
	g.adjustIO()
	w2, h2 := g.Width(), g.Height()
	items := tg.Items()
	delete(items, g)
	for it := range items {
		if (it.X()+it.Width() >= g.X()) && (it.X() <= g.X()+w2) && (it.Y()+(h2-h1) >= g.Y()+h2) {
			it.Shift(0, h2-h1)
		}
		if (it.Y()+it.Height() >= g.Y()) && (it.Y() <= g.Y()+h2) && (it.X()+(w2-w1) >= g.X()+w2) {
			it.Shift(w2-w1, 0)
		}
	}
}

func (tg *teg) removePlace(p *place) {
	if p.in != nil {
		p.in.unlink(p, false, false)

	}
	if p.out != nil {
		p.out.unlink(p, true, false)
	}
	for i, place := range tg.places {
		if place == p {
			tg.places = append(tg.places[:i], tg.places[i+1:]...)
			return
		}
	}
}

func (tg *teg) addTransition(x, y float64) *transition {
	transition := newTransition(x, y)
	transition.parent = tg
	tg.transitions = append(tg.transitions, transition)
	return transition
}

func (tg *teg) removeTransition(t *transition) {
	if t.proxy != nil {
		return
	}
	if proxy, ok := tg.findProxy(t); ok {
		if proxy.kind == TransitionOutput {
			for _, p := range proxy.in {
				p.out = nil
				p.outControl = nil
			}
		} else if proxy.kind == TransitionInput {
			for _, p := range proxy.out {
				p.in = nil
				p.inControl = nil
			}
		}
	}
	for _, p := range t.in {
		p.out = nil
		p.outControl = nil
	}
	for _, p := range t.out {
		p.in = nil
		p.inControl = nil
	}
	for i, t2 := range tg.transitions {
		if t2 == t {
			tg.transitions = append(tg.transitions[:i], tg.transitions[i+1:]...)
			return
		}
	}
}

func (tg *teg) removeGroup(g *group) {
	for _, t := range g.inputs {
		for _, p := range t.in {
			p.out = nil
			p.outControl = nil
		}
	}
	for _, t := range g.outputs {
		for _, p := range t.out {
			p.in = nil
			p.inControl = nil
		}
	}
	g.inputs = nil
	g.outputs = nil
	g.model = nil
	g.parent = nil
	for i, gg := range tg.groups {
		if gg == g {
			tg.groups = append(tg.groups[:i], tg.groups[i+1:]...)
			return
		}
	}
}

func (tg *teg) fakeData() {
	p1 := tg.addPlace(-138.863281, -96.941406)
	p1.timer, p1.counter = 8, 0
	p2 := tg.addPlace(-138.468750, 2.511719)
	p2.timer, p2.counter = 3, 0
	p3 := tg.addPlace(67.714844, -56.714844)
	p3.timer, p3.counter = 5, 3
	p4 := tg.addPlace(-61.046875, -56.257812)
	p4.timer, p4.counter = 3, 4
	p5 := tg.addPlace(4.554688, 149.593750)
	p5.timer, p5.counter = 2, 2
	p5.label = "delayed\ncycle"
	p6 := tg.addPlace(155.371094, -63.152344)
	p7 := tg.addPlace(118.007812, 8.843750)

	t1 := tg.addTransition(-265.976562, -46.679688)
	t1.label = "input sink"
	t2 := tg.addTransition(6.859375, 29.503906)
	t2.label = "x1"
	t3 := tg.addTransition(7.425781, -145.574219)
	t3.label = "x2"
	t4 := tg.addTransition(253.656250, -35.816406)
	t4.label = "output sink"

	t1.link(p1, false)
	t1.link(p2, false)
	t2.link(p2, true)
	t2.link(p3, true)
	t2.link(p4, false)
	t2.link(p5, true)
	t2.link(p5, false)
	t2.link(p7, false)
	t3.link(p1, true)
	t3.link(p3, false)
	t3.link(p4, true)
	t3.link(p6, false)
	t4.link(p6, true)
	t4.link(p7, true)

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

	for _, p := range tg.places {
		p.Align()
	}
	for _, t := range tg.transitions {
		t.Align()
	}
}

func (tg *teg) deselectAll() {
	for k := range tg.selected {
		delete(tg.selected, k)
	}
}

func (tg *teg) deselectItem(it item) {
	delete(tg.selected, it)
}

func (tg *teg) selectItem(it item) {
	if it != nil {
		tg.selected[it] = true
	}
}

func (p *place) IsSelected() bool {
	if p.parent != nil {
		return p.parent.isSelected(p)
	}
	return false
}

func (t *transition) IsSelected() bool {
	var selected bool
	if t.parent != nil {
		selected = t.parent.isSelected(t)
	}
	return selected
}

func (g *group) IsSelected() bool {
	if g.parent != nil {
		return g.parent.isSelected(g)
	}
	return false
}

func (tg *teg) isSelected(it item) bool {
	_, ok := tg.selected[it]
	return ok
}

func (t *transition) nextKind() {
	if t.proxy != nil {
		return
	}
	in := len(t.in) > 0
	out := len(t.out) > 0
	switch t.kind {
	case TransitionInternal:
		if !in {
			t.kind = TransitionInput
		} else if !out {
			t.kind = TransitionOutput
		}
	case TransitionInput:
		if !out {
			t.kind = TransitionOutput
		} else {
			t.kind = TransitionInternal
		}
	case TransitionOutput:
		if !in && !out {
			t.kind = TransitionInternal
		} else if !in {
			t.kind = TransitionInput
		} else {
			t.kind = TransitionInternal
		}
	}
}

func (t *transition) isLinked(p *place, inbound bool) bool {
	if inbound {
		return p.out == t
	}
	return p.in == t
}

func (t *transition) refineSize() {
	if t.horizontal {
		w := calcTransitionHeight(len(t.in), len(t.out))
		t.Resize(w, TransitionWidth) // rotated
	} else {
		h := calcTransitionHeight(len(t.in), len(t.out))
		t.Resize(TransitionWidth, h) // normal
	}
}

func (t *transition) link(p *place, inbound bool) {
	if inbound && t.kind == TransitionInput {
		return
	} else if !inbound && t.kind == TransitionOutput {
		return
	}
	if t.isLinked(p, inbound) {
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
		t.refineSize()
		if t.proxy != nil {
			t.group.updateIO()
			t.group.adjustIO()
		}
	}
}

func (t *transition) unlink(p *place, inbound bool, blitz bool) {
	if !t.isLinked(p, inbound) {
		return
	}
	var changed bool
	if inbound && p.out != nil {
		p.out = nil
		for i, it := range t.in {
			if it == p {
				t.in = append(t.in[:i], t.in[i+1:]...)
				break
			}
		}
		t.OrderArcs(true)
		p.outControl = nil
		changed = true
	} else if !inbound && p.in != nil {
		p.in = nil
		for i, it := range t.out {
			if it == p {
				t.out = append(t.out[:i], t.out[i+1:]...)
				break
			}
		}
		t.OrderArcs(false)
		p.inControl = nil
		changed = true
	}
	if changed {
		t.refineSize()
		if !blitz && t.proxy != nil {
			t.group.updateIO()
			t.group.adjustIO()
		}
	}
}

func calcTransitionHeight(in, out int) float64 {
	return math.Max(math.Max(float64(in)*TransitionHeight/2.0,
		float64(out)*TransitionHeight/2.0), TransitionHeight)
}

func (tg *teg) transferItems(tg2 *teg, items map[item]bool) {
	var movP, movT, movG bool
	// copy refs to tg
	for it := range items {
		if p, ok := it.(*place); ok {
			p.parent = tg
			tg.places = append(tg.places, p)
			movP = true
		} else if t, ok := it.(*transition); ok {
			t.parent = tg
			tg.transitions = append(tg.transitions, t)
			movT = true
		} else if g, ok := it.(*group); ok {
			g.parent = tg
			g.model.parent = tg
			tg.groups = append(tg.groups, g)
			movG = true
		}
	}
	// cleanup refs in tg2
	if movP {
		places := make([]*place, 0, 256)
		for _, p := range tg2.places {
			if !items[p] {
				places = append(places, p)
			}
		}
		tg2.places = places
	}
	if movT {
		transitions := make([]*transition, 0, 256)
		for _, t := range tg2.transitions {
			if !items[t] {
				transitions = append(transitions, t)
			}
		}
		tg2.transitions = transitions
	}
	if movG {
		groups := make([]*group, 0, 32)
		for _, g := range tg2.groups {
			if !items[g] {
				groups = append(groups, g)
			}
		}
		tg2.groups = groups
	}
}

func (tg *teg) cloneItems(items map[item]bool) (clones map[item]item) {
	clones = make(map[item]item, len(items))
	ignore := make(map[item]bool, len(items))
	for it := range items {
		if t, ok := it.(*transition); ok {
			if t.proxy != nil {
				ignore[t] = true
				continue
			}
			tNew := t.Copy().(*transition)
			tNew.parent = tg
			clones[t] = tNew
			tg.transitions = append(tg.transitions, tNew)
			for _, p := range t.in {
				if _, ok := items[p]; ok {
					var pNew *place
					if pNew, ok = clones[p].(*place); !ok {
						pNew = p.Copy().(*place)
						pNew.parent = tg
						clones[p] = pNew
						tg.places = append(tg.places, pNew)
					}
					tNew.link(pNew, true)
					if p.outControl.modified {
						pNew.outControl.Move(p.outControl.Center().X, p.outControl.Center().Y)
						pNew.outControl.modified = true
					}
				}
			}
			for _, p := range t.out {
				if _, ok := items[p]; ok {
					var pNew *place
					if pNew, ok = clones[p].(*place); !ok {
						pNew = p.Copy().(*place)
						pNew.parent = tg
						clones[p] = pNew
						tg.places = append(tg.places, pNew)
					}
					tNew.link(pNew, false)
					if p.inControl.modified {
						pNew.inControl.Move(p.inControl.Center().X, p.inControl.Center().Y)
						pNew.inControl.modified = true
					}
				}
			}
			// adjust position (size changed)
			tNew.Move(t.Center().X, t.Center().Y)
		}
	}
	for k := range ignore {
		delete(items, k)
	}
	for k := range clones {
		delete(items, k)
	}
	for it := range items {
		if p, ok := it.(*place); ok {
			if pNew, ok := clones[p].(*place); !ok {
				pNew = p.Copy().(*place)
				pNew.parent = tg
				clones[p] = pNew
				tg.places = append(tg.places, pNew)
			}
		}
	}
	for k := range clones {
		delete(items, k)
	}
	for it := range items {
		if g, ok := it.(*group); ok {
			gNew := g.Copy().(*group)
			gNew.parent = tg
			clones[g] = gNew
			tg.groups = append(tg.groups, gNew)
			gNew.updateIO()

			for _, t := range g.inputs {
				for _, p := range t.in {
					if newPlace, ok := clones[p].(*place); ok {
						proxy := t.proxy
						for _, t2 := range gNew.inputs {
							if t2.proxy == proxy {
								t2.link(newPlace, true)
							}
						}
					}
				}
			}
			for _, t := range g.outputs {
				for _, p := range t.out {
					if newPlace, ok := clones[p].(*place); ok {
						proxy := t.proxy
						for _, t2 := range gNew.outputs {
							if t2.proxy == proxy {
								t2.link(newPlace, false)
							}
						}
					}
				}
			}
			for _, t := range gNew.inputs {
				if old, ok := g.iostate[t.proxy]; ok {
					t.label = old.label
					if old.horizontal != t.horizontal {
						t.rotate()
					}
				}
			}
			gNew.updateBounds(false)
			gNew.adjustIO()
		}
	}
	return
}

func (tg *teg) findDrawable(x float64, y float64) (interface{}, bool) {
	for _, p := range tg.places {
		if tg.isSelected(p) {
			if p.in != nil && p.inControl.Has(x, y) {
				return p.inControl, true
			} else if p.out != nil && p.outControl.Has(x, y) {
				return p.outControl, true
			}
		}
	}
	for _, p := range tg.places {
		if p.Has(x, y) {
			return p, true
		}
	}
	for _, t := range tg.transitions {
		if t.Has(x, y) {
			return t, true
		}
	}
	for _, g := range tg.groups {
		for _, t := range g.inputs {
			if t.Has(x, y) {
				return t, true
			}
		}
		for _, t := range g.outputs {
			if t.Has(x, y) {
				return t, true
			}
		}
		if g.Has(x, y) {
			return g, true
		}
	}
	return nil, false
}

func (tg *teg) updateInfos() {
	var updated bool
	for id, info := range tg.infos {
		it := tg.findById(id)
		if t, ok := it.(*transition); !ok || t.kind == TransitionInternal {
			delete(tg.infos, id)
			updated = true
		} else if ok {
			if t.kind == TransitionInput && !info.IsInput() {
				delete(tg.infos, id)
				updated = true
			} else if t.kind == TransitionOutput && info.IsInput() {
				delete(tg.infos, id)
				updated = true
			}
		}
	}
	var k int
	for _, t := range tg.transitions {
		if t.kind == TransitionInput {
			k++
			if info, ok := tg.infos[t.id]; ok {
				label := t.label
				if len(t.label) < 1 {
					label = fmt.Sprintf("unnamed %d", k)
				}
				if info.Label() != label {
					info.SetLabel(label)
					updated = true
				}
			} else {
				label := t.label
				if len(label) < 1 {
					label = fmt.Sprintf("unnamed %d", k)
				}
				plane := planeview.NewPlane(t.id, label, true)
				plane.FakeData()
				plane.SetColor(PlaneColors[len(tg.infos)%9])
				tg.infos[t.id] = plane
				updated = true
			}
		}
	}
	if updated {
		tg.updatedInfo <- nil
	}
}

func newTeg() *teg {
	return &teg{
		util:        &utility{},
		infos:       make(map[string]*planeview.Plane, 20),
		places:      make([]*place, 0, 256),
		transitions: make([]*transition, 0, 256),
		groups:      make([]*group, 0, 32),
		selected:    make(map[item]bool, 256),
		updated:     make(chan interface{}, 100),
		updatedInfo: make(chan interface{}, 100),
		id:          util.GenUUID(),
	}
}
