package planeview

import (
	"errors"
	"log"
	"math"

	"github.com/xlab/teg-workshop/dioid"
	"github.com/xlab/teg-workshop/geometry"
	"gopkg.in/qml.v0"
)

const (
	EventMousePress = iota
	EventMouseRelease
	EventMouseMove
	EventMouseHover
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
	DrawShadows        bool

	Title       string
	ErrorText   string
	VertexText  string
	Layers      *Layers
	ActiveLayer int
	Updated     bool // fake trigger

	ModifierKeyControl bool
	ModifierKeyShift   bool

	models  []*Plane
	enabled map[string]bool

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

func (c *Ctrl) MouseHovered(x, y float64) {
	c.events <- &mouseEvent{x: x, y: y, kind: EventMouseHover}
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

func (c *Ctrl) ColorAt(i int) (color string) {
	if i < len(c.models) {
		color = c.models[i].color
	}
	return
}

func (c *Ctrl) LabelAt(i int) (label string) {
	if i < len(c.models) {
		label = c.models[i].Label()
	}
	return
}

func (c *Ctrl) IsInputAt(i int) (is bool) {
	if i < len(c.models) {
		is = c.models[i].input
	}
	return
}

func (c *Ctrl) SetEnabledAt(i int, enabled bool) {
	c.enabled[c.models[i].ioId] = enabled
}

func (c *Ctrl) EnabledAt(i int) (enabled bool) {
	if i < len(c.models) {
		enabled = c.enabled[c.models[i].ioId]
	}
	return
}

func (c *Ctrl) DioidAt(i int) (expr string) {
	if i < len(c.models) {
		expr = c.models[i].Dioid().String()
	}
	return
}

func (c *Ctrl) Dioid() (expr string) {
	if len(c.models) > 0 {
		expr = c.Active().Dioid().String()
	}
	return
}

func (c *Ctrl) SetDioid(expr string) bool {
	serie, err := dioid.Eval(expr)
	if err != nil {
		c.Error(err)
		return false
	}
	c.Active().SetDioid(serie)
	c.Active().update()
	return true
}

func (c *Ctrl) SetActive(i int) {
	c.Active().deselectAll()
	c.Layers.active = c.models[i].ioId
	c.ActiveLayer = i
}

func (c *Ctrl) Active() *Plane {
	return c.models[c.ActiveLayer]
}

func (c *Ctrl) Flush() {
	if c.Layers.Length < 1 {
		return
	}
	c.Active().update()
}

func (c *Ctrl) Fix() {
	c.Active().deselectAll()
	c.Active().MergeTemporary()
	c.Active().update()
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
				if c.Layers.Length < 1 {
					continue
				}
				c.handleKeyEvent(ev)
			case *mouseEvent:
				if c.Layers.Length < 1 {
					continue
				}
				x, y := c.WindowCoordsToRelativeGlobal(ev.x, ev.y)

				switch ev.kind {
				case EventMouseHover:
					v, found := c.Active().findV(x, y)
					var text string
					if found {
						text = dioid.Gd{G: v.G(), D: v.D()}.String()
					}
					if text != c.VertexText {
						c.VertexText = text
						qml.Changed(c, &c.VertexText)
					}
				case EventMousePress:
					x0, y0 = x, y
					v, found := c.Active().findV(x, y)

					if c.ModifierKeyShift && !found {
						c.Active().deselectAll()
						v := c.Active().placeV(x, y)
						c.Active().selectV(v)
						focused = v
					} else if !found {
						if !c.ModifierKeyControl {
							c.Active().deselectAll()
							c.Active().update()
						}
						c.Active().util.min = &geometry.Point{x, y}
					} else {
						focused = v
						x0, y0 = v.X, v.Y
						if len(c.Active().selected) > 1 {
							if v.isSelected() && c.ModifierKeyControl {
								c.Active().deselectV(v)
							} else if c.ModifierKeyControl {
								c.Active().selectV(v)
							} else if !v.isSelected() {
								c.Active().deselectAll()
								c.Active().selectV(v)
							}
						} else if !c.ModifierKeyControl {
							c.Active().deselectAll()
							c.Active().selectV(v)
						} else {
							c.Active().selectV(v)
						}
						c.Active().update()
					}
				case EventMouseMove:
					dx, dy := x-x0, y-y0
					x0, y0 = x, y

					if focused != nil {
						c.Active().util.kind = UtilNone
						c.Active().util.max = nil
						//c.Active().placeV(x, y)
						for v := range c.Active().selected {
							v.shift(dx, dy)
						}

						c.Active().update()
					} else {
						c.Active().util.kind = UtilRect
						c.Active().util.max = &geometry.Point{x, y}
						dx := c.Active().util.max.X - c.Active().util.min.X
						dy := c.Active().util.max.Y - c.Active().util.min.Y
						var rect *geometry.Rect
						if dx == 0 || dy == 0 {
							continue
						}
						w, h := math.Abs(dx), math.Abs(dy)
						if dx < 0 && dy < 0 {
							rect = geometry.NewRect(
								c.Active().util.max.X,
								c.Active().util.max.Y,
								w, h,
							)
						} else if dx < 0 /* && dy > 0 */ {
							rect = geometry.NewRect(
								c.Active().util.max.X,
								c.Active().util.max.Y-h,
								w, h,
							)
						} else if dx > 0 && dy < 0 {
							rect = geometry.NewRect(
								c.Active().util.max.X-w,
								c.Active().util.max.Y,
								w, h,
							)
						} else /* dx > 0 && dy > 0 */ {
							rect = geometry.NewRect(
								c.Active().util.min.X,
								c.Active().util.min.Y,
								w, h,
							)
						}
						for _, v := range c.Active().defined {
							if v.bound().Intersect(rect) {
								c.Active().selectV(v)
							} else {
								c.Active().deselectV(v)
							}
						}
						for _, v := range c.Active().temporary {
							if v.bound().Intersect(rect) {
								c.Active().selectV(v)
							} else {
								c.Active().deselectV(v)
							}
						}
						c.Active().update()
					}

				case EventMouseRelease:
					for v := range c.Active().selected {
						v.align()
					}
					focused = nil
					c.Active().util.kind = UtilNone
					c.Active().update()
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
			for _, v := range c.Active().defined {
				c.Active().selectV(v)
			}
			for _, v := range c.Active().temporary {
				c.Active().selectV(v)
			}
		case 16777219, 16777223, 8:
			for v := range c.Active().selected {
				c.Active().removeV(v)
				updated = true
			}
			if updated {
				c.Active().SetDioid(c.Active().Dioid())
			}
		}
	}

	if updated {
		c.Active().update()
	}
}
