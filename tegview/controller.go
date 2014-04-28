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

	model  *TegModel
	events chan interface{}
	mode   int
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
						} else if _, cp := smth.(*controlPoint); !cp {
							focused = smth
							c.model.deselectAll()
							c.model.selectItem(smth.(item))
						}
						c.model.MagicStroke.X0 = x
						c.model.MagicStroke.Y0 = y
						c.model.updated()
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
						}
						c.model.MagicStroke.X0 = x
						c.model.MagicStroke.Y0 = y
						c.model.updated()
					} else {
						focused = smth
						if control, ok := smth.(*controlPoint); ok {
							control.modified = true
						} else {
							it := smth.(item)
							x0, y0 = it.Center()[0], it.Center()[1]
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
						c.model.updated()
					}
				case EventMouseMove:
					dx, dy := x-x0, y-y0
					x0, y0 = x, y

					if c.ModifierKeyAlt {
						c.model.MagicStroke.X1 = x
						c.model.MagicStroke.Y1 = y
						c.model.MagicStrokeUsed = true
						c.model.MagicRectUsed = false
						c.model.updatedMagicStroke()
						if focused != nil {
							// search for connect
							smth, found := c.model.findDrawable(x, y)
							if _, cp := smth.(*controlPoint); found && !cp {
								c.model.selectItem(smth.(item))
							} else if _, cp := focused.(*controlPoint); !found && !cp {
								c.model.deselectAll()
								c.model.selectItem(focused.(item))
							}
						}
						c.model.updated()
					} else if c.ModifierKeyShift && !copied {
						copied = true
						clones := make(map[item]item, len(c.model.selected)*2)
						for it := range c.model.selected {
							if t, ok := it.(*transition); ok {
								tNew := t.Copy().(*transition)
								focused = tNew
								clones[t] = tNew
								c.model.transitions = append(c.model.transitions, tNew)
								c.model.TransitionsLen = len(c.model.transitions)
								for _, p := range t.in {
									if c.model.isSelected(p) {
										var pNew *place
										if pNew, ok = clones[p].(*place); !ok {
											pNew = p.Copy().(*place)
											clones[p] = pNew
											focused = pNew
											c.model.places = append(c.model.places, pNew)
											c.model.PlacesLen = len(c.model.places)
										}
										c.model.connectItems(tNew, pNew, true)
										if p.outControl.modified {
											pNew.outControl = newControlPoint(p.outControl.X(),
												p.outControl.Y(), true)
										}
									}
								}
								for _, p := range t.out {
									if c.model.isSelected(p) {
										var pNew *place
										if pNew, ok = clones[p].(*place); !ok {
											pNew = p.Copy().(*place)
											clones[p] = pNew
											focused = pNew
											c.model.places = append(c.model.places, pNew)
											c.model.PlacesLen = len(c.model.places)
										}
										c.model.connectItems(tNew, pNew, false)
										if p.inControl.modified {
											pNew.inControl = newControlPoint(p.inControl.X(),
												p.inControl.Y(), true)
										}
									}
								}
							}
						}
						for k := range clones {
							c.model.deselectItem(k)
						}
						for it := range c.model.selected {
							if p, ok := it.(*place); ok {
								var pNew *place
								if pNew, ok = clones[p].(*place); !ok {
									pNew = p.Copy().(*place)
									clones[p] = pNew
									focused = pNew
									c.model.places = append(c.model.places, pNew)
									c.model.PlacesLen = len(c.model.places)
								}
							}
						}
						c.model.deselectAll()
						for _, v := range clones {
							c.model.selectItem(v)
						}
						c.model.updated()
					} else if focused != nil {
						if point, cp := focused.(*controlPoint); cp {
							point.Move(x, y)
							c.model.updated()
							continue
						}

						toOrder := make(map[*transition]bool, len(c.model.transitions))
						toResetIn := make(map[*place]bool, len(c.model.places))
						toResetOut := make(map[*place]bool, len(c.model.places))
						for it := range c.model.selected {
							c.model.shiftItem(it, dx, dy)
							if place, ok := it.(*place); ok {
								if place.in != nil {
									toOrder[place.in] = true
								}
								if place.out != nil {
									toOrder[place.out] = true
								}
							} else if transition, ok := it.(*transition); ok {
								toOrder[transition] = true
								for _, p := range transition.out {
									if !p.inControl.modified {
										toResetIn[p] = true
									}
								}
								for _, p := range transition.in {
									if !p.outControl.modified {
										toResetOut[p] = true
									}
								}
							}
						}
						for t, _ := range toOrder {
							t.OrderArcs(true)
							t.OrderArcs(false)
						}
						for p, _ := range toResetIn {
							p.resetControlPoint(true)
						}
						for p, _ := range toResetOut {
							p.resetControlPoint(false)
						}
						c.model.updated()

					} else {
						c.model.MagicStroke.X1 = x
						c.model.MagicStroke.Y1 = y
						dx := c.model.MagicStroke.X1 - c.model.MagicStroke.X0
						dy := c.model.MagicStroke.Y1 - c.model.MagicStroke.Y0
						var rect *geometry.Rect
						if dx == 0 || dy == 0 {
							continue
						}
						w, h := math.Abs(dx), math.Abs(dy)
						if dx < 0 && dy < 0 {
							rect = geometry.NewRect(
								c.model.MagicStroke.X1,
								c.model.MagicStroke.Y1,
								w, h,
							)
						} else if dx < 0 /* && dy > 0 */ {
							rect = geometry.NewRect(
								c.model.MagicStroke.X1,
								c.model.MagicStroke.Y1-h,
								w, h,
							)
						} else if dx > 0 && dy < 0 {
							rect = geometry.NewRect(
								c.model.MagicStroke.X1-w,
								c.model.MagicStroke.Y1,
								w, h,
							)
						} else /* dx > 0 && dy > 0 */ {
							rect = geometry.NewRect(
								c.model.MagicStroke.X0,
								c.model.MagicStroke.Y0,
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
						if !c.model.MagicRectUsed {
							c.model.MagicStrokeUsed = true
							c.model.MagicRectUsed = true
							c.model.updatedMagicStroke()
						}
						c.model.updated()
					}

				case EventMouseRelease:
					if c.ModifierKeyAlt && c.model.MagicStrokeUsed {
						if it, ok := c.model.findDrawable(x, y); focused != nil && focused != it && ok {
							if p, ok := it.(*place); ok && p.in == nil {
								if t, ok := focused.(*transition); ok {
									c.model.connectItems(t, p, false)
								} else if p2, ok := focused.(*place); ok && p2.out == nil {
									// p -- p2, new transition between
									pCenter, p2Center := p.Center(), p2.Center()
									t := c.model.addTransition((pCenter[0]+p2Center[0])/2,
										(pCenter[1]+p2Center[1])/2)
									c.model.connectItems(t, p2, true)
									c.model.connectItems(t, p, false)
								} else {
									log.Println("Unable to connect smth to place")
								}
							} else if t, ok := it.(*transition); ok {
								if p, ok := focused.(*place); ok {
									c.model.connectItems(t, p, true)
								} else if t2, ok := focused.(*transition); ok {
									// t -- t2, new place between
									tCenter, t2Center := t.Center(), t2.Center()
									p := c.model.addPlace((tCenter[0]+t2Center[0])/2,
										(tCenter[1]+t2Center[1])/2)
									c.model.connectItems(t2, p, false)
									c.model.connectItems(t, p, true)
								} else {
									log.Println("Unable to connect smth to transition")
								}
							}
							c.model.deselectAll()
						} else if focused == nil && !ok {
							// stroking the void, cut the connections
							start := geometry.NewPoint(c.model.MagicStroke.X0, c.model.MagicStroke.Y0)
							end := geometry.NewPoint(c.model.MagicStroke.X1, c.model.MagicStroke.Y1)

							var toDisconnect []*place
							for _, t := range c.model.transitions {
								centerT := t.Center()
								toDisconnect = make([]*place, 0, len(t.in))
								for _, p := range t.in {
									centerP := p.Center()
									control := p.outControl.Center()
									if geometry.CheckSegmentsCrossing(start, end, centerP, control) ||
										geometry.CheckSegmentsCrossing(start, end, control, centerT) {
										toDisconnect = append(toDisconnect, p)
									}
								}
								for _, p := range toDisconnect {
									c.model.disconnectItems(t, p, true)
								}
								toDisconnect = make([]*place, 0, len(t.out))
								for _, p := range t.out {
									centerP := p.Center()
									control := p.inControl.Center()
									if geometry.CheckSegmentsCrossing(start, end, centerP, control) ||
										geometry.CheckSegmentsCrossing(start, end, control, centerT) {
										toDisconnect = append(toDisconnect, p)
									}
								}
								for _, p := range toDisconnect {
									c.model.disconnectItems(t, p, false)
								}
							}
						}
					} else {
						first := true
						var dx, dy float64
						for it := range c.model.selected {
							if first {
								dx, dy = it.Align()
								first = false
							} else {
								it.Shift(dx, dy)
								if p, ok := it.(*place); ok {
									p.shiftControls(dx, dy)
								}
							}
						}
					}

					focused = nil
					copied = false
					c.model.MagicStrokeUsed = false
					c.model.MagicRectUsed = false
					c.model.updatedMagicStroke()
					c.model.updated()

				case EventMouseDoubleClick:
					if c.ModifierKeyControl || c.ModifierKeyShift {
						continue
					}
					if focused != nil {
						c.model.deselectAll()
						c.model.selectItem(focused.(item))
						if transition, ok := focused.(*transition); ok {
							transition.Rotate()
							c.model.updated()
						}
						if place, ok := focused.(*place); ok {
							if c.ModifierKeyAlt {
								place.timer++
							} else {
								place.counter++
							}
							c.model.updated()
						}
					}
				}
			default:
				log.Println("Event not supported")
			}
		}
	}()
}

