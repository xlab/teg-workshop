function relativeCoords(ctx, x, y) {
    return {
        "x":-ctx.canvas.canvasSize.width / 2 - ctx.canvas.canvasWindow.width / 2 + x,
        "y":-ctx.canvas.canvasSize.height / 2 - ctx.canvas.canvasWindow.height / 2 + y,
    }
}

function absCoords(ctx, x, y){
    return {
        "x":ctx.canvas.canvasSize.width / 2 + ctx.canvas.canvasWindow.width / 2 + x,
        "y":ctx.canvas.canvasSize.height / 2 + ctx.canvas.canvasWindow.height / 2 + y,
    }
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

        if(kind === "rect") {
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
            var fontspec = "" + it.fontSize + "px '" + it.font + "'"
            if(it.oblique) {
                fontspec = "oblique " + fontspec
            }
            ctx.font = fontspec
            ctx.textAlign = it.align
            if(it.fill) {
                ctx.fillText(it.label, it.x, it.y)
            }
            if(it.stroke) {
                ctx.strokeText(it.label, it.x, it.y)
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

    renderBuf(ctx, region, cache, "circle")
    renderBuf(ctx, region, cache, "rect")
    renderBuf(ctx, region, cache, "line")
    renderBuf(ctx, region, cache, "bezier")
    renderBuf(ctx, region, cache, "poly")
    renderBuf(ctx, region, cache, "text")
    renderBuf(ctx, region, cache, "chain")
}
