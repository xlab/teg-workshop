var sel_color = "#b10000"
var def_color = "#000000"

var iii = 0
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
        renderArc(ctx, zoom, model.arcs[idx])
    }

    if(model.magicStrokeAvailable) {
        renderMagicStroke(ctx, zoom, model.magicStroke)
    }
}

function renderMagicStroke(ctx, zoom, stroke) {
    var xy0 = absCoord(ctx, stroke.x0*zoom, stroke.y0*zoom)
    var xy1 = absCoord(ctx, stroke.x1*zoom, stroke.y1*zoom)
    var thick = 1.5 * zoom

    ctx.beginPath()
    ctx.lineWidth =  thick
    ctx.strokeStyle = "#2980b9"
    ctx.moveTo(xy0.x, xy0.y)
    ctx.lineTo(xy1.x, xy1.y)
    ctx.stroke()
    ctx.reset()
}

function renderArc(ctx, zoom, arc) {
    var transition = arc.transition
    var place = arc.place
    var index = arc.index
    var inbound = arc.inbound
    var control

    if(inbound) {
        control = place.out_control
    } else {
        control = place.in_control
    }
    
    var xy_p = absCoord(ctx, place.x*zoom, place.y*zoom)
    var xy_t = absCoord(ctx, transition.x*zoom, transition.y*zoom)
    var xy_c = absCoord(ctx, control.x*zoom, control.y*zoom)
    var wh_p = calcDimensions(zoom, place)
    var wh_t = calcDimensions(zoom, transition)
    var wh_c = {"w":10*zoom, "h":10*zoom}
    
    var thick = 2.0 * zoom
    
    // collect variables to draw transition-end of bezier
    var data_t = {
        "x": xy_t.x, "y": xy_t.y,
        "w": wh_t.w, "h": wh_t.h,
        "cx": xy_t.x + wh_t.w/2.0,
        "cy": xy_t.y + wh_t.h/2.0,
    }
    
    var count = inbound ? transition.in : transition.out
    var horizontal = transition.horizontal
    // calc the point near transition
    var xyA_t = calcBorderPointTransition(zoom, data_t, xy_c.x, xy_c.y, horizontal, thick, inbound, count, index)
    // calc the point near place
    var xyA_p = calcBorderPointPlace(zoom, wh_p.w/2.0, xy_p.x+wh_p.w/2.0, xy_p.y+wh_p.h/2.0, xy_c.x, xy_c.y)
    
    // get the points to draw
    var endXYA = inbound ? xyA_t : xyA_p
    
    // nice ∆ pointer
    var p1X = endXYA.x_tip
    var p1Y = endXYA.y_tip
    var p2X = p1X - 3*thick
    var p2Y = p1Y + 5*thick
    var p3X = p1X
    var p3Y = p1Y + (5 * 2/3 *thick)
    var p4X = p1X + 3*thick
    var p4Y = p1Y + 5*thick
    
    // rotated ∆
    var p1 = calcRotatePoint(p1X, p1Y, p1X, p1Y, -endXYA.angle)
    var p2 = calcRotatePoint(p2X, p2Y, p1X, p1Y, -endXYA.angle)
    var p3 = calcRotatePoint(p3X, p3Y, p1X, p1Y, -endXYA.angle)
    var p4 = calcRotatePoint(p4X, p4Y, p1X, p1Y, -endXYA.angle)
    
    var cp_t
    if(inbound) {
        if(horizontal) {
            cp_t = {"x": xyA_t.x, "y": xyA_t.y - 50.0*zoom}
        } else {
            cp_t = {"x": xyA_t.x - 50.0*zoom, "y": xyA_t.y}
        }
    } else {
        if(horizontal) {
            cp_t = {"x": xyA_t.x, "y": xyA_t.y + 50.0*zoom}
        } else {
            cp_t = {"x": xyA_t.x + 50.0*zoom, "y": xyA_t.y}
        }
    }
    
    drawPointedBezierCurve(ctx, zoom, thick, place.selected && transition.selected,
                           {"x0":xyA_p.x, "y0":xyA_p.y, "x1":xyA_t.x, "y1":xyA_t.y,
                               "cx0":xy_c.x + wh_c.w/2, "cy0":xy_c.y + wh_c.h/2,
                               "cx1":cp_t.x, "cy1":cp_t.y},
                           {"p1":p1, "p2":p2, "p3":p3, "p4":p4})

    if(model.altPressed) {
        ctx.beginPath()
        ctx.strokeStyle = "#3498db"
        ctx.lineWidth = 1.0 * zoom
        if(horizontal) {
            ctx.moveTo(xy_p.x + wh_p.h/2, xy_p.y + wh_p.w/2)
            ctx.lineTo(xy_c.x + wh_c.h/2, xy_c.y + wh_c.w/2)
            ctx.lineTo(xy_t.x + wh_t.h/2, xy_t.y + wh_t.w/2)
        } else {
            ctx.moveTo(xy_p.x + wh_p.w/2, xy_p.y + wh_p.h/2)
            ctx.lineTo(xy_c.x + wh_c.w/2, xy_c.y + wh_c.h/2)
            ctx.lineTo(xy_t.x + wh_t.w/2, xy_t.y + wh_t.h/2)
        }

        ctx.stroke()
    }
    
    ctx.beginPath()
    if(place.selected) {
        ctx.fillStyle = "#f1c40f"
        ctx.rect(xy_c.x, xy_c.y, wh_c.w, wh_c.h)
        ctx.fill()
    }
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
    var x0 = xy0.x
    var y0 = xy0.y
    
    // color of place
    ctx.strokeStyle = place.selected ? sel_color : def_color
    ctx.fillStyle = ctx.strokeStyle
    
    // circle of w=2
    ctx.save()
    //ctx.fillStyle = "#ecf0f1"
    ctx.lineWidth = 2.0 * zoom
    ctx.ellipse(x0, y0, w, h)
    //ctx.fill()
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
    var x0 = xy0.x
    var y0 = xy0.y
    
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
            // fix horizontal text alignment
            renderText(ctx, labelSize, x0 + w + labelSize, y0 + h/2, h, transition.label, true)
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
    } else if (timer > 0 && timer <= 4 && counter <= 3) {
        var rows = 1
        var vmargin = calcCenteringMargin(height, 4*thick, rows)
        
        if (counter > 0) {
            var hspacing = calcSpacing(width, thick, timer + 1)
            var hmargin = calcCenteringMargin(width, thick, timer + 1)
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
    
    ctx.stroke() // not a bug
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
        var w = 50.0 * zoom
        var h = 50.0 * zoom
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

function calcBorderPointPlace(zoom, radius, x_center, y_center, x_control, y_control) {
    var angle = Math.atan2(x_control - x_center, y_control - y_center)
    
    var dX_tip = (radius + 3.0*zoom) * Math.sin(angle)
    var dY_tip = (radius + 3.0*zoom) * Math.cos(angle)
    var dX = (radius + 5.0*zoom) * Math.sin(angle)
    var dY = (radius + 5.0*zoom) * Math.cos(angle)
    
    return { "x": x_center + dX, "y": y_center + dY, "angle": angle,
        "x_tip": x_center + dX_tip, "y_tip": y_center + dY_tip}
}

function calcBorderPointTransition(zoom, data, cx, cy, horizontal, thick, inbound, count, index) {
    var offset = 3 * zoom
    var offset_tip = 2 * zoom
    
    var point
    if(horizontal && inbound) {
        point = {"x": data.x, "y": data.y - offset}
    } else if(inbound) {
        point = {"x": data.x - offset, "y": data.y}
    } else if(horizontal) {
        point = {"x": data.x, "y": data.y + data.w + offset}
    } else {
        point = {"x": data.x + data.w + offset, "y": data.y}
    }
    
    var dX = 0
    var dY = 0
    var spacing = calcSpacing(data.h, thick, count)
    var margin = calcCenteringMargin(data.h, thick, count)
    
    if(horizontal) {
        dX = margin + index*(thick + spacing)
    } else {
        dY = margin + index*(thick + spacing)
    }
    
    var angle
    var finalX, finalX_tip
    var finalY, finalY_tip
    
    if(horizontal) {
        finalX = point.x + dX
        finalY = point.y
        
        if (inbound) {
            angle = Math.PI
            finalX_tip = point.x + dX
            finalY_tip = point.y + offset_tip
        } else {
            angle = 0
            finalX_tip = point.x + dX
            finalY_tip = point.y - offset_tip
        }
    } else {
        finalX = point.x
        finalY = point.y + dY
        if(inbound) {
            angle = -Math.PI/2
            finalX_tip = point.x + offset_tip
            finalY_tip = point.y + dY
        } else {
            angle = Math.PI/2
            finalX_tip = point.x - offset_tip
            finalY_tip = point.y + dY
        }
    }
    
    return { "x":finalX, "y":finalY, "angle":angle, "x_tip": finalX_tip, "y_tip": finalY_tip }
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
    return {
        "x":ctx.canvas.canvasSize.width / 2 + ctx.canvas.canvasWindow.width / 2 + x,
        "y":ctx.canvas.canvasSize.height / 2 + ctx.canvas.canvasWindow.height / 2 + y,
    }
}
