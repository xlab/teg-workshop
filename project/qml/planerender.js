var gridGap = 16.0
var notchHeight = 3.0

function renderBuf(ctx, region, cache, kind) {
    if(!cache[kind] || cache[kind].length < 1) {
        return
    }

    for(var i = 0; i < cache[kind].length; i++) {
        var it = cache[kind][i]
        ctx.lineWidth = it.lineWidth
        ctx.strokeStyle = it.strokeStyle
        ctx.fillStyle = it.fillStyle

        if(kind === "rect") {
            ctx.rect(it.x, it.y, it.w, it.h)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "circle" || kind === "pad") {
            ctx.ellipse(it.x, it.y, it.d, it.d)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "poly") {
            if(it.points.length > 2) {
                ctx.beginPath()
                ctx.moveTo(it.points[0].x, it.points[0].y)
                for(var j = 1; j < it.points.length; ++j) {
                    ctx.lineTo(it.points[j].x, it.points[j].y)
                }
                ctx.closePath()
                draw(ctx, it.stroke, it.fill)
            }
        } else if(kind === "chain") {
            if(it.points.length > 2) {
                ctx.beginPath()
                ctx.moveTo(it.points[0].x, it.points[0].y)
                for(var j = 1; j < it.points.length; ++j) {
                    ctx.lineTo(it.points[j].x, it.points[j].y)
                }
                draw(ctx, it.stroke, it.fill)
            }
        } else if(kind === "line") {
            ctx.beginPath()
            ctx.moveTo(it.start.x, it.start.y)
            ctx.lineTo(it.end.x, it.end.y)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "text") {
            var x, y
            if(it.vertical) {
                x = 0
                y = 0
                ctx.save()
                ctx.translate(it.x, it.y)
                ctx.rotate(Math.PI/2)
            } else {
                x = it.x
                y = it.y
            }
            var fontspec = "" + it.fontSize + "px '" + it.font + "'"
            if(it.oblique) {
                fontspec = "oblique " + fontspec
            }
            ctx.font = fontspec
            ctx.textAlign = it.align
            if(it.fill) {
                ctx.fillText(it.label, x, y)
            }
            if(it.stroke) {
                ctx.strokeText(it.label, x, y)
            }
            if(it.vertical) {
                ctx.restore()
            }
        }
        ctx.reset()
    }
}

function draw(ctx, stroke, fill) {
    if(fill) {
        ctx.fill()
    }
    if(stroke) {
        ctx.stroke()
    }
}

function render(ctx, region, cache) {
    var rx = region.x
    var ry = region.y
    var rw = region.width
    var rh = region.height
    ctx.clearRect(rx, ry, rw, rh)

    renderBuf(ctx, region, cache, "rect")
    renderBuf(ctx, region, cache, "chain")
    renderBuf(ctx, region, cache, "pad")
    renderBuf(ctx, region, cache, "circle")
    renderBuf(ctx, region, cache, "line")
    renderBuf(ctx, region, cache, "poly")
    renderBuf(ctx, region, cache, "text")
}

function background(ctx, region, zoom, cw, ch, ww, wh, wx, wy) {
    var rx = region.x
    var ry = region.y
    var rw = region.width
    var rh = region.height
    ctx.clearRect(rx, ry, rw, rh)

    var gap = gridGap * zoom

    var cx = cw+ww
    var cy = ch+wh

    var gx = wx + ww*2 - 20
    var gy = cy - 20
    var dx = cx + 20
    var dy = wy + 20

    var kX = Math.floor((cx - rx) / gap)
    var kY = Math.floor((cy - ry) / gap)
    var sx = cx - kX * gap
    var sy = cy - kY * gap

    // grid
    if(zoom > 0.6) {
        ctx.strokeStyle = "#bdc3c7"
        ctx.lineWidth = 1
        ctx.translate(0.5, 0)

        for(var x = sx; x < sx+rw; x += gap) {
            ctx.moveTo(x, sy)
            ctx.lineTo(x, sy+rh)
        }
        for(var y = sy; y < sy+rh; y += gap) {
            ctx.moveTo(sx, y)
            ctx.lineTo(sx+rw, y)
        }

        ctx.stroke()
        ctx.reset()
        ctx.beginPath()
    }
    // end of grid

    // bars
    ctx.strokeStyle = "#000000"
    ctx.lineWidth = 1
    ctx.translate(0.5, 0)

    ctx.moveTo(cx, sy)
    ctx.lineTo(cx, sy+rh)
    ctx.moveTo(sx, cy)
    ctx.lineTo(sx+rw, cy)

    ctx.stroke()
    ctx.reset()
    ctx.beginPath()
    // end of bars

    // notches
    if(zoom > 0.6) {
        ctx.strokeStyle = "#000000"
        ctx.lineWidth = 1
        ctx.translate(0.5, 0)
        ctx.font = "14px Times New Roman"
        ctx.textAlign = "right"

        var offX = -kX
        for(var x = sx; x < sx+rw; x += gap) {
            ctx.moveTo(x, cy+notchHeight)
            ctx.lineTo(x, cy-notchHeight)
            if(offX !== 0) {
                if(zoom > 1 || offX%2 == 0) {
                    ctx.fillText(offX, x+4, cy+notchHeight+12)
                }
            }
            offX++
        }

        var offY = kY
        for(var y = sy; y < sy+rh; y += gap) {
            ctx.moveTo(cx+notchHeight, y)
            ctx.lineTo(cx-notchHeight, y)
            if(offY !== 0) {
                if(zoom > 1 || offY%2 == 0) {
                    ctx.fillText(offY, cx-notchHeight-3, y+4)
                }
            }
            offY--
        }

        ctx.stroke()
        ctx.reset()
        ctx.beginPath()
    }
    // end of notches

    // labels
    ctx.fillStyle = "#000000"
    ctx.translate(0.5, 0)
    ctx.font = "20px Times New Roman"

    ctx.fillText("ɣ", gx, gy)
    ctx.fillText("δ", dx, dy)

    ctx.reset()
    ctx.beginPath()
    // end of lables
}
