import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Controls.Styles 1.1
import QtQuick.Layouts 1.1
import QtQuick.Dialogs 1.1
import TegCtrl 1.0
import 'tegrender.js' as R

ApplicationWindow {
    id: view
    width: 800
    height: 600
    color: "#ecf0f1"

    property bool sane: true
    property string errorText
    property string label: ctrl.title
    property string keyhint

    property alias ctrl: ctrl
    property alias lock: tglLock.enabled

    property string panelBtnFgColor: "black"
    property string panelBtnBgColor: "#15000000"
    property string panelBtnFgPressedColor: "white"
    property string panelBtnBgPressedColor: "#3498db"

    onActiveChanged: {
        if(!active) {
            ctrl.modifierKeyShift = false
            ctrl.modifierKeyControl = false
            ctrl.modifierKeyAlt = false
            view.keyhint = ""
        } else {
            ctrl.flush()
        }
    }

    title: "TEG Workshop / v1.0 beta"
    toolBar: ToolBar {
        style: ToolBarStyle {
            padding {
                left: 8; right: 8 ; top: 3; bottom: 3
            }

            background: Rectangle {
                implicitWidth: 100
                implicitHeight: 60
                Rectangle {
                    anchors.left: parent.left
                    anchors.bottom: parent.bottom
                    anchors.right: parent.right
                    height: 1
                    color: "#999"
                }
            }
        }
        RowLayout {
            anchors.fill: parent

            XButton {
                imageSrc: "icons/application-blue.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: ctrl.newWindow()
            }

            XButton {
                imageSrc: "icons/folder-horizontal-open.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: openFile.open()
            }

            XButton {
                imageSrc: "icons/disk.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: saveFile.open()
            }

            XSeparator{}

            XButton {
                imageSrc: "icons/magnifier-zoom-in.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: {
                    pinchArea.zoom = pinchArea.limit(pinchArea.zoom + 0.2)
                }
            }

            XButton {
                imageSrc: "icons/magnifier-zoom-out.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: {
                    pinchArea.zoom = pinchArea.limit(pinchArea.zoom - 0.2)
                }
            }

            XButton {
                imageSrc: "icons/magnifier-zoom-fit.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: {
                    cv.canvasWindow.x = cv.canvasSize.width / 2
                    cv.canvasWindow.y = cv.canvasSize.height / 2
                    pinchArea.zoom = 1.0
                }
            }

            XSeparator{}

            XButton {
                imageSrc: "icons/table-sheet.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: ctrl.planeView()
            }

            XButton {
                imageSrc: "icons/camera.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: savePic.open()
            }

            XSeparator{}

            XToggle {
                id: tglLock
                imageSrc: "icons/lock.png"
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
            }

            Item { Layout.fillWidth: true }

            XButton {
                Layout.alignment: Layout.Right
                imageSrc: "icons/lifebuoy.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                //onClicked:
            }

            XButton {
                Layout.alignment: Layout.Right
                imageSrc: "icons/information.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                //onClicked:
            }
        }
    }

    statusBar: StatusBar {
        RowLayout {
            anchors.fill: parent
            Label {
                visible: view.keyhint.length > 0
                text: view.keyhint
            }
            Item { Layout.fillWidth: true }
            Rectangle {
                width: 10
                height: 10
                color: view.sane ? "#16a085" : "#c0392b"
            }
            Label {
                text: view.sane ? "Ready" : view.errorText
            }
            XSeparator{ color: "#2c3e50" }
            Label { text: view.label }
            XSeparator{ visible: view.lock; color: "#2c3e50" }
            Label {
                visible: view.lock
                text: "View only"
            }
        }
    }

    function takeScreenshot(name) {
        var w, h, x, y
        w = cv.canvasWindow.width
        h = cv.canvasWindow.height
        x = cv.canvasWindow.x
        y = cv.canvasWindow.y

        var scene = ctrl.prepareScene()

        cv.canvasWindow.width = scene.width
        cv.canvasWindow.height = scene.height
        cv.canvasWindow.x = scene.x
        cv.canvasWindow.y = scene.y

        var done = cv.save(name)

        cv.canvasWindow.width = w
        cv.canvasWindow.height = h
        cv.canvasWindow.x = x
        cv.canvasWindow.y = y

        return done
    }

    FileDialog {
        id: savePic
        title: "Choose file to save snapshot"
        selectExisting: false
        nameFilters: [ "PNG Images (*.png)", "All files (*)" ]
        onAccepted: {
            var ok = takeScreenshot(("" + fileUrl).replace("file://", ""))
            if(!ok) {
                ctrl.qmlError("Unable to save snapshot")
            }
        }
        onRejected: {
            ctrl.qmlError("Snapshot canceled")
        }
    }

    FileDialog {
        id: openFile
        title: "Choose file to load model"
        selectExisting: true
        nameFilters: [ "TEG files (*.teg *.json)", "All files (*)" ]
        onAccepted: {
            var path = "" + fileUrl
            ctrl.openFile(path.replace("file://", ""))
        }
        onRejected: {
            ctrl.qmlError("Opening canceled")
        }
    }

    FileDialog {
        id: saveFile
        title: "Choose file to save model"
        selectExisting: false
        nameFilters: [ "TEG files (*.teg *.json)", "All files (*)" ]
        onAccepted: {
            var path = "" + fileUrl
            ctrl.saveFile(path.replace("file://", ""))
        }
        onRejected: {
            ctrl.qmlError("Saving canceled")
        }
    }

    Canvas {
        id: cv
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
            R.render(ctx, region, renderer.cache)
        }

        onCanvasWindowChanged: {
            if(canvasWindow.width !== ctrl.canvasWindowWidth ||
                    canvasWindow.height !== ctrl.canvasWindowHeight) {
                ctrl.canvasWindowWidth = canvasWindow.width
                ctrl.canvasWindowHeight = canvasWindow.height
                ctrl.flush()
            }
        }

        Component.onCompleted: {
            canvasWindow.x = canvasSize.width / 2
            canvasWindow.y = canvasSize.height / 2
            coldstart.start()
        }
    }



    PinchArea {
        id: pinchArea
        anchors.fill: parent
        z: 10
        property real zoom: 1.0
        property real initialZoom
        onZoomChanged: ctrl.flush()

        Behavior on zoom {
            PropertyAnimation {
                duration: 50
            }
        }

        function limit(x) {
            if(x > 10.0) {
                x = 10.0
            } else if (x < 0.2) {
                x = 0.2
            }
            return x
        }

        onPinchStarted: {
            initialZoom = zoom
        }

        onPinchUpdated: {
            //   cv.canvasWindow.x += 4.0*(pinch.previousCenter.x - pinch.center.x)/pinchArea.zoom
            //   cv.canvasWindow.y += 4.0*(pinch.previousCenter.y - pinch.center.y)/pinchArea.zoom
            zoom = limit(initialZoom * pinch.scale)
        }
    }

    MouseArea {
        id: mouseArea
        anchors.fill: parent
        acceptedButtons: Qt.LeftButton | Qt.RightButton
        property real dragOffset: 50.0
        property real cx0
        property real cy0
        property real x0
        property real y0
        property real dx
        property real dy
        property bool peeked
        property bool rightPressed

        focus: true
        Keys.onPressed: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            ctrl.modifierKeyAlt = (event.modifiers & Qt.AltModifier) ? true : false

            if(ctrl.modifierKeyControl && event.key === Qt.Key_L) {
                tglLock.enabled = !tglLock.enabled
            } else {
                ctrl.keyPressed(event.key, event.text)
            }

            if(ctrl.modifierKeyControl) {
                view.keyhint = getHint(event.text.replace(/[^\u0000-\u007E]/g, ""))
            }
            event.accepted = true
        }
        Keys.onReleased: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            ctrl.modifierKeyAlt = (event.modifiers & Qt.AltModifier) ? true : false

            view.keyhint = ""
            event.accepted = true
            ctrl.flush()
        }

        function getHint(keystr) {
            var text = ""
            var plus = false
            if(ctrl.modifierKeyControl){
                text += "Ctrl"
                plus = true
            }
            if(ctrl.modifierKeyAlt) {
                text += (plus ? " + " : "") + "Alt"
                plus = true
            }
            if(ctrl.modifierKeyShift) {
                text += (plus ? " + " : "") + "Shift"
                plus = true
            }
            if(keystr.length > 0) {
                text += (plus ? " + ": "") + keystr.toUpperCase()
            }
            return text
        }

        onPressed: {
            rightPressed = (mouse.button === Qt.RightButton)
            if(view.lock || rightPressed) {
                cx0 = cv.canvasWindow.x
                cy0 = cv.canvasWindow.y
                x0 = mouse.x
                y0 = mouse.y
                dx = 0
                dy = 0
                peeked = false
            } else if(!view.lock) {
                ctrl.mousePressed(mouse.x, mouse.y)
                peeked = false
            }
        }

        onPositionChanged: {
            if(x0 != mouse.x || y0 != mouse.y) {
                if (view.lock || rightPressed) {
                    dx = (x0 - mouse.x)
                    dy = (y0 - mouse.y)
                    cv.canvasWindow.x = cx0 + dx
                    cv.canvasWindow.y = cy0 + dy
                    cv.requestPaint()
                    peeked = true
                } else if(!view.lock) {
                    if(mouse.x < 0 + dragOffset) {
                        cv.canvasWindow.x += -5.0/pinchArea.zoom
                    } else if(mouse.x > cv.canvasWindow.width - dragOffset) {
                        cv.canvasWindow.x += 5.0/pinchArea.zoom
                    }
                    if(mouse.y < 0 + dragOffset) {
                        cv.canvasWindow.y += -5.0/pinchArea.zoom
                    } else if(mouse.y > cv.canvasWindow.height - dragOffset) {
                        cv.canvasWindow.y += 5.0/pinchArea.zoom
                    }
                    ctrl.mouseMoved(mouse.x, mouse.y)
                }
            }
        }

        onReleased: {
            if (!view.lock && !rightPressed) {
                ctrl.mouseReleased(mouse.x, mouse.y)
            } else if (peeked){
                dx = dx - dx/pinchArea.zoom
                dy = dy - dy/pinchArea.zoom
                cv.canvasWindow.x -= dx
                cv.canvasWindow.y -= dy
                ctrl.flush()
            }

            peeked = false
            rightPressed = !(mouse.button === Qt.RightButton)
        }

        onDoubleClicked: {
            if (!view.lock && !rightPressed) {
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

    Ctrl {
        id: ctrl
        canvasWidth: cv.canvasSize.width
        canvasHeight: cv.canvasSize.height
        canvasWindowX: cv.canvasWindow.x
        canvasWindowY: cv.canvasWindow.y
        canvasWindowHeight: cv.canvasWindow.height
        canvasWindowWidth: cv.canvasWindow.width
        zoom: pinchArea.zoom

        onErrorTextChanged: {
            if(errorText.length > 0) {
                view.sane = false
                view.errorText = errorText
                errorHide.restart()
            } else {
                view.sane = true
                view.errorText = ""
            }
        }
    }

    Timer {
        id: errorHide
        interval: 5000
        repeat: false
        onTriggered: {
            view.sane = true
            view.errorText = ""
        }
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

