import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Controls.Styles 1.1
import QtQuick.Layouts 1.1
import QtQuick.Dialogs 1.1
import PlaneCtrl 1.0
import 'planerender.js' as R

ApplicationWindow {
    id: view
    width: 600
    height: 400
    color: "#ecf0f1"

    property bool sane: true
    property string errorText
    property string label: ctrl.title

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
        } else {
            ctrl.flush()
            cv.requestPaint()
        }
    }

    title: "I/O Editor"
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
                imageSrc: "icons/camera.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: ctrl.test()
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
        }
    }

    statusBar: StatusBar {
        RowLayout {
            anchors.fill: parent
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

    FileDialog {
        id: savePic
        title: "Choose file to save snapshot"
        selectExisting: false
        nameFilters: [ "PNG Images (*.png)", "All files (*)" ]
        onAccepted: {
            var path = "" + fileUrl
            ctrl.saveFile(path.replace("file://", ""))
        }
        onRejected: {
            ctrl.qmlError("Snapshot canceled")
        }
    }

    Repeater {
        id: layers
        anchors.fill: parent
        property var updated: ctrl.updated
        onUpdatedChanged: layers.repaint()

        z: 1

        model: ctrl.layers.length
        PlaneLayer {
            anchors.fill: parent
            screen: ctrl.layers.at(index)
            layerId: index

            canvasSize.width: cv.canvasSize.width
            canvasSize.height: cv.canvasSize.height
            canvasWindow.width: cv.canvasWindow.width
            canvasWindow.height: cv.canvasWindow.height
            canvasWindow.x: cv.canvasWindow.x
            canvasWindow.y: cv.canvasWindow.y
        }

        function repaint() {
            for(var i = 0; i < layers.count; i++){
                layers.itemAt(i).requestPaint()
            }
        }
    }

    Canvas {
        id: cv

        z: -1

        anchors.fill: parent
        canvasSize.width: 16536
        canvasSize.height: 16536
        canvasWindow.width: width
        canvasWindow.height: height
        tileSize: "1024x1024"

        onPainted: {
            layers.repaint()
        }

        onPaint: {
            var ctx = cv.getContext("2d")
            R.background(ctx, region, pinchArea.zoom,
                         cv.canvasSize.width/2, cv.canvasSize.height/2,
                         cv.canvasWindow.width/2, cv.canvasWindow.height/2,
                         cv.canvasWindow.x, cv.canvasWindow.y)
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
        onZoomChanged: {
            cv.requestPaint()
            ctrl.flush()
        }

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
        property bool rightPressed

        focus: true
        Keys.onPressed: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false

            if(ctrl.modifierKeyControl && event.key === Qt.Key_L) {
                tglLock.enabled = !tglLock.enabled
            } else {
                ctrl.keyPressed(event.key, event.text)
            }

            event.accepted = true
        }
        Keys.onReleased: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            event.accepted = true
            ctrl.flush()
        }

        onPressed: {
            rightPressed = (mouse.button === Qt.RightButton)
            if(view.lock || rightPressed) {
                cx0 = cv.canvasWindow.x
                cy0 = cv.canvasWindow.y
                x0 = mouse.x
                y0 = mouse.y
            } else if(!view.lock) {
                ctrl.mousePressed(mouse.x, mouse.y)
            }
        }

        onPositionChanged: {
            if(x0 != mouse.x || y0 != mouse.y) {
                if (view.lock || rightPressed) {
                    cv.canvasWindow.x = cx0 + (x0 - mouse.x)
                    cv.canvasWindow.y = cy0 + (y0 - mouse.y)
                    cv.requestPaint()
                } else if(!view.lock) {
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
                    cv.requestPaint()
                }
            }
        }

        onReleased: {
            if (!view.lock && !rightPressed) {
                ctrl.mouseReleased(mouse.x, mouse.y)
            }

            rightPressed = !(mouse.button === Qt.RightButton)
        }
    }

    Timer {
        id: coldstart
        interval: 100
        onTriggered: {
            cv.requestPaint()
        }
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
}
