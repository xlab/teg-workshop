import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import TegView 1.0
import 'tegrender.js' as R

ApplicationWindow {
    id: view
    width: 700
    height: 600
    color: "#ecf0f1"
    property alias ctrl: ctrl
    property alias editMode: editbtn.checked
    property alias viewMode: viewbtn.checked

    toolBar: ToolBar {
        RowLayout {
            ToolButton {
                text: "+"
                onClicked: cv.zoom += 0.2
            }
            ToolButton {
                text: "-"
                onClicked: cv.zoom -= 0.2
            }
            Rectangle {
                width: 3
                anchors.top: parent.top
                anchors.bottom: parent.bottom
                anchors.margins: 2
                color: "#34495e"
            }
            ExclusiveGroup { id: mode }
            RadioButton {
                id: viewbtn
                exclusiveGroup: mode
                text: "View"
            }
            RadioButton {
                id: editbtn
                checked: true
                exclusiveGroup: mode
                text: "Edit"
            }
        }
    }

    Ctrl {
        id: ctrl
        canvasWidth: cv.canvasSize.width
        canvasHeight: cv.canvasSize.height
        canvasWindowX: cv.canvasWindow.x
        canvasWindowY: cv.canvasWindow.y
        canvasWindowHeight: cv.canvasWindow.height
        canvasWindowWidth: cv.canvasWindow.width
        zoom: cv.zoom
    }

    RowLayout {
        id: keyHintRow
        z: 2
        property string fontColor: "#34495e"
        anchors.left: parent.left
        anchors.bottom: parent.bottom
        anchors.leftMargin: 10
        anchors.bottomMargin: 10
        spacing: 10
        Text {
            id: ctrlIndicator
            text: "Ctrl"
            font.capitalization: Font.SmallCaps
            font.pointSize: 16
            color: keyHintRow.fontColor
            visible: ctrl.modifierKeyControl
        }

        Text {
            id: altIndicator
            text: (ctrlIndicator.visible ? "+ ": "") + "Alt"
            font.capitalization: Font.SmallCaps
            font.pointSize: 16
            color: keyHintRow.fontColor
            visible: ctrl.modifierKeyAlt
        }

        Text {
            id: shiftIndicator
            text: ((ctrlIndicator.visible || altIndicator.visible)  ? "+ ": "") + "Shift"
            font.pointSize: 16
            font.capitalization: Font.SmallCaps
            color: keyHintRow.fontColor
            visible: ctrl.modifierKeyShift
        }

        Text {
            id: keyHint
            font.pointSize: 16
            font.capitalization: Font.SmallCaps
            color: keyHintRow.fontColor
            visible: false

            function setText(text) {
                if(text.length > 0) {
                    this.visible = true
                    this.text = ((ctrlIndicator.visible || altIndicator.visible || shiftIndicator.visible)  ? "+ ": "")
                    this.text += text
                } else {
                    this.visible = false
                    this.text = ""
                }
            }
        }
    }

    Text {
        id: modeHint
        font.pointSize: 20
        font.capitalization: Font.SmallCaps
        color: "#b4b4b4"
        text: editMode ? "Edit mode" : "View mode"
        anchors.top: parent.top
        anchors.right: parent.right
        anchors.rightMargin: 20
        anchors.topMargin: 15
    }

    /*
    Image {
        anchors.fill: parent
        fillMode: Image.Tile
        source: "grid.png"
    }
    */

    Canvas {
        id: cv
        property real zoom: 1.0
        anchors.fill: parent
        canvasSize.width: 16536
        canvasSize.height: 16536
        canvasWindow.width: width
        canvasWindow.height: height
        tileSize: "1024x1024"

        onPaint: {
            var ctx = cv.getContext("2d")
            if(!renderer.cache) {
                console.error("error: cache broken")
                return
            }
            R.render(ctx, region, cv.zoom, renderer.cache)
        }

        onCanvasWindowChanged: {
            if(canvasWindow.width !== ctrl.canvasWindowWidth ||
                    canvasWindow.height !== ctrl.canvasWindowHeight) {
                ctrl.canvasWindowWidth = canvasWindow.width
                ctrl.canvasWindowHeight = canvasWindow.height
                ctrl.flush()
            }
        }

        onZoomChanged: {
            ctrl.zoom = cv.zoom
            ctrl.flush()
        }

        Component.onCompleted: {
            canvasWindow.x = canvasSize.width / 2
            canvasWindow.y = canvasSize.height / 2
            coldstart.start()
        }
    }

    MouseArea {
        id: io
        anchors.fill: parent
        acceptedButtons: Qt.LeftButton | Qt.RightButton
        property real dragOffset: 50.0
        property int cx0
        property int cy0
        property int x0
        property int y0
        property bool rightPressed

        focus: true
        Keys.onPressed: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            ctrl.modifierKeyAlt = (event.modifiers & Qt.AltModifier) ? true : false

            if(ctrl.modifierKeyControl && event.key === Qt.Key_V) {
                viewbtn.checked = true
            } else if (ctrl.modifierKeyControl && event.key === Qt.Key_E) {
                editbtn.checked = true
            } else {
                //console.log(event.key, event.text)
                ctrl.keyPressed(event.key, event.text)
            }

            if(ctrl.modifierKeyControl) {
                keyHint.setText(event.text)
            }
            event.accepted = true
        }
        Keys.onReleased: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            ctrl.modifierKeyAlt = (event.modifiers & Qt.AltModifier) ? true : false

            keyHint.setText("")
            event.accepted = true
            ctrl.flush()
        }

        onPressed: {
            rightPressed = (mouse.button === Qt.RightButton)
            if(viewMode || rightPressed) {
                cx0 = cv.canvasWindow.x
                cy0 = cv.canvasWindow.y
                x0 = mouse.x
                y0 = mouse.y
            } else if(editMode) {
                ctrl.mousePressed(mouse.x, mouse.y)
            }
        }

        onPositionChanged: {
            if(x0 != mouse.x || y0 != mouse.y) {
                if (viewMode || rightPressed) {
                    cv.canvasWindow.x = cx0 + (x0 - mouse.x)
                    cv.canvasWindow.y = cy0 + (y0 - mouse.y)
                    cv.requestPaint()
                } else if(editMode) {
                    if(mouse.x < 0 + dragOffset) {
                        cv.canvasWindow.x += -5.0
                    } else if(mouse.x > cv.canvasWindow.width - dragOffset) {
                        cv.canvasWindow.x += 5.0
                    }
                    if(mouse.y < 0 + dragOffset) {
                        cv.canvasWindow.y += -5.0
                    } else if(mouse.y > cv.canvasWindow.height - dragOffset) {
                        cv.canvasWindow.y += 5.0
                    }
                    ctrl.mouseMoved(mouse.x, mouse.y)
                }
            }
        }

        onReleased: {
            if (editMode && !rightPressed) {
                ctrl.mouseReleased(mouse.x, mouse.y)
            }
            rightPressed = !(mouse.button === Qt.RightButton)
        }

        onDoubleClicked: {
            if (editMode && !rightPressed) {
                ctrl.mouseDoubleClicked(mouse.x, mouse.y)
            }
        }
    }

    Timer {
        id: coldstart
        interval: 1000
        onTriggered: ctrl.flush()
        repeat: false
    }

    Item {
        id: renderer
        property var screen: tegRenderer.screen
        property var cache

        onScreenChanged: {
            var cache = prepareCache(screen)
            if(!cache) return
            renderer.cache = cache
            cv.requestPaint()
        }

        // see bug https://groups.google.com/d/msg/go-qml/h5gDOjyE8Yc/-oWP6GLaXzIJ
        function prepareCache(screen) {
            var cache = {
                "circle": [], "rect": [], "line": [], "rrect": [],
                "bezier": [], "poly": [], "text": [], "chain": []
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

            buf = screen.rRects
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.rrect[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "x": it.x, "y": it.y, "w": it.w, "h": it.h, "r": it.r
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

            buf = screen.bezier
            if(!buf) return
            for(i = 0; i < buf.length; ++i) {
                it = buf.at(i)
                if(!it) return
                style = it.style
                if(!style) return
                cache.bezier[i] = {
                    "lineWidth": style.lineWidth,
                    "fill": style.fill, "stroke": style.stroke,
                    "strokeStyle": style.strokeStyle, "fillStyle": style.fillStyle,
                    "start": it.start, "end": it.end, "c1": it.c1, "c2": it.c2
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