func (c *Ctrl) handleKeyEvent(ev *keyEvent) {
	var updated bool
	if c.ModifierKeyControl {
		for it := range c.model.selected {
			if place, ok := it.(*place); ok {
				switch ev.keycode {
				case KeyCodeD:
					place.counter++
				case KeyCodeC:
					place.counter--
				case KeyCodeF:
					place.resetProperties()
				case KeyCodeT:
					place.timer++
				case KeyCodeG:
					place.timer--
				case 16777219, 16777223, 8:
					c.model.deselectItem(it)
					c.model.removePlace(place)
				}
				if place.counter > MaxPlaceCounter {
					place.counter = MaxPlaceCounter
				} else if place.counter < MinPlaceCounter {
					place.counter = 0
				}
				if place.timer > MaxPlaceTimer {
					place.timer = MaxPlaceTimer
				} else if place.timer < MinPlaceTimer {
					place.timer = 0
				}
				updated = true
			} else if transition, ok := it.(*transition); ok {
				switch ev.keycode {
				case KeyCodeF:
					transition.resetProperties()
					updated = true
				case 16777219, 16777223, 8:
					c.model.deselectItem(it)
					c.model.removeTransition(transition)
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
					it.SetLabel(l[:len(l)-1])
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
			// displaying connections
			updated = true
		}
	}
	if updated {
		c.model.updated()
	}
}
