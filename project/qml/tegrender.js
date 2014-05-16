function renderBuf(ctx, region, zoom, cache, kind) {
    if(!cache[kind] || cache[kind].length < 1) {
        return
    }
    /*
    var mz = 0
    var maxX = region.width + mz
    var maxY = region.height + mz
    var dx, dy
    */
    for(var i = 0; i < cache[kind].length; i++) {
        var it = cache[kind][i]
        ctx.lineWidth = it.lineWidth
        ctx.strokeStyle = it.strokeStyle
        ctx.fillStyle = it.fillStyle

        if(kind === "rrect") {
            /*
            dx = it.x - region.x
            dy = it.y - region.y
            if (dx > -mz && dy > -mz && dx < maxX - it.w && dy < maxY - it.h) {
            */
            ctx.roundedRect(it.x, it.y, it.w, it.h, it.r, it.r)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "rect") {
            /*
            dx = it.x - region.x
            dy = it.y - region.y
            if (dx > -mz && dy > -mz && dx < maxX - it.w && dy < maxY - it.h) {
            */
            ctx.rect(it.x, it.y, it.w, it.h)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "circle") {
            /*
            dx = it.x - region.x
            dy = it.y - region.y
            if (dx > -mz && dy > -mz && dx < maxX - it.d && dy < maxY - it.d) {
            */
            ctx.ellipse(it.x, it.y, it.d, it.d)
            draw(ctx, it.stroke, it.fill)
        } else if(kind === "bezier") {
            /*
            var dSx = it.start.x - region.x
            var dSy = it.start.y - region.y
            var dEx = it.end.x - region.x
            var dEy = it.end.x - region.y
            var dC1x = it.c1.x - region.x
            var dC1y = it.c1.y - region.y
            var dC2x = it.c2.x - region.x
            var dC2y = it.c2.y - region.y
            if ( (dSx > -mz && dSy > -mz && dSx < maxX && dSy < maxY) ||
                    (dEx > -mz && dEy > -mz && dEx < maxX && dEy < maxY) ||
                    (dC1x > -mz && dC1y > -mz && dC1x < maxX && dC1y < maxY) ||
                    (dC2x > -mz && dC2y > -mz && dC2x < maxX && dC2y < maxY) ) {
                    */
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

function render(ctx, region, zoom, cache) {
    var rx = region.x
    var ry = region.y
    var rw = region.width
    var rh = region.height
    ctx.clearRect(rx, ry, rw, rh)

    renderBuf(ctx, region, zoom, cache, "rrect")
    renderBuf(ctx, region, zoom, cache, "circle")
    renderBuf(ctx, region, zoom, cache, "rect")
    renderBuf(ctx, region, zoom, cache, "line")
    renderBuf(ctx, region, zoom, cache, "bezier")
    renderBuf(ctx, region, zoom, cache, "poly")
    renderBuf(ctx, region, zoom, cache, "text")
    renderBuf(ctx, region, zoom, cache, "chain")
}
