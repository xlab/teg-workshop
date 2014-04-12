var sel_color = "#b10000"
var def_color = "#000000"

function render(cv, region, zoom, model) {
    var x0 = region.x
    var y0 = region.y
    var w = region.width
    var h = region.height
    var ctx = cv.getContext("2d")

    ctx.clearRect(x0, y0, w, h)

    for(var idx in model.places) {
        renderPlace(ctx, zoom, model.places[idx])
    }

    for(var idx in model.transitions) {
        renderTransition(ctx, zoom, model.transitions[idx])
    }
    
    for(var idx in model.arcs) {
        var arc = model.arcs[idx]
        var start
        var end
        if(arc.start.type === "place") {
            start = model.places[arc.start.id]
            end = model.transitions[arc.end.id]
        } else if (arc.start.type === "transition") {
            start = model.transitions[arc.start.id]
            end = model.places[arc.end.id]
        }

        renderArc(ctx, zoom, start, end, arc.index)
    }
}

function renderArc(ctx, zoom, start, end, index) {
    var transition = start.transition ? start : end
    var place = start.place ? start : end

    var xy0 = absCoord(ctx, start.x*zoom, start.y*zoom)
    var xy1 = absCoord(ctx, end.x*zoom, end.y*zoom)
    var cxy0 = absCoord(ctx, start.control.x*zoom, start.control.y*zoom)
    var cxy1 = absCoord(ctx, end.control.x*zoom, end.control.y*zoom)

    var wh0 = calcDimensions(zoom, start)
    var wh1 = calcDimensions(zoom, end)
    var cwh0 = {"w":10, "h":10}
    var cwh1 = {"w":10, "h":10}

    var x0 = xy0[0]
    var y0 = xy0[1]
    var x1 = xy1[0]
    var y1 = xy1[1]

    var cx0 = cxy0[0]
    var cy0 = cxy0[1]
    var cx1 = cxy1[0]
    var cy1 = cxy1[1]

    var tx, ty, tcx, tcy, twh
    var px, py, pcx, pcy, pwh

    if(start.transition) {
        tx = x0
        px = x1
        ty = y0
        py = y1
        tcx = cx0
        pcx = cx1
        tcy = cy0
        pcy = cy1
        twh = wh0
        pwh = wh1
    } else {
        tx = x1
        px = x0
        ty = y1
        py = y0
        tcx = cx1
        pcx = cx0
        tcy = cy1
        pcy = cy0
        twh = wh1
        pwh = wh0
    }

    ctx.fillStyle = "#ff0000"
    ctx.rect(cx0 - 5, cy0 - 5, 10, 10)
    ctx.fill()
    ctx.reset()
    ctx.fillStyle = "#00ff00"
    ctx.rect(cx1 - 5, cy1 - 5, 10, 10)
    ctx.fill()

    var thick = 2.0 * zoom

    // collect variables to draw transition-end of bezier
    var data_t = { "x": tx - twh.w/2, "y": ty - twh.h/2, "cx": tx, "cy": ty, "w": twh.w, "h": twh.h }
    var inbound = start.place && end.transition ? true : false
    var count = inbound ? transition.in : transition.out
    var horizontal = transition.horizontal
    // calc the point near transition
    var xyA_t = calcBorderPointTransition(zoom, data_t, tcx, tcy, horizontal, thick, inbound, count, index)
    // calc the point near place
    var xyA_p = calcBorderPointPlace(zoom, pwh.w, px, py, pcx, pcy)

    // get the points to draw
    var startXYA = start.place ? xyA_p : xyA_t
    var endXYA = end.place ? xyA_p : xyA_t

    // nice ∆ pointer
    var p1X = endXYA.x
    var p1Y = endXYA.y
    var p2X = p1X - 3
    var p2Y = p1Y + 5
    var p3X = p1X
    var p3Y = p1Y + (5 * 2/3)
    var p4X = p1X + 3
    var p4Y = p1Y + 5

    // rotated ∆
    var p1 = calcRotatePoint(p1X, p1Y, p1X, p1Y, -endXYA.angle)
    var p2 = calcRotatePoint(p2X, p2Y, p1X, p1Y, -endXYA.angle)
    var p3 = calcRotatePoint(p3X, p3Y, p1X, p1Y, -endXYA.angle)
    var p4 = calcRotatePoint(p4X, p4Y, p1X, p1Y, -endXYA.angle)

    drawPointedBezierCurve(ctx, zoom, thick, start.selected && end.selected,
                           {"x0":startXYA.x, "y0":startXYA.y, "x1":endXYA.x, "y1":endXYA.y,
                               "cx0":cx0, "cy0":cy0, "cx1":cx1, "cy1":cy1},
                           {"p1":p1, "p2":p2, "p3":p3, "p4":p4})
    ctx.reset()
}

