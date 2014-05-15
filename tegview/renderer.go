package tegview

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/xlab/teg-workshop/geometry"
	"github.com/xlab/teg-workshop/render"
)

const (
	TextFontNormal = "Georgia"
)

const (
	ColorSelected      = "#b10000"
	ColorDefault       = "#000000"
	ColorComments      = "#7f8c8d"
	ColorUtility       = "#3498db"
	ColorUtilityShadow = "#202980b9"
	ColorControlPoint  = "#f1c40f"
	ColorTransitionPad = "#90bdc3c7"
)

const (
	Thickness     = 6.0
	Padding       = 2.0
	Margin        = 50.0
	TipHeight     = 5.0
	TipSide       = 3.0
	PlaceFontSize = 14.0
	TextFontSize  = 14.0

	BorderTransitionDist    = 3.0
	BorderTransitionTipDist = 2.0
	BorderPlaceDist         = 5.0
	BorderPlaceTipDist      = 3.0
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

type TegBuffer struct {
	Circles *List
	Rects   *List
	Bezier  *List
	Polys   *List
	Texts   *List
	Lines   *List
	Chains  *List
}

func newTegBuffer() *TegBuffer {
	return &TegBuffer{
		Circles: list(make([]interface{}, 0, 256)),
		Rects:   list(make([]interface{}, 0, 256)),
		Bezier:  list(make([]interface{}, 0, 256)),
		Polys:   list(make([]interface{}, 0, 256)),
		Texts:   list(make([]interface{}, 0, 1024)),
		Lines:   list(make([]interface{}, 0, 256)),
		Chains:  list(make([]interface{}, 0, 100)),
	}
}

type tegRenderer struct {
	task   chan interface{}
	ready  chan interface{}
	ctrl   *Ctrl
	buf    *TegBuffer
	Screen *TegBuffer
	Ready  bool

	zoom          float64
	canvasWidth   float64
	canvasHeight  float64
	viewboxWidth  float64
	viewboxHeight float64
}

func newTegRenderer(ctrl *Ctrl) *tegRenderer {
	return &tegRenderer{
		task:   make(chan interface{}, 100),
		ready:  make(chan interface{}, 100),
		Screen: newTegBuffer(),
		buf:    newTegBuffer(),
		ctrl:   ctrl,
		zoom:   1.0,
	}
}

func (tr *tegRenderer) fixViewport() {
	tr.zoom = tr.ctrl.Zoom
	tr.canvasWidth = tr.ctrl.CanvasWidth
	tr.canvasHeight = tr.ctrl.CanvasHeight
	tr.viewboxWidth = tr.ctrl.CanvasWindowWidth
	tr.viewboxHeight = tr.ctrl.CanvasWindowHeight
}

func (tr *tegRenderer) process(model *teg) {
	for {
		<-tr.task
		tr.fixViewport()
		tr.renderModel(model, false)
		tr.Screen = tr.buf
		tr.buf = newTegBuffer()
		tr.ready <- nil
	}
}

func (tr *tegRenderer) renderModel(tg *teg, nested bool) {
	for _, p := range tg.places {
		tr.renderPlace(p)
	}
	for _, t := range tg.transitions {
		if !nested || t.proxy == nil {
			tr.renderTransition(t)

			for i, p := range t.in {
				tr.renderArc(t, p, true, i)
			}
			for i, p := range t.out {
				tr.renderArc(t, p, false, i)
			}
		}
	}
	for _, g := range tg.groups {
		tr.renderModel(g.model, true)
		for t := range g.inputs {
			for i, p := range t.in {
				tr.renderArc(t, p, true, i)
			}
			for i, p := range t.out {
				tr.renderArc(t, p, false, i)
			}
		}
		for t := range g.outputs {
			for i, p := range t.in {
				tr.renderArc(t, p, true, i)
			}
			for i, p := range t.out {
				tr.renderArc(t, p, false, i)
			}
		}
	}
	if !nested && tr.ctrl.ModifierKeyAlt {
		for _, t := range tg.transitions {
			for i, p := range t.in {
				tr.renderConnection(t, p, true, i)
			}
			for i, p := range t.out {
				tr.renderConnection(t, p, false, i)
			}
		}
	}
	if !nested && tg.util.kind != UtilNone {
		tr.renderUtility(tg.util)
	}
}

func (tr *tegRenderer) renderUtility(u *utility) {
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
			X: tr.absX(tr.scale(min.X)),
			Y: tr.absY(tr.scale(min.Y)),
			W: tr.scale(w), H: tr.scale(h),
		}
		tr.buf.Rects.Put(rect)
	} else if u.kind == UtilStroke {
		line := &render.Line{
			Style: &render.Style{
				LineWidth:   1.5,
				Stroke:      true,
				StrokeStyle: ColorUtility,
			},
			Start: tr.absPoint(tr.scalePoint(min)),
			End:   tr.absPoint(tr.scalePoint(max)),
		}
		tr.buf.Lines.Put(line)
	}
}

