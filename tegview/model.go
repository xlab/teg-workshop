package tegview

import (
	"math"
	"sort"

	"github.com/xlab/teg-workshop/geometry"
)

const (
	TransitionWidth    = 6.0
	TransitionHeight   = 30.0
	GroupMargin        = 10.0
	GroupIOSpacing     = 10.0
	GroupMinSize       = 150.0
	ControlPointWidth  = 10.0
	ControlPointHeight = 10.0
	PlaceRadius        = 25.0
	GridDefaultGap     = 16.0
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
	Copy() (item, map[item]item)
	IsSelected() bool
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
	parent     *teg
}

type transition struct {
	*geometry.Rect
	in         []*place
	out        []*place
	proxy      *transition
	label      string
	horizontal bool
	parent     *teg
	kind       int
}

type group struct {
	*geometry.Rect
	model   *teg
	inputs  []*transition
	outputs []*transition
	label   string
	folded  bool
	parent  *teg
}

type utility struct {
	min, max *geometry.Point
	kind     int
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
		model:   newTeg(),
		inputs:  make([]*transition, 0, 8),
		outputs: make([]*transition, 0, 8),
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
	if p.in != nil && !p.inControl.modified {
		p.resetControlPoint(true)
	}
	if p.out != nil && !p.outControl.modified {
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
	for it := range g.model.Items() {
		it.Shift(dx, dy)
	}
	for _, t := range g.inputs {
		t.Shift(dx, dy)
	}
	for _, t := range g.outputs {
		t.Shift(dx, dy)
	}
}

func (p *place) Copy() (item, map[item]item) {
	clonemap := make(map[item]item, 1)
	place := newPlace(p.Center().X, p.Center().Y)
	place.label = p.label
	place.parent = p.parent
	place.counter = p.counter
	place.timer = p.timer
	clonemap[p] = place
	return item(place), clonemap
}

func (t *transition) Copy() (item, map[item]item) {
	clonemap := make(map[item]item, 1)
	transition := newTransition(t.Center().X, t.Center().Y)
	transition.label = t.label
	transition.proxy = t.proxy
	transition.kind = t.kind
	if t.horizontal {
		transition.Rotate()
	}
	clonemap[t] = transition
	return item(transition), clonemap
}

func (g *group) Copy() (item, map[item]item) {
	group := newGroup()
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
	group.label = g.label
	group.parent = g.parent
	group.folded = g.folded
	group.updateIO()
	group.adjustIO()
	return item(group), clonemap
}

func (t *transition) BorderPoint(inbound bool, index int) *geometry.Point {
	var count int
	if inbound {
		count = len(t.in)
	} else {
		count = len(t.out)
	}
	end := calcBorderPointTransition(t, inbound, count, index)
	if inbound {
		return pt(end.xTip, end.yTip)
	}
	return pt(end.x, end.y)
}

func (t *transition) Align() (float64, float64) {
	if t.proxy != nil {
		return 0, 0 //hack
	}
	x, y := t.Center().X, t.Center().Y
	shiftX, shiftY := geometry.Align(x, y, GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	//t.Move(x, y)
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
	x, y := p.Center().X, p.Center().Y
	shiftX, shiftY := geometry.Align(x, y, GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	//p.Move(x, y)
	p.Shift(shiftX, shiftY)
	return shiftX, shiftY
}

func (g *group) Align() (float64, float64) {
	x, y := g.Center().X, g.Center().Y
	shiftX, shiftY := geometry.Align(x, y, GridDefaultGap)
	if shiftX == 0 && shiftY == 0 {
		return 0, 0
	}
	//g.Move(x, y)
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

func (t *transition) Rotate() {
	t.horizontal = !t.horizontal
	t.Rect.Rotate(t.horizontal)
	t.OrderArcs(true)
	t.OrderArcs(false)
}

func (p *place) KindInGroup(items map[item]bool) int {
	inFound, outFound := false, false
	if p.in == nil {
		inFound = true
	} else if p.in.parent.parent != nil {
		if items[p.in.parent.parent] {
			inFound = true
		}
	}
	if p.out == nil {
		outFound = true
	} else if p.out.parent.parent != nil {
		if items[p.out.parent.parent] {
			outFound = true
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
		t.Rotate()
	}
}

func (g *group) resetProperties() {
	g.label = ""
	g.folded = false
	for _, t := range g.inputs {
		if t.horizontal {
			t.Rotate()
		}
	}
	for _, t := range g.outputs {
		if t.horizontal {
			t.Rotate()
		}
	}
	g.adjustIO()
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

func (t *transition) findProxy(inbound bool) *transition {
	g := t.parent.parent
	if g == nil {
		return nil
	}
	if inbound {
		for _, tt := range g.inputs {
			if tt.proxy == t {
				return tt
			}
		}
	} else {
		for _, tt := range g.outputs {
			if tt.proxy == t {
				return tt
			}
		}
	}
	return nil
}

func (p *place) resetControlPoint(inbound bool) {
	if inbound {
		t := p.in
		proxy := t.findProxy(false)
		if proxy != nil {
			t = proxy
		}
		centerT := t.Center()
		point := p.BorderPoint(centerT.X, centerT.Y, 15.0)
		p.inControl = newControlPoint(point.X, point.Y, false)
	} else {
		t := p.out
		proxy := t.findProxy(true)
		if proxy != nil {
			t = proxy
		}
		centerT := t.Center()
		point := p.BorderPoint(centerT.X, centerT.Y, 15.0)
		p.outControl = newControlPoint(point.X, point.Y, false)
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

type teg struct {
	util        *utility
	places      []*place
	transitions []*transition
	groups      []*group
	inputs      []*transition
	outputs     []*transition
	parent      *group
	selected    map[item]bool
	updated     chan interface{}
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

func (tg *teg) update() {
	tg.updated <- nil
}

func (g *group) updateIO() {
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
		it, _ := t.Copy()
		c := it.(*transition)
		c.proxy = t
		for _, p := range t.in {
			c.in = append(c.in, p)
		}
		for _, p := range t.out {
			c.out = append(c.out, p)
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
	}
}

func (g *group) adjustIO() {
	g.updateBounds()
	sort.Sort(transitionsByMedian{g.inputs, true})
	sort.Sort(transitionsByMedian{g.outputs, false})
	var offi, offj float64
	for _, t := range g.inputs {
		if t.horizontal {
			base := g.X() + GroupMargin
			t.Move(base+offi+t.Width()/2, g.Y())
			offi += t.Width() + GroupIOSpacing
		} else {
			base := g.Y() + GroupMargin
			t.Move(g.X(), base+offj+t.Height()/2)
			offj += t.Height() + GroupIOSpacing
		}
	}
	offi, offj = 0.0, 0.0
	for _, t := range g.outputs {
		if t.horizontal {
			base := g.X() + g.Width() - GroupMargin
			t.Move(base-offi-t.Width()/2, g.Y()+g.Height())
			offi += t.Width() + GroupIOSpacing
		} else {
			base := g.Y() + g.Height() - GroupMargin
			t.Move(g.X()+g.Width(), base-offj-t.Height()/2)
			offj += t.Height() + GroupIOSpacing
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

func (g *group) updateBounds() {
	var w, h float64
	x0, y0, x1, y1 := detectBounds(g.model.Items())
	x, y := x0-GroupMargin, y0-GroupMargin
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
	g.Rect = geometry.NewRect(x, y, w, h)
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
	group.model.parent = group
	group.model.transferItems(tg, items)
	tg.groups = append(tg.groups, group)
	return group
}

func (tg *teg) flatGroup(g *group) map[item]bool {
	items := g.model.Items()
	tg.transferItems(g.model, items)
	for it := range items {
		if t, ok := it.(*transition); ok {
			t.kind = TransitionInternal
		}
	}
	tg.removeGroup(g)
	return items
}

func (t *teg) foldGroup(g *group) {
	if g.folded {
		return
	}
	w1, h1 := g.Width(), g.Height()
	g.folded = true
	g.adjustIO()
	w2, h2 := g.Width(), g.Height()
	items := t.Items()
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

func (t *teg) unfoldGroup(g *group) {
	if !g.folded {
		return
	}
	w1, h1 := g.Width(), g.Height()
	g.folded = false
	g.adjustIO()
	w2, h2 := g.Width(), g.Height()
	items := t.Items()
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
		g := p.in.parent.parent
		if g != nil {
			for _, t := range g.outputs {
				if t.proxy == p.in {
					t.unlink(p, false)
				}
			}
		} else {
			p.in.unlink(p, false)
		}
	}
	if p.out != nil {
		g := p.out.parent.parent
		if g != nil {
			for _, t := range g.inputs {
				if t.proxy == p.out {
					t.unlink(p, true)
				}
			}
		} else {
			p.out.unlink(p, true)
		}
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
	for _, p := range t.in {
		p.out = nil
	}
	for _, p := range t.out {
		p.in = nil
	}
	for i, t2 := range tg.transitions {
		if t2 == t {
			tg.transitions = append(tg.transitions[:i], tg.transitions[i+1:]...)
			return
		}
	}
}

func (tg *teg) removeGroup(g *group) {
	g.inputs = nil
	g.outputs = nil
	for _, t := range g.model.transitions {
		g.model.removeTransition(t)
	}
	for _, p := range g.model.places {
		g.model.removePlace(p)
	}
	for _, gg := range g.model.groups {
		g.model.removeGroup(gg)
	}
	for i, gg := range tg.groups {
		if gg == g {
			tg.groups = append(tg.groups[:i], tg.groups[i+1:]...)
			return
		}
	}
	g.parent = nil
	g.model.parent = nil
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

	for _, place := range tg.places {
		place.Align()
	}
	for _, transition := range tg.transitions {
		transition.Align()
	}
}

func (tg *teg) deselectAll() {
	for k := range tg.selected {
		delete(tg.selected, k)
	}
}

func (tg *teg) deselectItem(it item) {
	delete(tg.selected, it)
	if g, ok := it.(*group); ok {
		for _, t := range g.inputs {
			delete(tg.selected, t)
		}
		for _, t := range g.outputs {
			delete(tg.selected, t)
		}
	}
}

func (tg *teg) selectItem(it item) {
	if it != nil {
		tg.selected[it] = true
	}
	if g, ok := it.(*group); ok {
		for _, t := range g.inputs {
			tg.selected[t] = true
		}
		for _, t := range g.outputs {
			tg.selected[t] = true
		}
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
	if !selected {
		selected = (t.proxy != nil && t.proxy.parent.parent != nil &&
			t.proxy.parent.parent.parent.isSelected(t))
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
	var g *group
	if t.proxy != nil {
		g = t.proxy.parent.parent
		t = t.proxy
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
		if g != nil {
			g.updateIO()
			g.adjustIO()
		}
	}
}

func (t *transition) unlink(p *place, inbound bool) {
	var g *group
	if t.proxy != nil {
		g = t.proxy.parent.parent
		t = t.proxy
	}
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
		if g != nil {
			g.updateIO()
			g.adjustIO()
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
			p.parent = tg
			tg.places = append(tg.places, p)
		} else if t, ok := it.(*transition); ok {
			t.parent = tg
			tg.transitions = append(tg.transitions, t)
		} else if g, ok := it.(*group); ok {
			g.parent = tg
			tg.groups = append(tg.groups, g)
		}
	}
	// cleanup refs in tg2
	if len(tg.places) > 0 {
		places := make([]*place, 0, 256)
		for _, p := range tg2.places {
			if !items[p] {
				places = append(places, p)
			}
		}
		tg2.places = places
	}
	if len(tg.transitions) > 0 {
		transitions := make([]*transition, 0, 256)
		for _, t := range tg2.transitions {
			if !items[t] {
				transitions = append(transitions, t)
			}
		}
		tg2.transitions = transitions
	}
	if len(tg.groups) > 0 {
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
			it, _ := t.Copy()
			tNew := it.(*transition)
			tNew.parent = tg
			clones[t] = tNew
			tg.transitions = append(tg.transitions, tNew)
			for _, p := range t.in {
				if _, ok := items[p]; ok {
					var pNew *place
					if clone, ok := clones[p]; !ok {
						clone, _ = p.Copy()
						pNew = clone.(*place)
						pNew.parent = tg
						clones[p] = pNew
						tg.places = append(tg.places, pNew)
					} else {
						pNew = clones[p].(*place)
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
					if clone, ok := clones[p]; !ok {
						clone, _ = p.Copy()
						pNew = clone.(*place)
						pNew.parent = tg
						clones[p] = pNew
						tg.places = append(tg.places, pNew)
					} else {
						pNew = clones[p].(*place)
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
			if clone, ok := clones[p]; !ok {
				clone, _ = p.Copy()
				pNew := clone.(*place)
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
			it, gClones := g.Copy()
			gNew := it.(*group)
			gNew.parent = tg
			gNew.model.parent = gNew
			clones[g] = gNew
			tg.groups = append(tg.groups, gNew)
			for _, t := range g.inputs {
				for _, p := range t.in {
					newPlace := clones[p].(*place)
					newProxy := gClones[t.proxy].(*transition)
					for _, t2 := range gNew.inputs {
						if t2.proxy == newProxy {
							t2.link(newPlace, true)
						}
					}
				}
			}
			for _, t := range g.outputs {
				for _, p := range t.out {
					newPlace := clones[p].(*place)
					newProxy := gClones[t.proxy].(*transition)
					for _, t2 := range gNew.outputs {
						if t2.proxy == newProxy {
							t2.link(newPlace, false)
						}
					}
				}
			}
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

func newTeg() *teg {
	return &teg{
		util:        &utility{},
		places:      make([]*place, 0, 256),
		transitions: make([]*transition, 0, 256),
		groups:      make([]*group, 0, 32),
		selected:    make(map[item]bool, 256),
		inputs:      make([]*transition, 0, 8),
		outputs:     make([]*transition, 0, 8),
		updated:     make(chan interface{}, 100),
	}
}