function calcRotatePoint(pX, pY, oX, oY, angle) {
    return [oX + (pX - oX) * Math.cos(angle) - (pY - oY) * Math.sin(angle),
            oY + (pX - oX) * Math.sin(angle) + (pY - oY) * Math.cos(angle)]
}

function drawPointedBezierCurve(ctx, zoom, thick, selected, b, p) {
    ctx.strokeStyle = selected ? sel_color : def_color
    ctx.fillStyle = ctx.strokeStyle

    ctx.save()
    ctx.lineWidth = thick
    ctx.beginPath()
    ctx.moveTo(b.x0, b.y0)
    ctx.bezierCurveTo(b.cx0, b.cy0, b.cx1, b.cy1, b.x1, b.y1)
    ctx.stroke()
    ctx.restore()

    ctx.beginPath()
    ctx.moveTo(p.p1[0], p.p1[1])
    ctx.lineTo(p.p2[0], p.p2[1])
    ctx.lineTo(p.p3[0], p.p3[1])
    ctx.lineTo(p.p4[0], p.p4[1])
    ctx.lineTo(p.p1[0], p.p1[1])
    ctx.closePath()
    ctx.fill()
    // c.stroke()
}

function renderPlace(ctx, zoom, place) {
    var wh = calcDimensions(zoom, place)
    var w = wh.w
    var h = wh.h
    var xy0 = absCoord(ctx, place.x*zoom, place.y*zoom)
    var x0 = xy0[0] - w / 2
    var y0 = xy0[1] - h / 2

    // color of place
    ctx.strokeStyle = place.selected ? sel_color : def_color
    ctx.fillStyle = ctx.strokeStyle

    // circle of w=2
    ctx.save()
    ctx.lineWidth = 2.0 * zoom
    ctx.ellipse(x0, y0, w, h)
    ctx.stroke()
    ctx.restore()
    ctx.beginPath()

    var d = ctx.lineWidth
    renderPlaceValue(ctx, zoom, x0 + 2*d, y0 + 2*d, w - 4*d, h - 4*d,
                     place.counter, place.timer)
    if(place.label) {
        var labelSize = 14.0 * zoom
        renderText(ctx, labelSize, x0, y0 + h + labelSize, w, place.label)
    }

    ctx.reset()
}

function renderTransition(ctx, zoom, transition) {
    if(!transition.in) transition.in = 0
    if(!transition.out) transition.out = 0
    var wh = calcDimensions(zoom, transition)
    var w = wh.w
    var h = wh.h
    if(transition.horizontal) {
        w = wh.h
        h = wh.w
    }

    var xy0 = absCoord(ctx, transition.x*zoom, transition.y*zoom)
    var x0 = xy0[0] - w / 2
    var y0 = xy0[1] - h / 2

    // color of transition
    ctx.strokeStyle = transition.selected ? sel_color : def_color
    ctx.fillStyle = ctx.strokeStyle

    // rect w x h
    ctx.rect(x0, y0, w, h)
    ctx.fill()
    ctx.beginPath()

    if(transition.label) {
        var labelSize = 14.0 * zoom
        if(transition.horizontal) {
            renderText(ctx, labelSize, x0 + w + labelSize, y0 , h, transition.label, true)
        } else {
            renderText(ctx, labelSize, x0, y0 + h + labelSize, w, transition.label)
        }
    }

    ctx.reset()
}