func (tr *tegRenderer) renderPlace(p *place) {
	x, y := p.X(), p.Y()
	pad := &render.Circle{
		Style: &render.Style{
			LineWidth:   tr.scale(2.0),
			Stroke:      true,
			StrokeStyle: ColorDefault,
		},
		X: tr.absX(tr.scale(x)),
		Y: tr.absY(tr.scale(y)),
		D: tr.scale(p.Width()),
	}
	if p.IsSelected() {
		makeControlPoint := func(cp *controlPoint) *render.Rect {
			xc, yc := cp.X(), cp.Y()
			return &render.Rect{
				Style: &render.Style{
					Fill:      true,
					FillStyle: ColorControlPoint,
				},
				X: tr.absX(tr.scale(xc)),
				Y: tr.absY(tr.scale(yc)),
				W: tr.scale(cp.Width()),
				H: tr.scale(cp.Height()),
			}
		}
		if p.inControl != nil {
			tr.buf.Rects.Put(makeControlPoint(p.inControl))
		}
		if p.outControl != nil {
			tr.buf.Rects.Put(makeControlPoint(p.outControl))
		}
		pad.Style.StrokeStyle = ColorSelected
	}
	tr.buf.Circles.Put(pad)
	tr.renderPlaceValue(x+Padding, y+Padding, p.Width()-Padding*2, p.IsSelected(), p.counter, p.timer)
	if len(p.label) > 0 {
		tr.renderText(x, y+p.Height()+TextFontSize, p.Width(), false, p.label)
	}
}

func (tr *tegRenderer) renderTransition(t *transition) {
	x, y := t.X(), t.Y()
	rect := &render.Rect{
		Style: &render.Style{
			Fill:      true,
			FillStyle: ColorDefault,
		},
		X: tr.absX(tr.scale(x)),
		Y: tr.absY(tr.scale(y)),
		W: tr.scale(t.Width()),
		H: tr.scale(t.Height()),
	}
	if t.IsSelected() {
		var d float64
		if t.horizontal {
			d = t.Width() + 3*Padding
		} else {
			d = t.Height() + 3*Padding
		}
		shadow := &render.Circle{
			Style: &render.Style{
				Fill:      true,
				FillStyle: ColorTransitionPad,
			},
			X: tr.absX(tr.scale(x + (t.Width()-d)/2)),
			Y: tr.absY(tr.scale(y + (t.Height()-d)/2)),
			D: tr.scale(d),
		}
		tr.buf.Circles.Put(shadow)
		rect.Style.FillStyle = ColorSelected
	}
	tr.buf.Rects.Put(rect)
	if len(t.label) > 0 {
		if t.horizontal {
			tr.renderText(x+t.Width()+TextFontSize/2, y+t.Height()/2, t.Height(), true, t.label)
		} else {
			tr.renderText(x, y+t.Height()+TextFontSize, t.Width(), false, t.label)
		}
	}
}

