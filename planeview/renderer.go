package planeview

import (
	"github.com/xlab/teg-workshop/geometry"
	"github.com/xlab/teg-workshop/render"
	"github.com/xlab/teg-workshop/util"
)

const (
	TextFontNormal = "Georgia"
)

const (
	ColorSelected      = "#b10000"
	ColorDefault       = "#000000"
	ColorUtility       = "#3498db"
	ColorUtilityShadow = "#202980b9"
	ColorVertexPad     = "#90bdc3c7"
)

const (
	Thickness    = 6.0
	Padding      = 2.0
	TextFontSize = 14.0
)

type List struct {
	items  []interface{}
	Length int
}

func list(its []interface{}) *List {
	return &List{items: its, Length: len(its)}
}

func (l *List) Put(it interface{}) {
	l.items = append(l.items, it)
	l.Length++
}

func (l *List) At(i int) interface{} {
	return l.items[i]
}

type PlaneBuffer struct {
	Circles *List
	Pads    *List
	Rects   *List
	Polys   *List
	Texts   *List
	Lines   *List
	Chains  *List
}

func newPlaneBuffer() *PlaneBuffer {
	return &PlaneBuffer{
		Circles: list(make([]interface{}, 0, 256)),
		Pads:    list(make([]interface{}, 0, 256)),
		Rects:   list(make([]interface{}, 0, 256)),
		Polys:   list(make([]interface{}, 0, 10)),
		Texts:   list(make([]interface{}, 0, 100)),
		Lines:   list(make([]interface{}, 0, 256)),
		Chains:  list(make([]interface{}, 0, 100)),
	}
}

type planeRenderer struct {
	ctrl   *Ctrl
	buf    *PlaneBuffer
	Screen *PlaneBuffer
	Ready  bool

	zoom          float64
	canvasWidth   float64
	canvasHeight  float64
	viewboxWidth  float64
	viewboxHeight float64
	viewboxX      float64
	viewboxY      float64

	relateiveGlobalCenter *geometry.Point
}

func newPlaneRenderer(ctrl *Ctrl) *planeRenderer {
	return &planeRenderer{
		Screen: newPlaneBuffer(),
		buf:    newPlaneBuffer(),
		ctrl:   ctrl,
		zoom:   1.0,
	}
}

func (pr *planeRenderer) fixViewport() {
	pr.zoom = pr.ctrl.Zoom

	pr.canvasWidth = pr.ctrl.CanvasWidth
	pr.canvasHeight = pr.ctrl.CanvasHeight
	pr.viewboxWidth = pr.ctrl.CanvasWindowWidth
	pr.viewboxHeight = pr.ctrl.CanvasWindowHeight
	pr.viewboxX = pr.ctrl.CanvasWindowX
	pr.viewboxY = pr.ctrl.CanvasWindowY
}

func (pr *planeRenderer) process(p *Plane) {
	pr.fixViewport()
	pr.renderModel(p)
	pr.Screen = pr.buf
	pr.buf = newPlaneBuffer()
}

func (pr *planeRenderer) renderModel(p *Plane) {
	for _, v := range p.temporary {
		pr.renderVertex(v, ColorDefault)
	}

	points := make([]*geometry.Point, 0, len(p.defined)+len(p.generated))

	for _, v := range p.defined {
		pr.renderVertex(v, p.color)
		points = append(points, pt(v.X, v.Y))
	}
	for _, v := range p.generated {
		pr.renderVertex(v, p.color)
		points = append(points, pt(v.X, v.Y))
	}

	if len(points) > 0 {
		pr.renderChain(points, p.color)
	}

	if p.util.kind != UtilNone {
		pr.renderUtility(p.util)
	}
}

