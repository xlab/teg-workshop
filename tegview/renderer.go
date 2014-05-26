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
	ColorSelected        = "#b10000"
	ColorDefault         = "#000000"
	ColorTransitionIO    = "#2980b9"
	ColorComments        = "#7f8c8d"
	ColorGroupFrame      = "#34495e"
	ColorGroupBg         = "#2034495e"
	ColorGroupBgSelected = "#20e74c3c"
	ColorUtility         = "#3498db"
	ColorUtilityShadow   = "#202980b9"
	ColorControlPoint    = "#f1c40f"
	ColorTransitionPad   = "#90bdc3c7"
)

const (
	Thickness     = 6.0
	Padding       = 2.0
	Margin        = 50.0
	TipHeight     = 5.0
	TipSide       = 3.0
	PlaceFontSize = 14.0
	TextFontSize  = 14.0
	GroupFontSize = 18.0
	GroupFrameR   = 10.0

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
	RRects  *List
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
		RRects:  list(make([]interface{}, 0, 100)),
		Bezier:  list(make([]interface{}, 0, 256)),
		Polys:   list(make([]interface{}, 0, 256)),
		Texts:   list(make([]interface{}, 0, 256)),
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
	viewboxX      float64
	viewboxY      float64

	relateiveGlobalCenter *geometry.Point
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
	tr.viewboxX = tr.ctrl.CanvasWindowX
	tr.viewboxY = tr.ctrl.CanvasWindowY
}

func (tr *tegRenderer) process(model *teg) {
	<-tr.task
	tr.fixViewport()
	tr.renderModel(model, pt(0, 0), false)
	tr.Screen = tr.buf
	tr.buf = newTegBuffer()
	tr.ready <- nil
}

func (tr *tegRenderer) renderModel(tg *teg, shift *geometry.Point, nested bool) {
	for _, g := range tg.groups {
		tr.renderGroup(g, shift, nested)
	}
	for _, p := range tg.places {
		tr.renderPlace(p, shift, nested)
	}
	for _, t := range tg.transitions {
		if !nested || t.kind == TransitionInternal {
			tr.renderTransition(t, shift, nested)
			for i, p := range t.in {
				tr.renderArc(t, shift, p, shift, true, i)
			}
			for i, p := range t.out {
				tr.renderArc(t, shift, p, shift, false, i)
			}
		}
	}
	if !nested && tr.ctrl.ModifierKeyAlt {
		for _, t := range tg.transitions {
			for i, p := range t.in {
				tr.renderConnection(t, shift, p, shift, true, i)
			}
			for i, p := range t.out {
				tr.renderConnection(t, shift, p, shift, false, i)
			}
		}
	}
	if !nested && tg.util.kind != UtilNone {
		tr.renderUtility(tg.util)
	}
}

