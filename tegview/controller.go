package tegview

import (
	"log"
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/xlab/teg-workshop/geometry"
)

const (
	EventMousePress = iota
	EventMouseRelease
	EventMouseMove
	EventMouseClick
	EventMouseDoubleClick
	EventKeyPress
	EventKeyRelease
)

const (
	MaxPlaceCounter = 9999
	MinPlaceCounter = 0
	MaxPlaceTimer   = 9999
	MinPlaceTimer   = 0
)

const (
	KeyCodeC = 67
	KeyCodeD = 68
	KeyCodeF = 70
	KeyCodeG = 71
	KeyCodeT = 84
	KeyCodeJ = 74
	KeyCodeN = 78
	KeyCodeK = 75
	KeyCodeM = 77
	KeyCodeO = 79
	KeyCodeZ = 90
)

type mouseEvent struct {
	x    float64
	y    float64
	kind int
}

type keyEvent struct {
	keycode int
	text    string
	kind    int
}

type Ctrl struct {
	CanvasWidth        float64
	CanvasHeight       float64
	CanvasWindowX      float64
	CanvasWindowY      float64
	CanvasWindowHeight float64
	CanvasWindowWidth  float64
	Zoom               float64

	ModifierKeyControl bool
	ModifierKeyShift   bool
	ModifierKeyAlt     bool

	mode   int
	model  *teg
	events chan interface{}
}

func (c *Ctrl) KeyPressed(keycode int, text string) {
	c.events <- &keyEvent{keycode: keycode, text: text, kind: EventKeyPress}
}

func (c *Ctrl) MouseClicked(x, y float64) {
	c.events <- &mouseEvent{x: x, y: y, kind: EventMouseClick}
}

func (c *Ctrl) MouseDoubleClicked(x, y float64) {
	c.events <- &mouseEvent{x: x, y: y, kind: EventMouseDoubleClick}
}

func (c *Ctrl) MousePressed(x, y float64) {
	c.events <- &mouseEvent{x: x, y: y, kind: EventMousePress}
}

func (c *Ctrl) MouseReleased(x, y float64) {
	c.events <- &mouseEvent{x: x, y: y, kind: EventMouseRelease}
}

func (c *Ctrl) MouseMoved(x, y float64) {
	c.events <- &mouseEvent{x: x, y: y, kind: EventMouseMove}
}

func (c *Ctrl) WindowCoordsToRelativeGlobal(x, y float64) (x1, y1 float64) {
	xGlobal := c.CanvasWindowX + x
	yGlobal := c.CanvasWindowY + y
	x1 = (xGlobal - c.CanvasWidth/2 - c.CanvasWindowWidth/2) / c.Zoom
	y1 = (yGlobal - c.CanvasHeight/2 - c.CanvasWindowHeight/2) / c.Zoom
	return
}

func (c *Ctrl) Flush() {
	c.model.update()
}