func (tr *tegRenderer) renderArc(t *transition, p *place, inbound bool, index int) {
	thick := 2.0
	var control *controlPoint
	var count int

	if inbound {
		control = p.outControl
		count = len(t.in)
	} else {
		control = p.inControl
		count = len(t.out)
	}
	selected := t.IsSelected() && p.IsSelected()
	xyC1 := control.Center()
	endT := calcBorderPointTransition(t, inbound, count, index)
	endP := calcBorderPointPlace(p, xyC1.X, xyC1.Y)
	var endPointed *end
	if inbound {
		endPointed = endT
	} else {
		endPointed = endP
	}

	var xyC2 *geometry.Point
	if inbound {
		if t.horizontal {
			xyC2 = pt(endT.x, endT.y-Margin)
		} else {
			xyC2 = pt(endT.x-Margin, endT.y)
		}
	} else {
		if t.horizontal {
			xyC2 = pt(endT.x, endT.y+Margin)
		} else {
			xyC2 = pt(endT.x+Margin, endT.y)
		}
	}

	p1 := &geometry.Point{endPointed.xTip, endPointed.yTip}
	p2 := &geometry.Point{p1.X - TipSide*thick, p1.Y + TipHeight*thick}
	p3 := &geometry.Point{p1.X, p1.Y + TipHeight*2/3*thick}
	p4 := &geometry.Point{p1.X + TipSide*thick, p1.Y + TipHeight*thick}
	p1 = p1.Rotate(p1, -endPointed.angle)
	p2 = p2.Rotate(p1, -endPointed.angle)
	p3 = p3.Rotate(p1, -endPointed.angle)
	p4 = p4.Rotate(p1, -endPointed.angle)
	pointer := render.NewPoly(
		tr.absPoint(tr.scalePoint(p1)),
		tr.absPoint(tr.scalePoint(p2)),
		tr.absPoint(tr.scalePoint(p3)),
		tr.absPoint(tr.scalePoint(p4)),
	)
	pointer.Style = &render.Style{Fill: true, FillStyle: ColorDefault}
	if selected {
		pointer.Style.FillStyle = ColorSelected
	}

	curve := &render.BezierCurve{
		Style: &render.Style{
			LineWidth:   tr.scale(thick),
			Stroke:      true,
			StrokeStyle: ColorDefault,
		},
		Start: tr.absPoint(tr.scalePoint(pt(endP.x, endP.y))),
		End:   tr.absPoint(tr.scalePoint(pt(endT.x, endT.y))),
		C1:    tr.absPoint(tr.scalePoint(xyC1)),
		C2:    tr.absPoint(tr.scalePoint(xyC2)),
	}
	if selected {
		curve.Style.StrokeStyle = ColorSelected
	}
	tr.buf.Bezier.Put(curve)
	tr.buf.Polys.Put(pointer)
}

func (tr *tegRenderer) renderConnection(t *transition, p *place, inbound bool, index int) {
	var control *controlPoint
	var count int
	if inbound {
		control = p.outControl
		count = len(t.in)
	} else {
		control = p.inControl
		count = len(t.out)
	}
	xyC1 := control.Center()
	endT := calcBorderPointTransition(t, inbound, count, index)
	endP := calcBorderPointPlace(p, xyC1.X, xyC1.Y)
	var xyC2 *geometry.Point
	var start, end *geometry.Point
	if inbound {
		if t.horizontal {
			xyC2 = pt(endT.x, endT.y-Margin)
		} else {
			xyC2 = pt(endT.x-Margin, endT.y)
		}
		start = pt(endP.x, endP.y)
		end = pt(endT.xTip, endT.yTip)
	} else {
		if t.horizontal {
			xyC2 = pt(endT.x, endT.y+Margin)
		} else {
			xyC2 = pt(endT.x+Margin, endT.y)
		}
		start = pt(endT.x, endT.y)
		end = pt(endP.xTip, endP.yTip)
		xyC1, xyC2 = xyC2, xyC1
	}
	chain := render.NewChain(
		tr.absPoint(tr.scalePoint(start)),
		tr.absPoint(tr.scalePoint(xyC1)),
		tr.absPoint(tr.scalePoint(xyC2)),
		tr.absPoint(tr.scalePoint(end)),
	)
	chain.Style = &render.Style{
		LineWidth:   tr.scale(1.0),
		Stroke:      true,
		StrokeStyle: ColorUtility,
	}
	tr.buf.Chains.Put(chain)
}