func (tr *tegRenderer) renderGroup(g *group, shiftRoot *geometry.Point, nested bool) {
	gx, gy := g.X()+shiftRoot.X, g.Y()+shiftRoot.Y
	frame := &render.RoundedRect{
		Style: &render.Style{
			LineWidth:   tr.scale(2.0),
			Stroke:      true,
			StrokeStyle: ColorGroupFrame,
			Fill:        true,
			FillStyle:   ColorGroupBg,
		},
		X: tr.absX(tr.scaleX(gx)),
		Y: tr.absY(tr.scaleY(gy)),
		W: tr.scale(g.Width()),
		H: tr.scale(g.Height()),
		R: tr.scale(GroupFrameR),
	}
	if g.IsSelected() {
		frame.Style.StrokeStyle = ColorSelected
		frame.Style.FillStyle = ColorGroupBgSelected
	}
	tr.buf.RRects.Put(frame)

	// Internals rendering code
	center := g.Center()
	center.X, center.Y = center.X+shiftRoot.X, center.Y+shiftRoot.Y
	shift := calcItemsShift(center, g.model.Items())

	for _, t := range g.inputs {
		tr.renderTransition(t, shiftRoot, nested)
		if len(t.label) > 0 && !nested {
			tr.renderIOText(t, true)
		}
		if !g.folded {
			for i, p := range t.out {
				if nested {
					tr.renderArc(t, shiftRoot, p, shift, false, i)
				} else {
					tr.renderArc(t, pt(0, 0), p, shift, false, i)
				}
			}
		}
		for i, p := range t.in {
			if nested {
				tr.renderArc(t, shiftRoot, p, shiftRoot, true, i)
			} else {
				tr.renderArc(t, pt(0, 0), p, shiftRoot, true, i)
			}
			if !nested && tr.ctrl.ModifierKeyAlt {
				tr.renderConnection(t, pt(0, 0), p, pt(0, 0), true, i)
			}
		}
	}
	for _, t := range g.outputs {
		tr.renderTransition(t, shiftRoot, nested)
		if len(t.label) > 0 && !nested {
			tr.renderIOText(t, false)
		}
		if !g.folded {
			for i, p := range t.in {
				if nested {
					tr.renderArc(t, shiftRoot, p, shift, true, i)
				} else {
					tr.renderArc(t, pt(0, 0), p, shift, true, i)
				}
			}
		}
		for i, p := range t.out {
			if nested {
				tr.renderArc(t, shiftRoot, p, shiftRoot, false, i)
			} else {
				tr.renderArc(t, pt(0, 0), p, pt(0, 0), false, i)
			}
			if !nested && tr.ctrl.ModifierKeyAlt {
				tr.renderConnection(t, pt(0, 0), p, pt(0, 0), false, i)
			}
		}
	}
	if !g.folded {
		tr.renderModel(g.model, shift, true)

		// Render label
		if len(g.label) < 1 || nested {
			return
		}
		room := g.Width() - GroupMargin*2
		breaklen := int(math.Max(room/TextFontSize, 16))
		cfg := textConfig{
			x: gx + GroupMargin, y: gy + g.Height() + TextFontSize,
			room: room, color: ColorComments, align: render.TextAlignCenter,
			text: fmt.Sprintf("// %s", g.label), breaklen: breaklen, oblique: true,
		}
		tr.renderText(&cfg)
		return
	}

	// Render big label
	if len(g.label) < 1 {
		return
	}
	room := g.Width() - GroupMargin*2
	breaklen := int(math.Max(room/GroupFontSize, 16))
	text := g.label
	_, theight := calcTextFragments(text, GroupFontSize, breaklen)
	vmargin := calcCenteringMargin(g.Height(), theight, 1)

	cfg := textConfig{
		x: gx + GroupMargin, y: gy + GroupMargin + vmargin,
		room: room, color: ColorGroupFrame, align: render.TextAlignCenter,
		text: text, breaklen: breaklen, fontsize: GroupFontSize,
	}
	if g.IsSelected() {
		cfg.color = ColorSelected
	}
	tr.renderText(&cfg)
}

