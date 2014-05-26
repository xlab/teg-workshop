import QtQuick 2.0
import 'planerender.js' as R

Canvas {
    id: layer
    tileSize: "1024x1024"

    property alias screen: renderer.screen
    property int layerId

    onPaint: {
        var ctx = getContext("2d")
        if(!renderer.cache) {
            console.error("error: cache broken, id:", layerId)
            return
        }
        R.render(ctx, region, renderer.cache)
    }

    Item {
        id: renderer
        property var screen
        property var cache

        onScreenChanged: {
            var cache = prepareCache(screen)
            if(!cache) return
            renderer.cache = cache
            layer.requestPaint()
        }

        // see bug https://groups.google.com/d/msg/go-qml/h5gDOjyE8Yc/-oWP6GLaXzIJ
        function prepareCache(screen) {
            var cache = {
                "circle": [], "rect": [], "line": [], "pad": [],
                "chain": [], "poly": [], "text": [],
            }
            var i, j, buf, it, pos, style, points

            buf = screen.circles
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.circle[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "x": it.x, "y": it.y, "d": it.d
                }
            }

            buf = screen.pads
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.pad[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "x": it.x, "y": it.y, "d": it.d
                }
            }

            buf = screen.rects
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.rect[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "x": it.x, "y": it.y, "w": it.w, "h": it.h
                }
            }

            buf = screen.lines
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.line[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "start": it.start, "end": it.end
                }
            }

            buf = screen.texts
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.text[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "align": it.align, "vertical": it.vertical, "fontSize": it.fontSize,
                    "oblique": it.oblique, "font": it.font,
                    "x": it.x, "y": it.y, "label": it.label
                }
            }

            buf = screen.polys
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                points = []
                for(j = 0; j < it.length; ++j) {
                    points[j] = it.at(j)
                }
                cache.poly[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "points": points
                }
            }

            buf = screen.chains
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                points = []
                for(j = 0; j < it.length; ++j) {
                    points[j] = it.at(j)
                }
                cache.chain[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "points": points
                }
            }

            return cache
        }
    }
}
