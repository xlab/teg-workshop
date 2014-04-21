package tegview

import (
	"log"
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

func (c *Ctrl) shiftItem(it item, dx float64, dy float64) {
	it.Shift(dx, dy)

	if place, ok := it.(*place); ok {
		if c.ModifierKeyShift {
			place.resetProperties()
		} else {
			if place.in != nil {
				if place.inControl.modified {
					place.inControl.Shift(dx, dy)
				} else {
					place.resetControlPoint(true)
				}
			}

			if place.out != nil {
				if place.outControl.modified {
					place.outControl.Shift(dx, dy)
				} else {
					place.resetControlPoint(false)
				}
			}
		}
	}
}

func (c *Ctrl) handleEvents() {
	go func() {
		var x0, y0 float64
		var focused item
		for {
			switch ev := (<-c.events).(type) {
			case *keyEvent:
				c.handleKeyEvent(ev)
			case *mouseEvent:
				x, y := c.WindowCoordsToRelativeGlobal(ev.x, ev.y)

				switch ev.kind {
				case EventMousePress:
					x0, y0 = x, y
					it, ok := c.model.findDrawable(x, y)

					if c.ModifierKeyAlt {
						if !ok {
							c.model.deselectAll()
						} else if ok {
							focused = it
							c.model.deselectAll()
							c.model.selectItem(it)
						}
						c.model.MagicStroke.X0 = x
						c.model.MagicStroke.Y0 = y
						c.model.updated()
					} else if c.ModifierKeyShift && !ok {
						c.model.deselectAll()
						var it item
						if c.ModifierKeyControl {
							it = c.model.addTransition(x, y)
						} else {
							it = c.model.addPlace(x, y)
						}
						c.model.selectItem(it)
						focused = it
					} else {
						if !ok && !c.ModifierKeyControl {
							c.model.deselectAll()
						} else if ok {
							x0, y0 = it.Center()[0], it.Center()[1]
							focused = it
							if control, ok := it.(*controlPoint); ok {
								control.modified = true
							} else {
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
						}
						c.model.updated()
					}
				case EventMouseMove:
					dx, dy := x-x0, y-y0
					x0, y0 = x, y

					if c.ModifierKeyAlt {
						c.model.MagicStroke.X1 = x
						c.model.MagicStroke.Y1 = y
						c.model.MagicStrokeAvailable = true
						c.model.updatedMagicStroke()
						if focused != nil {
							// search for connect
							if it, ok := c.model.findDrawable(x, y); ok {
								c.model.selectItem(it)
							} else {
								c.model.deselectAll()
								c.model.selectItem(focused)
							}
						}
						c.model.updated()
					} else {
						if _, ok := focused.(*controlPoint); ok {
							focused.Move(x, y)
							c.model.updated()
						} else {
							if len(c.model.selected) > 1 {
								for it := range c.model.selected {
									c.shiftItem(it, dx, dy)
									if place, ok := it.(*place); ok {
										if place.in != nil {
											place.in.OrderArcs(false)
										}
										if place.out != nil {
											place.out.OrderArcs(true)
										}
									} else if transition, ok := it.(*transition); ok {
										transition.OrderArcs(true)
										transition.OrderArcs(false)
									}
								}
								c.model.updated()
							} else if focused != nil {
								c.shiftItem(focused, dx, dy)
								c.model.updated()
								if place, ok := focused.(*place); ok {
									if place.in != nil {
										place.in.OrderArcs(false)
									}
									if place.out != nil {
										place.out.OrderArcs(true)
									}
								} else if transition, ok := focused.(*transition); ok {
									transition.OrderArcs(true)
									transition.OrderArcs(false)
								}
							}
						}
					}

				case EventMouseRelease:
					if c.ModifierKeyAlt && c.model.MagicStrokeAvailable {
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

							for _, transition := range c.model.transitions {
								centerT := transition.Center()
								for _, place := range transition.in {
									centerP := place.Center()
									control := place.outControl.Center()
									if geometry.CheckSegmentsCrossing(start, end, centerP, control) ||
										geometry.CheckSegmentsCrossing(start, end, control, centerT) {
										c.model.disconnectItems(transition, place, true)
									}
								}
								for _, place := range transition.out {
									centerP := place.Center()
									control := place.inControl.Center()
									if geometry.CheckSegmentsCrossing(start, end, centerP, control) ||
										geometry.CheckSegmentsCrossing(start, end, control, centerT) {
										c.model.disconnectItems(transition, place, false)
									}
								}
							}
						}
					}

					focused = nil
					c.model.MagicStrokeAvailable = false
					c.model.updatedMagicStroke()
					c.model.updated()

				case EventMouseClick:
					// unused

				case EventMouseDoubleClick:
					if c.ModifierKeyControl || c.ModifierKeyShift {
						continue
					}
					if focused != nil {
						c.model.deselectAll()
						c.model.selectItem(focused)
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
					place.resetProperties()
					updated = true
				case 16777219, 16777223, 8:
					c.model.deselectItem(it)
					c.model.removeTransition(transition)
					updated = true
				}
			}
		}
	} else {
		// plaintext
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