func (tr *tegRenderer) renderText(x, y, room float64, valign bool, text string) {
	// commentary sign
	text = fmt.Sprintf("// %s", text)

	// detect all substrings with  1 <= length <= 16
	chunks := regexp.MustCompile(".{1,16}").FindAllString(text, -1)

	var offset float64
	for i := range chunks {
		subchunks := strings.Split(chunks[i], "\n")
		for _, str := range subchunks {
			label := &render.Text{
				Style: &render.Style{
					LineWidth: tr.scale(1.0),
					Fill:      true,
					FillStyle: ColorComments,
				},
				Font:     TextFontNormal,
				FontSize: tr.scale(TextFontSize),
				Oblique:  true,
				Label:    str,
			}
			if valign {
				label.X = tr.absX(tr.scale(x))
				label.Y = tr.absY(tr.scale(y + room/2 + offset))
			} else {
				label.Align = render.TextAlignCenter
				label.X = tr.absX(tr.scale(x + room/2))
				label.Y = tr.absY(tr.scale(y + offset + 2*Padding))
			}
			tr.buf.Texts.Put(label)
			offset += TextFontSize + Padding
		}
	}
}

func (tr *tegRenderer) renderDotRow(x, y, room float64, selected bool, count int) {
	hspace := calcSpacing(room, Thickness, count)
	hmargin := calcCenteringMargin(room, Thickness, count)
	style := &render.Style{
		LineWidth:   tr.scale(1.0),
		StrokeStyle: ColorDefault,
		FillStyle:   ColorDefault,
		Stroke:      true,
		Fill:        true,
	}
	if selected {
		style.StrokeStyle = ColorSelected
		style.FillStyle = ColorSelected
		for i := 0; i < count; i++ {
			tr.buf.Circles.Put(&render.Circle{
				Style: style,
				X:     tr.absX(tr.scale(x + (hmargin + float64(i)*(hspace+Thickness)))),
				Y:     tr.absY(tr.scale(y)),
				D:     tr.scale(Thickness),
			})
		}
	} else {
		for i := 0; i < count; i++ {
			tr.buf.Circles.Put(&render.Circle{
				Style: style,
				X:     tr.absX(tr.scale(x + (hmargin + float64(i)*(hspace+Thickness)))),
				Y:     tr.absY(tr.scale(y)),
				D:     tr.scale(Thickness),
			})
		}
	}
}

func (tr *tegRenderer) renderDotColumn(x, y, room float64, selected bool, count int) {
	vspace := calcSpacing(room, Thickness, count)
	vmargin := calcCenteringMargin(room, Thickness, count)
	style := &render.Style{
		LineWidth:   tr.scale(1.0),
		StrokeStyle: ColorDefault,
		FillStyle:   ColorDefault,
		Stroke:      true,
		Fill:        true,
	}
	if selected {
		style.StrokeStyle = ColorSelected
		style.FillStyle = ColorSelected
	}
	for i := 0; i < count; i++ {
		tr.buf.Circles.Put(&render.Circle{
			Style: style,
			X:     tr.absX(tr.scale(x)),
			Y:     tr.absY(tr.scale(y + (vmargin + float64(i)*(vspace+Thickness)))),
			D:     tr.scale(Thickness),
		})
	}

}

func (tr *tegRenderer) renderBarRow(x, y, room, Thickness float64, selected bool, count int) {
	w, h := Thickness/2, 4*Thickness
	hspace := calcSpacing(room, w, count)
	style := &render.Style{
		LineWidth:   tr.scale(1.0),
		StrokeStyle: ColorDefault,
		FillStyle:   ColorDefault,
		Stroke:      true,
		Fill:        true,
	}
	if selected {
		style.StrokeStyle = ColorSelected
		style.FillStyle = ColorSelected
	}
	for i, j := 1, 0; j < count; i, j = i+1, j+1 {
		tr.buf.Rects.Put(&render.Rect{
			Style: style,

			X: tr.absX(tr.scale(x + (float64(i)*hspace + float64(j)*w))),
			Y: tr.absY(tr.scale(y)),
			W: tr.scale(w),
			H: tr.scale(h),
		})
	}
}

