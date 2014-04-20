package tegview

import (
	"log"
	"unicode"
	"unicode/utf8"
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
		var focused item
		for {
			switch ev := (<-c.events).(type) {
			case *keyEvent:
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
							}
						case 13, 16777220, 16777221: // return
							it.SetLabel(l + "\n")
						case 10:
							it.SetLabel(l + " ")
						default:
							rune, _ := utf8.DecodeRuneInString(ev.text)
							if rune != utf8.RuneError && unicode.IsGraphic(rune) {
								it.SetLabel(l + string(rune))
							}
						}
						updated = true
					}
				}
				if updated {
					c.model.updated()
				}
			case *mouseEvent:
				if ev.x < 0 {
					c.CanvasWindowX += ev.x - 10.0
				} else if ev.x > c.CanvasWindowWidth {
					c.CanvasWindowX += ev.x - c.CanvasWindowWidth + 10.0
				}
				if ev.y < 0 {
					c.CanvasWindowY += ev.y - 10.0
				} else if ev.y > c.CanvasWindowHeight {
					c.CanvasWindowY += ev.y - c.CanvasWindowHeight + 10.0
				}
				x, y := c.WindowCoordsToRelativeGlobal(ev.x, ev.y)

				switch ev.kind {
				case EventMousePress:
					x0, y0 = x, y
					it, ok := c.model.findDrawable(x, y)

					if !ok && !c.ModifierKeyControl {
						c.model.deselectAll()
					} else if ok {
						x0, y0 = it.Center()[0], it.Center()[1]
						if control, ok := it.(*controlPoint); ok {
							control.modified = true
							focused = it
						} else {
							focused = it
							if len(c.model.selected) > 1 && c.model.isSelected(it) && c.ModifierKeyControl {
								c.model.deselectItem(it)
							} else {
								if !c.ModifierKeyControl {
									c.model.deselectAll()
								}
								c.model.selectItem(it)
							}
							if place, ok := focused.(*place); ok {
								if c.ModifierKeyShift {
									place.resetProperties()
								}
							}
						}
					}
					c.model.updated()
				case EventMouseMove:
					dx, dy := x-x0, y-y0
					x0, y0 = x, y
					if focused != nil {
						focused.Move(x, y)

						if place, ok := focused.(*place); ok {
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
						c.model.updated()
					}

				case EventMouseRelease:
					focused = nil

				case EventMouseClick:
					// unused

				case EventMouseDoubleClick:
					if c.ModifierKeyControl || c.ModifierKeyShift {
						continue
					}

					if focused != nil {
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
