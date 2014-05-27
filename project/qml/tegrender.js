function limit(x) {
    if(x > 5.0) {
        x = 5.0
    } else if (x < 0.3) {
        x = 0.3
    }
    return x
}

function renderBuf(ctx, region, cache, kind) {
    if(!cache[kind] || cache[kind].length < 1) {
        return
    }

    for(var i = 0; i < cache[kind].length; i++) {
        var it = cache[kind][i]
        ctx.lineWidth = it.lineWidth
        ctx.strokeStyle = it.strokeStyle
        ctx.fillStyle = it.fillStyle
        ctx.translate(0, 0.5)

        if(kind === "rrect") {
            ctx.roundedRect(it.x, it.y, it.w, it.h, it.r, it.r)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "rect") {
            ctx.rect(it.x, it.y, it.w, it.h)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "circle") {
            ctx.ellipse(it.x, it.y, it.d, it.d)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "bezier") {
            ctx.beginPath()
            ctx.moveTo(it.start.x, it.start.y)
            ctx.bezierCurveTo(it.c1.x, it.c1.y, it.c2.x, it.c2.y, it.end.x, it.end.y)
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

    renderBuf(ctx, region, cache, "rrect")
    renderBuf(ctx, region, cache, "circle")
    renderBuf(ctx, region, cache, "rect")
    renderBuf(ctx, region, cache, "line")
    renderBuf(ctx, region, cache, "bezier")
    renderBuf(ctx, region, cache, "poly")
    renderBuf(ctx, region, cache, "text")
    renderBuf(ctx, region, cache, "chain")
}
