package planeview

import (
	"errors"
	"log"
	"math"

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
	KeyCodeA = 65
	KeyCodeC = 67
	KeyCodeV = 86
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

type stopEvent struct{}

type Ctrl struct {
	CanvasWidth        float64
	CanvasHeight       float64
	CanvasWindowX      float64
	CanvasWindowY      float64
	CanvasWindowHeight float64
	CanvasWindowWidth  float64
	Zoom               float64

	Title     string
	ErrorText string
	Layers    *Layers
	Updated   bool // fake trigger

	ModifierKeyControl bool
	ModifierKeyShift   bool

	models  []*Plane
	active  *Plane
	events  chan interface{}
	actions chan interface{}
	errors  chan error
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

func (c *Ctrl) QmlError(text string) {
	c.errors <- errors.New(text)
}

func (c *Ctrl) Error(err error) {
	c.errors <- err
}

func (c *Ctrl) LabelAt(i int) string {
	return c.models[i].ioLabel
}

func (c *Ctrl) IsInputAt(i int) bool {
	return c.models[i].input
}

func (c *Ctrl) IsEnabledAt(i int) bool {
	return c.Layers.IsEnabled(c.models[i].ioId)
}

func (c *Ctrl) SetActive(i int) {
	if i < 0 || i >= len(c.models) {
		c.active = nil
		return
	}
	c.active = c.models[i]
}

func (c *Ctrl) Flush() {
	if c.active == nil {
		return
	}
	c.active.update()
}

func (c *Ctrl) stopHandling() {
	c.events <- &stopEvent{}
}

func (c *Ctrl) handleEvents() {
	go func() {
		var x0, y0 float64
		var focused *vertex
		for {
			switch ev := (<-c.events).(type) {
			case *stopEvent:
				return
			case *keyEvent:
				if c.active == nil {
					continue
				}
				c.handleKeyEvent(ev)
			case *mouseEvent:
				if c.active == nil {
					continue
				}
				x, y := c.WindowCoordsToRelativeGlobal(ev.x, ev.y)

				switch ev.kind {
				case EventMousePress:
					x0, y0 = x, y
					v, found := c.active.findV(x, y)

					if c.ModifierKeyShift && !found {
						c.active.deselectAll()
						v := c.active.placeV(x, y)
						c.active.selectV(v)
						focused = v
					} else if !found {
						if !c.ModifierKeyControl {
							c.active.deselectAll()
							c.active.update()
						}
						c.active.util.min = &geometry.Point{x, y}
					} else {
						focused = v
						x0, y0 = v.X, v.Y
						if len(c.active.selected) > 1 {
							if v.isSelected() && c.ModifierKeyControl {
								c.active.deselectV(v)
							} else if c.ModifierKeyControl {
								c.active.selectV(v)
							} else if !v.isSelected() {
								c.active.deselectAll()
								c.active.selectV(v)
							}
						} else if !c.ModifierKeyControl {
							c.active.deselectAll()
							c.active.selectV(v)
						} else {
							c.active.selectV(v)
						}
						c.active.update()
					}
				case EventMouseMove:
					dx, dy := x-x0, y-y0
					x0, y0 = x, y

					if focused != nil {
						c.active.util.kind = UtilNone
						c.active.util.max = nil

						for v := range c.active.selected {
							v.shift(dx, dy)
						}

						c.active.update()
					} else {
						c.active.util.kind = UtilRect
						c.active.util.max = &geometry.Point{x, y}
						dx := c.active.util.max.X - c.active.util.min.X
						dy := c.active.util.max.Y - c.active.util.min.Y
						var rect *geometry.Rect
						if dx == 0 || dy == 0 {
							continue
						}
						w, h := math.Abs(dx), math.Abs(dy)
						if dx < 0 && dy < 0 {
							rect = geometry.NewRect(
								c.active.util.max.X,
								c.active.util.max.Y,
								w, h,
							)
						} else if dx < 0 /* && dy > 0 */ {
							rect = geometry.NewRect(
								c.active.util.max.X,
								c.active.util.max.Y-h,
								w, h,
							)
						} else if dx > 0 && dy < 0 {
							rect = geometry.NewRect(
								c.active.util.max.X-w,
								c.active.util.max.Y,
								w, h,
							)
						} else /* dx > 0 && dy > 0 */ {
							rect = geometry.NewRect(
								c.active.util.min.X,
								c.active.util.min.Y,
								w, h,
							)
						}
						for _, v := range c.active.defined {
							if v.bound().Intersect(rect) {
								c.active.selectV(v)
							} else {
								c.active.deselectV(v)
							}
						}
						for _, v := range c.active.temporary {
							if v.bound().Intersect(rect) {
								c.active.selectV(v)
							} else {
								c.active.deselectV(v)
							}
						}
						c.active.update()
					}

				case EventMouseRelease:
					for v := range c.active.selected {
						v.align()
					}
					focused = nil
					c.active.util.kind = UtilNone
					c.active.update()
				}
			default:
				log.Println("Event not supported")
			}
		}
	}()
}

func (c *Ctrl) handleKeyEvent(ev *keyEvent) {
	var updated bool
	// log.Printf("key: %v (%v)", ev.keycode, ev.text)
	if c.ModifierKeyControl {
		switch ev.keycode {
		case KeyCodeA:
			for _, v := range c.active.defined {
				c.active.selectV(v)
			}
			for _, v := range c.active.temporary {
				c.active.selectV(v)
			}
		case 16777219, 16777223, 8:
			for v := range c.active.selected {
				c.active.removeV(v)
				updated = true
			}
		}
	}

	if updated {
		c.active.update()
	}
}