func (c *Ctrl) handleEvents() {
	go func() {
		var x0, y0 float64
		var focused interface{}
		var copied bool
		for {
			switch ev := (<-c.events).(type) {
			case *keyEvent:
				c.handleKeyEvent(ev)
			case *mouseEvent:
				x, y := c.WindowCoordsToRelativeGlobal(ev.x, ev.y)

				switch ev.kind {
				case EventMousePress:
					x0, y0 = x, y
					smth, found := c.model.findDrawable(x, y)

					if c.ModifierKeyAlt {
						if !found {
							c.model.deselectAll()
							c.model.update()
						} else if _, cp := smth.(*controlPoint); !cp {
							focused = smth
							c.model.deselectAll()
							c.model.selectItem(smth.(item))
							c.model.update()
						}
						c.model.util.min = &geometry.Point{x, y}
					} else if c.ModifierKeyShift && !found {
						var it item
						c.model.deselectAll()
						if c.ModifierKeyControl {
							it = c.model.addTransition(x, y)
						} else {
							it = c.model.addPlace(x, y)
						}
						c.model.selectItem(it)
						focused = it
						copied = true // prevent copy of item
					} else if !found {
						if !c.ModifierKeyControl {
							c.model.deselectAll()
							c.model.update()
						}
						c.model.util.min = &geometry.Point{x, y}
					} else {
						focused = smth
						if control, ok := smth.(*controlPoint); ok {
							control.modified = true
						} else {
							it := smth.(item)
							//x0, y0 = it.Center().X, it.Center().Y
							if len(c.model.selected) > 1 {
								if c.model.isSelected(it) && c.ModifierKeyControl {
									c.model.deselectItem(it)
								} else if c.ModifierKeyControl {
									c.model.selectItem(it)
								} else if !c.model.isSelected(it) {
									c.model.deselectAll()
									c.model.selectItem(it)
								}
							} else if !c.ModifierKeyControl {
								c.model.deselectAll()
								c.model.selectItem(it)
							} else {
								c.model.selectItem(it)
							}
						}
						c.model.update()
					}
				case EventMouseMove:
					dx, dy := x-x0, y-y0
					x0, y0 = x, y

					if c.ModifierKeyAlt {
						c.model.util.kind = UtilStroke
						c.model.util.max = &geometry.Point{x, y}
						if focused != nil {
							smth, found := c.model.findDrawable(x, y)
							if _, cp := smth.(*controlPoint); found && !cp {
								c.model.selectItem(smth.(item))
							} else if _, cp := focused.(*controlPoint); !found && !cp {
								c.model.deselectAll()
								c.model.selectItem(focused.(item))
							}
						}
						c.model.update()
					} else if c.ModifierKeyShift && !copied {
						copied = true
						clones := c.model.cloneItems(c.model.selected)
						c.model.deselectAll()
						for _, v := range clones {
							c.model.selectItem(v)
						}
						c.model.update()
					} else if focused != nil {
						c.model.util.kind = UtilNone
						if point, cp := focused.(*controlPoint); cp {
							point.Move(x, y)
							c.model.update()
							continue
						}
						toOrder := make(map[*transition]bool, len(c.model.transitions))
						for it := range c.model.selected {
							if p, ok := it.(*place); ok {
								p.Shift(dx, dy)
								if p.in != nil {
									toOrder[p.in] = true
								}
								if p.out != nil {
									toOrder[p.out] = true
								}
							} else if t, ok := it.(*transition); ok {
								if t.proxy == nil {
									t.Shift(dx, dy)
									toOrder[t] = true
								}
							} else {
								it.Shift(dx, dy)
							}
						}
						for t := range toOrder {
							t.OrderArcs(true)
							t.OrderArcs(false)
						}
						c.model.update()
					} else {
						c.model.util.kind = UtilRect
						c.model.util.max = &geometry.Point{x, y}
						dx := c.model.util.max.X - c.model.util.min.X
						dy := c.model.util.max.Y - c.model.util.min.Y
						var rect *geometry.Rect
						if dx == 0 || dy == 0 {
							continue
						}
						w, h := math.Abs(dx), math.Abs(dy)
						if dx < 0 && dy < 0 {
							rect = geometry.NewRect(
								c.model.util.max.X,
								c.model.util.max.Y,
								w, h,
							)
						} else if dx < 0 /* && dy > 0 */ {
							rect = geometry.NewRect(
								c.model.util.max.X,
								c.model.util.max.Y-h,
								w, h,
							)
						} else if dx > 0 && dy < 0 {
							rect = geometry.NewRect(
								c.model.util.max.X-w,
								c.model.util.max.Y,
								w, h,
							)
						} else /* dx > 0 && dy > 0 */ {
							rect = geometry.NewRect(
								c.model.util.min.X,
								c.model.util.min.Y,
								w, h,
							)
						}
						for _, p := range c.model.places {
							if p.Bound().Intersect(rect) {
								c.model.selectItem(p)
							} else {
								c.model.deselectItem(p)
							}
						}
						for _, t := range c.model.transitions {
							if t.Bound().Intersect(rect) {
								c.model.selectItem(t)
							} else {
								c.model.deselectItem(t)
							}
						}
						for _, g := range c.model.groups {
							if g.Bound().Intersect(rect) {
								c.model.selectItem(g)
							} else {
								c.model.deselectItem(g)
							}
						}
						c.model.update()
					}

				case EventMouseRelease:
					if c.ModifierKeyAlt && c.model.util.kind == UtilStroke {
						if it, ok := c.model.findDrawable(x, y); focused != nil && focused != it && ok {
							if p, ok := it.(*place); ok && p.in == nil {
								if t, ok := focused.(*transition); ok {
									t.link(p, false)
								} else if p2, ok := focused.(*place); ok && p2.out == nil {
									// p -- p2, new transition between
									pCenter, p2Center := p.Center(), p2.Center()
									t := c.model.addTransition((pCenter.X+p2Center.X)/2,
										(pCenter.Y+p2Center.Y)/2)
									t.link(p2, true)
									t.link(p, false)
								} else {
									log.Println("Unable to link smth to place")
								}
							} else if t, ok := it.(*transition); ok {
								if p, ok := focused.(*place); ok {
									t.link(p, true)
								} else if t2, ok := focused.(*transition); ok {
									// t -- t2, new place between
									tCenter, t2Center := t.Center(), t2.Center()
									p := c.model.addPlace((tCenter.X+t2Center.X)/2,
										(tCenter.Y+t2Center.Y)/2)
									t2.link(p, false)
									t.link(p, true)
								} else {
									log.Println("Unable to link smth to transition")
								}
							}
							c.model.deselectAll()
						} else if focused == nil && !ok {
							// stroking the void, cut the links
							start := c.model.util.min
							end := c.model.util.max
							var toUnlink []*place
							unlink := func(t *transition, p *place, inbound bool, i int) {
								if inbound {
									control := p.outControl.Center()
									borderP := p.BorderPoint(control.X, control.Y, BorderPlaceDist)
									borderT := t.BorderPoint(true, i)
									if geometry.CheckSegmentsCrossing(start, end, borderP, control) ||
										geometry.CheckSegmentsCrossing(start, end, control, borderT) {
										toUnlink = append(toUnlink, p)
									}
								} else {
									control := p.inControl.Center()
									borderP := p.BorderPoint(control.X, control.Y, BorderPlaceTipDist)
									borderT := t.BorderPoint(false, i)
									if geometry.CheckSegmentsCrossing(start, end, borderP, control) ||
										geometry.CheckSegmentsCrossing(start, end, control, borderT) {
										toUnlink = append(toUnlink, p)
									}
								}
							}
							for _, t := range c.model.transitions {
								toUnlink = make([]*place, 0, len(t.in))
								for i, p := range t.in {
									unlink(t, p, true, i)
								}
								for _, p := range toUnlink {
									t.unlink(p, true)
								}
								toUnlink = make([]*place, 0, len(t.out))
								for i, p := range t.out {
									unlink(t, p, false, i)
								}
								for _, p := range toUnlink {
									t.unlink(p, false)
								}
							}
							for _, g := range c.model.groups {
								for _, t := range g.inputs {
									toUnlink = make([]*place, 0, len(t.in))
									for i, p := range t.in {
										unlink(t, p, true, i)
									}
									for _, p := range toUnlink {
										t.unlink(p, true)
									}

								}
								for _, t := range g.outputs {
									toUnlink = make([]*place, 0, len(t.out))
									for i, p := range t.out {
										unlink(t, p, false, i)
									}
									for _, p := range toUnlink {
										t.unlink(p, false)
									}
								}
							}

						}
					} else {
						c.alignItems(c.model.selected)
					}

					focused = nil
					copied = false
					c.model.util.kind = UtilNone
					c.model.update()

				case EventMouseDoubleClick:
					if c.ModifierKeyControl || c.ModifierKeyShift {
						continue
					}
					if focused != nil {
						c.model.deselectAll()
						c.model.selectItem(focused.(item))
						if t, ok := focused.(*transition); ok {
							t.Rotate()
							if t.proxy != nil {
								g := t.proxy.parent.parent
								g.adjustIO()
							}
							c.model.update()
						}
						if p, ok := focused.(*place); ok {
							if c.ModifierKeyAlt {
								p.timer++
							} else {
								p.counter++
							}
							c.model.update()
						}
						if g, ok := focused.(*group); ok {
							if g.folded {
								c.model.unfoldGroup(g)
							} else {
								c.model.foldGroup(g)
							}
							c.model.update()
						}
					}
				}
			default:
				log.Println("Event not supported")
			}
		}
	}()
}