func (tr *tegRenderer) renderPlaceValue(x, y, room float64, selected bool, counter, timer int) {
	switch {
	case timer < 1 && counter > 0 && counter <= 9:
		var rows int
		var vspace, vmargin float64
		switch counter {
		case 3:
			rows = 2
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows) - 1.5 // ∆ hack
			tr.renderDotRow(x, y+vmargin, room, selected, 1)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 2)
		case 4:
			rows = 2
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 2)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 2)
		case 6:
			rows = 2
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 3)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 3)
		case 5:
			rows = 3
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 2)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 1)
			tr.renderDotRow(x, y+(vmargin+2*Thickness+2*vspace), room, selected, 2)
		case 7:
			rows = 3
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 3)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 1)
			tr.renderDotRow(x, y+(vmargin+2*Thickness+2*vspace), room, selected, 3)
		case 8:
			rows = 3
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 3)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 2)
			tr.renderDotRow(x, y+(vmargin+2*Thickness+2*vspace), room, selected, 3)
		case 9:
			rows = 3
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 3)
			tr.renderDotRow(x, y+(vmargin+Thickness+vspace), room, selected, 3)
			tr.renderDotRow(x, y+(vmargin+2*Thickness+2*vspace), room, selected, 3)
		default:
			rows = 1
			vspace = calcSpacing(room, Thickness, rows)
			vmargin = calcCenteringMargin(room, Thickness, rows)
			tr.renderDotRow(x, y+vmargin, room, selected, 1)
		}
	case timer > 0 && timer <= 4 && counter <= 3:
		rows := 1
		vmargin := calcCenteringMargin(room, 4*Thickness, rows)
		if counter > 0 {
			hspace := calcSpacing(room, Thickness, timer+1)
			hmargin := calcCenteringMargin(room, Thickness, timer+1)
			gap := hmargin + hspace
			tr.renderDotColumn(x+hmargin, y+Thickness, room-2*Thickness, selected, counter)
			tr.renderBarRow(x+gap, y+vmargin, room-gap, Thickness, selected, timer)
		} else {
			tr.renderBarRow(x, y+vmargin, room, Thickness, selected, timer)
		}
	case timer > 0, counter > 0:
		yPos := y + calcCenteringMargin(room, tr.scale(1.0), 1)
		separator := &render.Line{
			Style: &render.Style{
				LineWidth:   tr.scale(1.0),
				Stroke:      true,
				StrokeStyle: ColorDefault,
			},
			Start: tr.absPoint(tr.scalePoint(pt(x+Padding, yPos))),
			End:   tr.absPoint(tr.scalePoint(pt(x+room-Padding, yPos))),
		}
		if selected {
			separator.Style.StrokeStyle = ColorSelected
		}
		tr.buf.Lines.Put(separator)

		if counter > 0 && counter <= 4 {
			voffset := calcCenteringMargin(room/2, Thickness, 1)
			tr.renderDotRow(x, yPos-voffset-Thickness/2, room, selected, counter)
		} else if counter > 4 {
			voffset := calcCenteringMargin(room/2, PlaceFontSize, 1)
			label := &render.Text{
				Style: &render.Style{
					LineWidth:   tr.scale(1.0),
					Stroke:      true,
					Fill:        true,
					StrokeStyle: ColorDefault,
					FillStyle:   ColorDefault,
				},
				X:        tr.absX(tr.scale(x + room/2)),
				Y:        tr.absY(tr.scale(yPos - voffset - Padding)),
				Align:    render.TextAlignCenter,
				Font:     TextFontNormal,
				FontSize: tr.scale(TextFontSize),
				Label:    fmt.Sprintf("%v", counter),
			}
			if selected {
				label.Style.StrokeStyle = ColorSelected
				label.Style.FillStyle = ColorSelected
			}
			tr.buf.Texts.Put(label)
		}

		if timer > 0 && timer <= 4 {
			voffset := calcCenteringMargin(room/2, Thickness*2, 1)
			tr.renderBarRow(x, yPos+voffset, room, Thickness/2, selected, timer)
		} else if timer > 4 {
			voffset := calcCenteringMargin(room/2, PlaceFontSize, 1)
			label := &render.Text{
				Style: &render.Style{
					LineWidth:   tr.scale(1.0),
					Stroke:      true,
					Fill:        true,
					StrokeStyle: ColorDefault,
					FillStyle:   ColorDefault,
				},
				X:        tr.absX(tr.scale(x + room/2)),
				Y:        tr.absY(tr.scale(yPos + voffset + PlaceFontSize - 1.8)),
				Align:    render.TextAlignCenter,
				Font:     TextFontNormal,
				FontSize: tr.scale(TextFontSize),
				Label:    fmt.Sprintf("%v", timer),
			}
			if selected {
				label.Style.StrokeStyle = ColorSelected
				label.Style.FillStyle = ColorSelected
			}
			tr.buf.Texts.Put(label)
		}
	}
}