function renderPlaceValue(ctx, zoom, x0, y0, width, height, counter, timer) {
    var thick = 6.0 * zoom

    if(timer < 1 && counter > 0 && counter <= 9) {
        // defaults
        var rows = 1
        var vspacing = calcSpacing(height, thick, rows)
        var vmargin = calcCenteringMargin(height, thick, rows)

        switch(counter){
        case 3:
            rows = 2
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows) - (1.5*zoom) // ∆ hack
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 1)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 2)
            break
        case 4:
            rows = 2
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows)
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 2)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 2)
            break
        case 6:
            rows = 2
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows)
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 3)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 3)
            break
        case 5:
            rows = 3
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows)
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 2)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 1)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + 2*thick + 2*vspacing), width, 2)
            break
        case 7:
            rows = 3
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows)
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 3)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 1)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + 2*thick + 2*vspacing), width, 3)
            break
        case 8:
            rows = 3
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows)
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 3)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 2)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + 2*thick + 2*vspacing), width, 3)
            break
        case 9:
            rows = 3
            vspacing = calcSpacing(height, thick, rows)
            vmargin = calcCenteringMargin(height, thick, rows)
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, 3)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + thick + vspacing), width, 3)
            renderDotRow(ctx, thick, x0, y0 + (vmargin + 2*thick + 2*vspacing), width, 3)
            break
        default:
            renderDotRow(ctx, thick, x0, y0 + vmargin, width, counter)
        }
    } else if (timer > 1 && timer <= 4 && counter <= 3) {
        var rows = 1
        var vmargin = calcCenteringMargin(height, 4*thick, rows)

        if (counter > 1) {
            var hspacing = calcSpacing(width, thick / 2, timer+1)
            var hmargin = calcCenteringMargin(width, thick / 2, timer+1)
            var d = hmargin + hspacing

            renderDotColumn(ctx, thick, x0 + hmargin, y0 + thick, height - 2*thick, counter)
            renderBarRow(ctx, thick, x0 + d, y0 + vmargin, width - d, timer)
        } else {
            renderBarRow(ctx, thick, x0, y0 + vmargin, width, timer)
        }
    } else if (timer > 0 || counter > 0) {
        var y1 = y0 + calcCenteringMargin(height, 1, 1)
        ctx.save()
        ctx.lineWidth = 1.0 * zoom
        ctx.moveTo(x0 + 2.0 * zoom, y1)
        ctx.lineTo(x0 + width - 2.0 * zoom, y1)
        ctx.stroke()
        ctx.restore()
        ctx.beginPath()

        var labelSize = 14.0 * zoom
        if(counter > 0 && counter <= 4) {
            var voffset = calcCenteringMargin(height / 2, thick, 1)
            renderDotRow(ctx, thick, x0, y1 - voffset - thick/2, width, counter)
        } else if (counter > 0) {
            var voffset = calcCenteringMargin(height / 2, labelSize, 1)
            renderLabel(ctx, labelSize, x0, y1 - voffset - 2.0*zoom, width, "" + counter)
        }

        if(timer > 0 && timer <= 4) {
            var voffset = calcCenteringMargin(height / 2, thick*2, 1)
            renderBarRow(ctx, thick/2, x0, y1 + voffset, width, timer)
        } else if (timer > 0) {
            var voffset = calcCenteringMargin(height / 2, labelSize, 1)
            renderLabel(ctx, labelSize, x0, y1 + voffset + labelSize - 1.8*zoom, width, "" + timer)
        }
    }

    ctx.stroke() // not a bug! it's a style
    ctx.fill()
    ctx.reset()
}

function calcSpacing(space, thick, count) {
    return (space - (thick*count)) / (count+1)
}

function calcCenteringMargin(space, thick, count) {
    return (space - ((count-1)*calcSpacing(space, thick, count) + count*thick)) / 2
}

function calcDimensions(zoom, obj) {
    if(obj.transition) {
        var w = 6 * zoom
        var h = Math.max(obj.in * 15 * zoom, obj.out * 15 * zoom, 30 * zoom)
        return {"w":w, "h":h}
    } else if (obj.place) {
        var w = 50 * zoom
        var h = 50 * zoom
        return {"w":w, "h":h}
    }
    return {"w":-1, "h":-1}
}

function renderLabel(ctx, size, x0, y0, width, label) {
    ctx.font = "" + size + "px 'PT Serif'"
    ctx.textAlign = "center"
    ctx.strokeText(label, x0 + width / 2, y0)
    ctx.fillText(label, x0 + width / 2, y0)
}

function calcBorderPointPlace(zoom, w, x, y, cx, cy) {
    var angle = Math.atan2(cx - x, cy - y)

    var dX = (w + 3.0*zoom) * Math.sin(angle)
    var dY = (w + 3.0*zoom) * Math.cos(angle)

    return { "x": x + dX, "y": y + dY, "angle": angle }
}