func (tr *tegRenderer) renderIOText(t *transition, input bool) {
	cfg := textConfig{
		breaklen: 16, color: ColorComments, oblique: true,
	}
	_, theight := calcTextFragments("// "+t.label, TextFontSize, 16)
	if t.horizontal {
		if input {
			cfg.x = t.X() + t.Width()/2 + theight
			cfg.y = t.Y() - 3*Padding
			cfg.align = render.TextAlignRight
			cfg.text = t.label + " //"
		} else {
			cfg.x = t.X() - t.Width() + 3*Padding
			cfg.y = t.Y() + TextFontSize
			cfg.align = render.TextAlignLeft
			cfg.text = "// " + t.label
		}
		cfg.vertical = true
		cfg.room = GroupIOSpacing
		cfg.valign = true
		tr.renderText(&cfg)
	} else {
		if input {
			cfg.x = t.X() - 3*Padding
			cfg.y = t.Y() + t.Height()
			cfg.align = render.TextAlignRight
			cfg.text = "// " + t.label
		} else {
			cfg.x = t.X() + t.Width() + 3*Padding
			cfg.y = t.Y() - theight
			cfg.align = render.TextAlignLeft
			cfg.text = t.label + " //"
		}
		cfg.room = GroupIOSpacing
		cfg.valign = true
		tr.renderText(&cfg)
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
			X: tr.absX(tr.scaleX(min.X)),
			Y: tr.absY(tr.scaleY(min.Y)),
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

func (tr *tegRenderer) renderPlace(p *place, shift *geometry.Point, nested bool) {
	x, y := p.X()+shift.X, p.Y()+shift.Y
	pad := &render.Circle{
		Style: &render.Style{
			LineWidth:   tr.scale(2.0),
			Stroke:      true,
			StrokeStyle: ColorDefault,
		},
		X: tr.absX(tr.scaleX(x)),
		Y: tr.absY(tr.scaleY(y)),
		D: tr.scale(p.Width()),
	}
	if p.IsSelected() {
		makeControlPoint := func(cp *controlPoint) *render.Rect {
			xc, yc := cp.X()+shift.X, cp.Y()+shift.Y
			return &render.Rect{
				Style: &render.Style{
					Fill:      true,
					FillStyle: ColorControlPoint,
				},
				X: tr.absX(tr.scaleX(xc)),
				Y: tr.absY(tr.scaleY(yc)),
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
	if len(p.label) > 0 && !nested {
		cfg := textConfig{
			x: x, y: y + p.Height() + TextFontSize,
			room: p.Width(), color: ColorComments,
			text:     fmt.Sprintf("// %s", p.label),
			breaklen: 16, oblique: true, align: render.TextAlignCenter,
		}
		tr.renderText(&cfg)
	}
}

func (tr *tegRenderer) renderTransition(t *transition, shift *geometry.Point, nested bool) {
	x, y := t.X()+shift.X, t.Y()+shift.Y
	rect := &render.Rect{
		Style: &render.Style{
			Fill:      true,
			FillStyle: ColorDefault,
		},
		X: tr.absX(tr.scaleX(x)),
		Y: tr.absY(tr.scaleY(y)),
		W: tr.scale(t.Width()),
		H: tr.scale(t.Height()),
	}
	var knob *render.Rect
	kw, kh := Margin/5, Margin/17
	if t.proxy == nil && t.kind == TransitionInput {
		rect.Style.FillStyle = ColorTransitionIO
		knob = &render.Rect{
			Style: &render.Style{
				Fill:      true,
				FillStyle: ColorTransitionIO,
			},
		}
		if t.horizontal {
			kw, kh = kh, kw
			knob.X = tr.scaleX(t.Center().X + shift.X - kw/2)
			knob.Y = tr.scaleY(t.Center().Y + shift.Y - kh)
			knob.W = tr.scale(kw)
			knob.H = tr.scale(kh)
		} else {
			knob.X = tr.scaleX(t.Center().X + shift.X - kw)
			knob.Y = tr.scaleY(t.Center().Y + shift.Y - kh/2)
			knob.W = tr.scale(kw)
			knob.H = tr.scale(kh)
		}
		knob.X, knob.Y = tr.absX(knob.X), tr.absY(knob.Y)
	} else if t.proxy == nil && t.kind == TransitionOutput {
		rect.Style.FillStyle = ColorTransitionIO
		knob = &render.Rect{
			Style: &render.Style{
				Fill:      true,
				FillStyle: ColorTransitionIO,
			},
		}
		if t.horizontal {
			kw, kh = kh, kw
			knob.X = tr.scaleX(t.Center().X + shift.X - kw/2)
			knob.Y = tr.scaleY(t.Center().Y + shift.Y)
			knob.W = tr.scale(kw)
			knob.H = tr.scale(kh)
		} else {
			knob.X = tr.scaleX(t.Center().X + shift.X)
			knob.Y = tr.scaleY(t.Center().Y + shift.Y - kh/2)
			knob.W = tr.scale(kw)
			knob.H = tr.scale(kh)
		}
		knob.X, knob.Y = tr.absX(knob.X), tr.absY(knob.Y)
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
			X: tr.absX(tr.scaleX(x + (t.Width()-d)/2)),
			Y: tr.absY(tr.scaleY(y + (t.Height()-d)/2)),
			D: tr.scale(d),
		}
		tr.buf.Circles.Put(shadow)
		rect.Style.FillStyle = ColorSelected
		if knob != nil {
			knob.Style.FillStyle = ColorSelected
		}
	}
	if knob != nil {
		tr.buf.Rects.Put(knob)
	}
	tr.buf.Rects.Put(rect)
	if len(t.label) > 0 && !nested && t.proxy == nil {
		cfg := textConfig{
			text:     fmt.Sprintf("// %s", t.label),
			breaklen: 16, color: ColorComments, oblique: true,
			align: render.TextAlignCenter,
		}
		if t.horizontal {
			cfg.x = x + t.Width() + TextFontSize/2
			cfg.y = y + t.Height()/2
			cfg.room = t.Height()
			cfg.valign = true
			cfg.align = render.TextAlignLeft
			tr.renderText(&cfg)
		} else {
			cfg.x = x
			cfg.y = y + t.Height() + TextFontSize
			cfg.room = t.Width()
			tr.renderText(&cfg)
		}
	}
}

func (tr *tegRenderer) renderArc(t *transition, shiftT *geometry.Point,
	p *place, shiftP *geometry.Point, inbound bool, index int) {
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
	xyC1.X, xyC1.Y = xyC1.X+shiftP.X, xyC1.Y+shiftP.Y
	endT := calcBorderPointTransition(t, shiftT, inbound, count, index)
	endP := calcBorderPointPlace(p, shiftP, xyC1.X, xyC1.Y)
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

func (tr *tegRenderer) renderConnection(t *transition, shiftT *geometry.Point,
	p *place, shiftP *geometry.Point, inbound bool, index int) {
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
	xyC1.X, xyC1.Y = xyC1.X+shiftP.X, xyC1.Y+shiftP.Y
	endT := calcBorderPointTransition(t, shiftT, inbound, count, index)
	endP := calcBorderPointPlace(p, shiftP, xyC1.X, xyC1.Y)
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

func calcTextFragments(text string, fontsize float64, breaklen int) ([]string, float64) {
	array := make([]string, 0, 32)
	// detect all substrings with  1 <= length <= N
	rx := regexp.MustCompile(fmt.Sprintf(".{1,%d}", breaklen))
	chunks := rx.FindAllString(text, -1)

	var offset float64
	for i := range chunks {
		subchunks := strings.Split(chunks[i], "\n")
		for _, str := range subchunks {
			array = append(array, str)
			offset += fontsize + Padding
		}
	}
	return array, offset
}

type textConfig struct {
	fontsize   float64
	x, y, room float64
	oblique    bool
	valign     bool
	vertical   bool
	color      string
	text       string
	align      string
	breaklen   int
}

func (tr *tegRenderer) renderText(cfg *textConfig) {
	if cfg.fontsize < 8 {
		cfg.fontsize = TextFontSize
	}
	if cfg.breaklen < 1 {
		cfg.breaklen = 16
	}
	strings, _ := calcTextFragments(cfg.text, cfg.fontsize, cfg.breaklen)
	var offset float64
	for _, str := range strings {
		label := &render.Text{
			Style: &render.Style{
				LineWidth: tr.scale(1.0),
				Fill:      true,
				FillStyle: cfg.color,
			},
			Font:     TextFontNormal,
			FontSize: tr.scale(cfg.fontsize),
			Oblique:  cfg.oblique,
			Label:    str,
			Align:    cfg.align,
			Vertical: cfg.vertical,
		}
		if cfg.vertical {
			if cfg.valign {
				label.X = tr.absX(tr.scaleX(cfg.x + cfg.room/2 - offset))
				label.Y = tr.absY(tr.scaleY(cfg.y))
			} else {
				label.X = tr.absX(tr.scaleX(cfg.x - offset - 2*Padding))
				label.Y = tr.absY(tr.scaleY(cfg.y + cfg.room/2))
			}
		} else {
			if cfg.valign {
				label.X = tr.absX(tr.scaleX(cfg.x))
				label.Y = tr.absY(tr.scaleY(cfg.y + cfg.room/2 + offset))
			} else {
				label.X = tr.absX(tr.scaleX(cfg.x + cfg.room/2))
				label.Y = tr.absY(tr.scaleY(cfg.y + offset + 2*Padding))
			}
		}
		tr.buf.Texts.Put(label)
		offset += cfg.fontsize + Padding
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
				X:     tr.absX(tr.scaleX(x + (hmargin + float64(i)*(hspace+Thickness)))),
				Y:     tr.absY(tr.scaleY(y)),
				D:     tr.scale(Thickness),
			})
		}
	} else {
		for i := 0; i < count; i++ {
			tr.buf.Circles.Put(&render.Circle{
				Style: style,
				X:     tr.absX(tr.scaleX(x + (hmargin + float64(i)*(hspace+Thickness)))),
				Y:     tr.absY(tr.scaleY(y)),
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
			X:     tr.absX(tr.scaleX(x)),
			Y:     tr.absY(tr.scaleY(y + (vmargin + float64(i)*(vspace+Thickness)))),
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

			X: tr.absX(tr.scaleX(x + (float64(i)*hspace + float64(j)*w))),
			Y: tr.absY(tr.scaleY(y)),
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
				X:        tr.absX(tr.scaleX(x + room/2)),
				Y:        tr.absY(tr.scaleY(yPos - voffset - Padding)),
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
				X:        tr.absX(tr.scaleX(x + room/2)),
				Y:        tr.absY(tr.scaleY(yPos + voffset + PlaceFontSize - 1.8)),
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

func (tr *tegRenderer) scaleX(x float64) float64 {
	regX := tr.viewboxX - tr.canvasWidth/2
	vecX := (regX - x) * tr.zoom
	return regX - vecX
}

func (tr *tegRenderer) scaleY(y float64) float64 {
	regY := tr.viewboxY - tr.canvasHeight/2
	vecY := (regY - y) * tr.zoom
	return regY - vecY
}

func (tr *tegRenderer) scalePoint(p *geometry.Point) *geometry.Point {
	return &geometry.Point{tr.scaleX(p.X), tr.scaleY(p.Y)}
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

func calcBorderPointPlace(p *place, shift *geometry.Point, x, y float64) *end {
	px, py := p.Center().X+shift.X, p.Center().Y+shift.Y
	angle := math.Atan2(x-px, y-py)
	radius := p.Width() / 2
	dxTip := (radius + BorderPlaceTipDist) * math.Sin(angle)
	dyTip := (radius + BorderPlaceTipDist) * math.Cos(angle)
	dx := (radius + BorderPlaceDist) * math.Sin(angle)
	dy := (radius + BorderPlaceDist) * math.Cos(angle)
	return &end{
		x: px + dx, y: py + dy,
		xTip: px + dxTip, yTip: py + dyTip,
		angle: angle,
	}
}

func calcBorderPointTransition(t *transition, shift *geometry.Point, inbound bool, count, index int) *end {
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
	x, y = x+shift.X, y+shift.Y
	if t.horizontal {
		space := calcSpacing(t.Width(), thick, count)
		margin := calcCenteringMargin(t.Width(), 1.0, count)
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
		margin := calcCenteringMargin(t.Height(), 1.0, count)
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