func (c *Ctrl) alignItems(items map[item]bool) {
	first := true
	var dx, dy float64
	toOrder := make(map[*transition]bool, len(c.model.transitions))
	for it := range items {
		if first {
			dx, dy = it.Align()
			if p, ok := it.(*place); ok {
				if p.in != nil {
					toOrder[p.in] = true
				}
				if p.out != nil {
					toOrder[p.out] = true
				}
			}
			first = false
			continue
		}
		if p, ok := it.(*place); ok {
			p.Shift(dx, dy)
			if p.in != nil {
				toOrder[p.in] = true
			}
			if p.out != nil {
				toOrder[p.out] = true
			}
		} else if t, ok := it.(*transition); ok {
			if t.proxy == nil {
				t.Shift(dx, dy)
			}
		} else {
			it.Shift(dx, dy)
		}
	}
	for t := range toOrder {
		t.OrderArcs(true)
		t.OrderArcs(false)
	}
}

func (c *Ctrl) groupItems(items map[item]bool) {
	data := make(map[item]bool, len(items))
	for it := range items {
		switch it.(type) {
		case *transition:
			if it.(*transition).KindInGroup(items) == TransitionExposed {
				return // no way
			} else if it.(*transition).proxy == nil {
				data[it] = true
			}
		case *place:
			if it.(*place).KindInGroup(items) == PlaceExposed {
				return // no way
			}
			data[it] = true
		case *group:
			if it.(*group).KindInGroup(items) == GroupExposed {
				return // no way
			}
			data[it] = true

		}
	}
	g := c.model.addGroup(data)
	g.updateIO()
	g.adjustIO()
	g.Align()
	c.model.deselectAll()
	c.model.selectItem(g)
}

