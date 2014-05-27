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
	if c.enabled[c.models[i].ioId] != enabled {
		c.enabled[c.models[i].ioId] = enabled
		if enabled {
			c.SetActive(i)
		}
		c.models[i].update()
	}
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
	active := c.Active()
	if active != nil {
		expr = active.Dioid().String()
	}
	return
}

func (c *Ctrl) DioidLatex() (expr string) {
	active := c.Active()
	if active != nil {
		expr = dioid.Latex(active.Dioid().String())
	}
	return
}

func (c *Ctrl) SetDioid(expr string) bool {
	active := c.Active()
	if active == nil {
		return false
	}
	serie, err := dioid.Eval(expr)
	if err != nil {
		c.Error(err)
		return false
	}
	active.SetDioid(serie)
	active.update()
	return true
}

func (c *Ctrl) SetActive(i int) {
	active := c.Active()
	if active != nil {
		active.deselectAll()
		active.mergeTemporary()
		active.update()
	}
	c.ActiveLayer = i
	if i >= 0 && c.enabled[c.models[i].ioId] {
		c.Layers.active = c.models[i].ioId
		c.models[i].update()
	}
}

func (c *Ctrl) Active() *Plane {
	if c.ActiveLayer < 0 {
		return nil
	}
	if len(c.models) < 1 {
		return nil
	}
	return c.models[c.ActiveLayer]
}

func (c *Ctrl) Flush() {
	for _, m := range c.models {
		m.update()
	}
}

func (c *Ctrl) Fix() {
	active := c.Active()
	if active == nil {
		return
	}
	active.deselectAll()
	active.mergeTemporary()
	active.update()
}

func (c *Ctrl) Star() {
	active := c.Active()
	if active == nil {
		return
	}
	set := make([]*vertex, 0, len(active.selected))
	for v := range active.selected {
		set = append(set, v)
	}
	active.deselectAll()
	active.star(set...)
	active.update()
}

func (c *Ctrl) HasTemporary() bool {
	active := c.Active()
	if active == nil {
		return false
	}
	return len(c.Active().temporary) > 0
}

func (c *Ctrl) DefinedSelected() bool {
	active := c.Active()
	if active == nil {
		return false
	}
	for v := range c.Active().selected {
		if !v.temp {
			return true
		}
	}
	return false
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
					active := c.Active()
					if active == nil {
						continue
					}
					v, found := active.findV(x, y)
					var text string
					if found {
						text = dioid.Gd{G: v.G(), D: v.D()}.String()
					}
					if text != c.VertexText {
						c.VertexText = text
						qml.Changed(c, &c.VertexText)
					}
				case EventMousePress:
					active := c.Active()
					if active == nil {
						continue
					}
					x0, y0 = x, y
					v, found := active.findV(x, y)

					if c.ModifierKeyShift && !found {
						active.deselectAll()
						v := active.placeV(x, y)
						active.selectV(v)
						focused = v
					} else if !found {
						if !c.ModifierKeyControl {
							active.deselectAll()
							active.update()
						}
						active.util.min = &geometry.Point{x, y}
					} else {
						focused = v
						x0, y0 = v.X, v.Y
						if len(active.selected) > 1 {
							if v.isSelected() && c.ModifierKeyControl {
								active.deselectV(v)
							} else if c.ModifierKeyControl {
								active.selectV(v)
							} else if !v.isSelected() {
								active.deselectAll()
								active.selectV(v)
							}
						} else if !c.ModifierKeyControl {
							active.deselectAll()
							active.selectV(v)
						} else {
							active.selectV(v)
						}
						active.update()
					}
				case EventMouseMove:
					active := c.Active()
					if active == nil {
						continue
					}
					dx, dy := x-x0, y-y0
					x0, y0 = x, y

					if focused != nil {
						active.util.kind = UtilNone
						active.util.max = nil
						//active.placeV(x, y)
						for v := range active.selected {
							v.shift(dx, dy)
						}

						active.update()
					} else {
						active.util.kind = UtilRect
						active.util.max = &geometry.Point{x, y}
						dx := active.util.max.X - active.util.min.X
						dy := active.util.max.Y - active.util.min.Y
						var rect *geometry.Rect
						if dx == 0 || dy == 0 {
							continue
						}
						w, h := math.Abs(dx), math.Abs(dy)
						if dx < 0 && dy < 0 {
							rect = geometry.NewRect(
								active.util.max.X,
								active.util.max.Y,
								w, h,
							)
						} else if dx < 0 /* && dy > 0 */ {
							rect = geometry.NewRect(
								active.util.max.X,
								active.util.max.Y-h,
								w, h,
							)
						} else if dx > 0 && dy < 0 {
							rect = geometry.NewRect(
								active.util.max.X-w,
								active.util.max.Y,
								w, h,
							)
						} else /* dx > 0 && dy > 0 */ {
							rect = geometry.NewRect(
								active.util.min.X,
								active.util.min.Y,
								w, h,
							)
						}
						for _, v := range active.defined {
							if v.bound().Intersect(rect) {
								active.selectV(v)
							} else {
								active.deselectV(v)
							}
						}
						for _, v := range active.temporary {
							if v.bound().Intersect(rect) {
								active.selectV(v)
							} else {
								active.deselectV(v)
							}
						}
						active.update()
					}

				case EventMouseRelease:
					active := c.Active()
					if active == nil {
						continue
					}
					for v := range active.selected {
						v.align()
					}
					focused = nil
					active.util.kind = UtilNone
					active.update()
				}
			default:
				log.Println("Event not supported")
			}
		}
	}()
}

func (c *Ctrl) handleKeyEvent(ev *keyEvent) {
	var updated bool
	active := c.Active()
	if active == nil {
		return
	}
	// log.Printf("key: %v (%v)", ev.keycode, ev.text)
	if c.ModifierKeyControl {
		switch ev.keycode {
		case KeyCodeA:
			for _, v := range active.defined {
				active.selectV(v)
			}
			for _, v := range active.temporary {
				active.selectV(v)
			}
		case 16777219, 16777223, 8:
			for v := range active.selected {
				active.removeV(v)
				updated = true
			}
			if updated {
				active.SetDioid(active.Dioid())
			}
		}
	}

	if updated {
		active.update()
	}
}