func (tr *tegRenderer) scale(f float64) float64 {
	return f * tr.zoom
}

func (tr *tegRenderer) scalePoint(p *geometry.Point) *geometry.Point {
	return &geometry.Point{p.X * tr.zoom, p.Y * tr.zoom}
}

func (tr *tegRenderer) absX(x float64) float64 {
	return tr.canvasWidth/2 + tr.viewboxWidth/2 + x
}

func (tr *tegRenderer) absY(y float64) float64 {
	return tr.canvasHeight/2 + tr.viewboxHeight/2 + y
}

func (tr *tegRenderer) absPoint(p *geometry.Point) *render.Point {
	return rpt(tr.canvasWidth/2+tr.viewboxWidth/2+p.X,
		tr.canvasHeight/2+tr.viewboxHeight/2+p.Y)
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

type end struct {
	x, y, angle, xTip, yTip float64
}

func calcBorderPointPlace(p *place, x, y float64) *end {
	angle := math.Atan2(x-p.Center().X, y-p.Center().Y)
	radius := p.Width() / 2
	dxTip := (radius + BorderPlaceTipDist) * math.Sin(angle)
	dyTip := (radius + BorderPlaceTipDist) * math.Cos(angle)
	dx := (radius + BorderPlaceDist) * math.Sin(angle)
	dy := (radius + BorderPlaceDist) * math.Cos(angle)
	return &end{
		x: p.Center().X + dx, y: p.Center().Y + dy,
		xTip: p.Center().X + dxTip, yTip: p.Center().Y + dyTip,
		angle: angle,
	}
}

func calcBorderPointTransition(t *transition, inbound bool, count, index int) *end {
	thick := 2.0
	var x, y float64

	end := new(end)
	if t.horizontal && inbound {
		x = t.X()
		y = t.Y() - BorderTransitionDist
	} else if inbound {
		x = t.X() - BorderTransitionDist
		y = t.Y()
	} else if t.horizontal {
		x = t.X()
		y = t.Y() + t.Height() + BorderTransitionDist
	} else {
		x = t.X() + t.Width() + BorderTransitionDist
		y = t.Y()
	}
	if t.horizontal {
		space := calcSpacing(t.Width(), thick, count)
		margin := calcCenteringMargin(t.Width(), thick, count)
		dx := margin + float64(index)*(thick+space)
		end.x, end.y = x+dx, y
		if inbound {
			end.angle = math.Pi
			end.xTip = x + dx
			end.yTip = y + BorderTransitionTipDist
		} else {
			end.angle = math.Pi
			end.xTip = x + dx
			end.yTip = y - BorderTransitionTipDist
		}
	} else {
		space := calcSpacing(t.Height(), thick, count)
		margin := calcCenteringMargin(t.Height(), thick, count)
		dy := margin + float64(index)*(thick+space)
		end.x, end.y = x, y+dy
		if inbound {
			end.angle = -math.Pi / 2
			end.xTip = x + BorderTransitionTipDist
			end.yTip = y + dy
		} else {
			end.angle = math.Pi / 2
			end.xTip = x - BorderTransitionTipDist
			end.yTip = y + dy
		}
	}
	return end
}