func (c *Ctrl) ungroupItems(g *group) {
	c.model.deselectAll()
	items := c.model.flatGroup(g)
	c.alignItems(items)
	for it := range items {
		c.model.selectItem(it)
	}
}

func (c *Ctrl) handleKeyEvent(ev *keyEvent) {
	var updated bool
	// log.Printf("key: %v (%v)", ev.keycode, ev.text)
	if c.ModifierKeyControl {
		switch ev.keycode {
		case KeyCodeG:
			proxies := 0
			for it := range c.model.selected {
				if t, ok := it.(*transition); ok && t.proxy != nil {
					proxies++
				}
			}
			switch l := len(c.model.selected) - proxies; {
			case l > 1:
				c.groupItems(c.model.selected)
				c.model.update()
			case l == 1:
				for it := range c.model.selected {
					if g, ok := it.(*group); ok {
						c.ungroupItems(g)
						c.model.update()
						return
					}
				}
			}
			return
		}

		for it := range c.model.selected {
			if g, ok := it.(*group); ok {
				switch ev.keycode {
				case KeyCodeF:
					g.resetProperties()
					updated = true
				case KeyCodeZ:
					g.folded = !g.folded
					updated = true
				case 16777219, 16777223, 8:
					c.model.deselectItem(it)
					c.model.removeGroup(g)
					updated = true
				}
			} else if p, ok := it.(*place); ok {
				switch ev.keycode {
				case KeyCodeJ:
					p.counter++
				case KeyCodeN:
					p.counter--
				case KeyCodeF:
					p.resetProperties()
				case KeyCodeK:
					p.timer++
				case KeyCodeM:
					p.timer--
				case 16777219, 16777223, 8:
					c.model.deselectItem(it)
					c.model.removePlace(p)
				}
				if p.counter > MaxPlaceCounter {
					p.counter = MaxPlaceCounter
				} else if p.counter < MinPlaceCounter {
					p.counter = 0
				}
				if p.timer > MaxPlaceTimer {
					p.timer = MaxPlaceTimer
				} else if p.timer < MinPlaceTimer {
					p.timer = 0
				}
				updated = true
			} else if t, ok := it.(*transition); ok {
				switch ev.keycode {
				case KeyCodeF:
					t.resetProperties()
					updated = true
				case KeyCodeT:
					t.nextKind()
					updated = true
				case 16777219, 16777223, 8:
					c.model.deselectItem(it)
					c.model.removeTransition(t)
					updated = true
				}
			}
		}
	} else {
		// plaintext input
		for it := range c.model.selected {
			l := it.Label()
			switch ev.keycode {
			case 8, 16777219, 16777223: // backspace
				if len(l) > 0 {
					_, size := utf8.DecodeLastRuneInString(l)
					it.SetLabel(l[:len(l)-size])
					updated = true
				}
			case 13, 16777220, 16777221: // return
				it.SetLabel(l + "\n")
				updated = true
			case 10:
				it.SetLabel(l + " ")
				updated = true
			default:
				rune, _ := utf8.DecodeRuneInString(ev.text)
				if rune != utf8.RuneError && unicode.IsGraphic(rune) {
					it.SetLabel(l + string(rune))
					updated = true
				}
			}
		}
		if !updated && c.ModifierKeyAlt {
			// displaying links
			updated = true
		}
	}
	if updated {
		c.model.update()
	}
}