function calcBorderPointTransition(zoom, data, cx, cy, horizontal, thick, inbound, count, index) {
    var offset = 3 * zoom
    var offsetH = 8 * zoom

    // line 2 - variant vertical inbound
    var line21 = [
                data.x - offset, // x3 line2[0]
                data.y,          // y3 line2[1]
                data.x - offset, // x4 line2[2]
                data.y + data.h  // y4 line2[3]
            ]

    // line 2 - variant vertical outbound
    var line22 = [
                data.x + data.w + offset, // x3 line2[0]
                data.y,                   // y3 line2[1]
                data.x + data.w + offset, // x4 line2[2]
                data.y + data.h,          // y4 line2[3]
            ]

    // line 2 - variant horizontal inbound
    var line23 = [
                data.x - data.h / 2 + data.w / 2,          // x3 line2[0]
                data.y + data.h / 2 - data.w - offset,     // y3 line2[1]
                data.x - data.h / 2 + data.w / 2 + data.h, // x4 line2[2]
                data.y + data.h / 2 - offset               // y4 line2[3]
            ]

    // line 2 - variant horizontal outbound
    var line24 = [
                data.x - data.h / 2 + data.w / 2,          // x3 line2[0]
                data.y + data.h / 2 + offsetH,             // y3 line2[1]
                data.x - data.h / 2 + data.w / 2 + data.h, // x4 line2[2]
                data.y + data.h / 2 + data.w + offsetH     // y4 line2[3]
            ]

    // line 1
    var line11 = [data.cx, data.cy, Math.min(cx, line21[0]), cy]
    var line12 = [data.cx, data.cy, Math.min(cx, line22[0]), cy]
    var line13 = [data.cx, data.cy, cx, Math.min(cy, line23[1])]
    var line14 = [data.cx, data.cy, cx, Math.min(cy, line24[1])]

    var line1 = horizontal ? (inbound ? line13 : line14) : (inbound ? line11 : line12)
    var line2 = horizontal ? (inbound ? line23 : line24) : (inbound ? line21 : line22)

    var dX = 0
    var dY = 0
    var spacing = calcSpacing(data.h, thick, count)
    var margin = calcCenteringMargin(data.h, thick, count)

    if(horizontal) {
        dX = margin + index*(thick + spacing)
    } else {
        dY = margin + index*(thick + spacing)
    }

    var angle = Math.atan2(cx - data.cx, cy - data.cy)
    var finalX
    var finalY

    if(horizontal) {
        finalX = line2[0] + dX
        finalY = line2[1]
    } else {
        finalY = line2[1] + dY
        finalX = line2[0]
    }

    return { "x":finalX, "y":finalY, "angle":angle }
}

function renderText(ctx, size, x0, y0, space, text, valign) {
    text = "// "+text
    var chunks = text.match(/.{1,16}/g);
    var offset = 0
    for(var i in chunks) {
        var subchunks = chunks[i].split('\n')
        for(var j in subchunks) {
            var str = subchunks[j]
            ctx.font = "oblique " + size + "px 'PT Serif'"
            ctx.fillStyle = "#7f8c8d"
            if(valign) {
                ctx.fillText(str, x0, y0 + space / 2 + offset)
            } else {
                ctx.textAlign = "center"
                ctx.fillText(str, x0 + space / 2, y0 + offset + 4.0)
            }
            offset += size + 2.0
        }
    }
}

function renderDotRow(ctx, thick, x0, y0, width, value) {
    var hspacing = calcSpacing(width, thick, value)
    var hmargin = calcCenteringMargin(width, thick, value)
    for(var i = 0; i < value; ++i) {
        ctx.ellipse(x0 + (hmargin + i*(hspacing + thick)), y0, thick, thick)
    }
}

function renderDotColumn(ctx, thick, x0, y0, height, value) {
    var vspacing = calcSpacing(height, thick, value)
    var vmargin = calcCenteringMargin(height, thick, value)
    for(var i = 0; i < value; ++i) {
        ctx.ellipse(x0, y0 + (vmargin + i*(vspacing + thick)), thick, thick)
    }
}

function renderBarRow(ctx, thick, x0, y0, width, value) {
    var w = thick / 2
    var h = 4 * thick
    var hspacing = calcSpacing(width, w, value)
    for(var i = 1, j = 0; j < value; ++i, ++j) {
        ctx.rect(x0 + (i*hspacing + j*w), y0, w, h)
    }
}

function absCoord(ctx, x, y){
    return [ctx.canvas.canvasSize.width / 2 + ctx.canvas.canvasWindow.width / 2 + x,
            ctx.canvas.canvasSize.height / 2 + ctx.canvas.canvasWindow.height / 2 + y]
}