func (pr *planeRenderer) renderChain(points []*geometry.Point, color string) {
	max := rpt(pr.canvasWidth, pr.canvasHeight)
	firstX, lastY := pr.absX(pr.scaleX(points[0].X)), pr.absY(pr.scaleY(points[len(points)-1].Y))

	pts := make([]*render.Point, 0, len(points)+2)
	pts = append(pts, rpt(firstX, max.Y))
	for i := 0; i < len(points); i++ {
		if i > 0 {
			p1 := points[i-1]
			p2 := points[i]
			pts = append(pts, pr.absPoint(pr.scalePoint(pt(p2.X, p1.Y))))
		}
		pts = append(pts, pr.absPoint(pr.scalePoint(points[i])))
	}
	pts = append(pts, rpt(max.X, lastY))

	for i := 0; i < len(pts)-1; i++ {
		bg := &render.Rect{
			Style: &render.Style{
				Fill:      true,
				FillStyle: util.AlphaHex(color, 60),
			},
			X: pts[i].X, Y: pts[i].Y,
			W: pts[i+1].X - pts[i].X,
			H: max.Y - pts[i].Y,
		}
		pr.buf.Rects.Put(bg)
	}

	chain := render.NewChain(pts...)
	chain.Style = &render.Style{
		LineWidth:   pr.scale(2.0),
		Stroke:      true,
		StrokeStyle: color,
	}
	pr.buf.Chains.Put(chain)
}

func (pr *planeRenderer) renderVertex(v *vertex, color string) {
	point := &render.Circle{
		Style: &render.Style{
			Fill:      true,
			FillStyle: color,
		},
		X: pr.absX(pr.scaleX(v.X - PointRadius)),
		Y: pr.absY(pr.scaleY(v.Y - PointRadius)),
		D: pr.scale(PointRadius * 2),
	}
	if v.isSelected() {
		point.Style.FillStyle = ColorSelected
		pad := &render.Circle{
			Style: &render.Style{
				Fill:      true,
				FillStyle: ColorVertexPad,
			},
			X: pr.absX(pr.scaleX(v.X - PointRadius - PadOffset)),
			Y: pr.absY(pr.scaleY(v.Y - PointRadius - PadOffset)),
			D: pr.scale((PointRadius + PadOffset) * 2),
		}
		pr.buf.Pads.Put(pad)
	}
	pr.buf.Circles.Put(point)
}

func (pr *planeRenderer) renderUtility(u *utility) {
	min, max := u.min, u.max
	if u.kind == UtilRect {
		w, h := max.X-min.X, max.Y-min.Y
		if w == 0 || h == 0 {
			return
		}
		rect := &render.Rect{
			Style: &render.Style{
				LineWidth:   1.0,
				Stroke:      true,
				Fill:        true,
				StrokeStyle: ColorUtility,
				FillStyle:   ColorUtilityShadow,
			},
			X: pr.absX(pr.scaleX(min.X)),
			Y: pr.absY(pr.scaleY(min.Y)),
			W: pr.scale(w), H: pr.scale(h),
		}
		pr.buf.Rects.Put(rect)
	}
}

func (pr *planeRenderer) scale(f float64) float64 {
	return f * pr.zoom
}

func (pr *planeRenderer) scaleX(x float64) float64 {
	return x * pr.zoom
}

func (pr *planeRenderer) scaleY(y float64) float64 {
	return y * pr.zoom
}

func (pr *planeRenderer) scalePoint(p *geometry.Point) *geometry.Point {
	return &geometry.Point{pr.scaleX(p.X), pr.scaleY(p.Y)}
}

func (pr *planeRenderer) absX(x float64) float64 {
	return pr.canvasWidth/2 + pr.viewboxWidth/2 + x
}

func (pr *planeRenderer) absY(y float64) float64 {
	return pr.canvasHeight/2 + pr.viewboxHeight/2 + y
}

func (pr *planeRenderer) absPoint(p *geometry.Point) *render.Point {
	return rpt(pr.canvasWidth/2+pr.viewboxWidth/2+p.X,
		pr.canvasHeight/2+pr.viewboxHeight/2+p.Y)
}

func pt(x, y float64) *geometry.Point {
	return &geometry.Point{x, y}
}

func rpt(x, y float64) *render.Point {
	return &render.Point{x, y}
}

func calcSpacing(room, each float64, count int) float64 {
	return (room - (each * float64(count))) / (float64(count) + 1)
}

func calcCenteringMargin(room, each float64, count int) float64 {
	c := float64(count)
	return (room - ((c-1)*calcSpacing(room, each, count) + c*each)) / 2
}
