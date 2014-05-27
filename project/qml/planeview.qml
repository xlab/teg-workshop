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
    height: 500
    color: "#ecf0f1"

    property bool sane: true
    property string errorText
    property string label: ctrl.title

    property alias ctrl: ctrl
    property var lock: tglLock.enabled || view.text
    property bool text: false
    onTextChanged: {
        if(view.text) {
            ctrl.fix()
        }
    }

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

            XToggle {
                id: tglLock
                imageSrc: "icons/lock.png"
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
            }

            XToggle {
                id: tglEraser
                imageSrc: "icons/eraser.png"
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
            Label {
                visible: ctrl.vertexText.length > 0
                text: ctrl.vertexText
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

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        RowLayout {
            spacing: 0
            Layout.fillHeight: true
            Layout.fillWidth: true

            Item {
                id: viewbox
                Layout.fillWidth: true
                Layout.fillHeight: true
                Layout.minimumHeight: 300
                Layout.minimumWidth: 300
                Repeater {
                    z: 1
                    id: layers
                    anchors.fill: parent
                    property var updated: ctrl.updated
                    onUpdatedChanged: {
                        dioid.update()
                        cv.requestPaint()
                    }

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
                    z: -1
                    id: cv
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
                    z: 10
                    id: pinchArea
                    anchors.fill: parent
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
                        } else if (x < 0.6) {
                            x = 0.6
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
                    anchors.fill: parent
                    acceptedButtons: Qt.LeftButton | Qt.RightButton

                    property real dragOffset: 50.0
                    property real cx0
                    property real cy0
                    property real x0
                    property real y0
                    property bool rightPressed

                    focus: !view.text
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
                        ctrl.mouseHovered(mouse.x, mouse.y)
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
            }

            Rectangle {
                Layout.fillHeight: true
                implicitWidth: 2
                color: "#34495e"
            }

            Rectangle {
                Layout.preferredWidth: 200
                Layout.fillHeight: true
                color: "#ecf0f1"

                Component {
                    id: layerHighlight
                    Rectangle {
                        width: 200; height: 30
                        color: "#80f1c40f"
                    }
                }

                Component {
                    id: layerItem
                    Item {
                        width: 200; height: 30
                        MouseArea {
                            anchors.fill: parent
                            onPressed: {
                                layerList.currentIndex = index
                                ctrl.setActive(index)
                            }
                            RowLayout {
                                anchors.fill: parent
                                anchors.leftMargin: 10
                                Rectangle {
                                    Layout.preferredHeight: 10
                                    Layout.preferredWidth: 15
                                    color: ctrl.colorAt(index)

                                }
                                CheckBox {
                                    checked: ctrl.enabledAt(index)
                                    Layout.fillWidth: true
                                    text: ctrl.labelAt(index)
                                    onCheckedChanged: {
                                        ctrl.setEnabledAt(index, checked)
                                    }
                                }
                            }
                        }
                    }
                }

                Component {
                    id: listHeader
                    Item {
                        width: 200; height: 30
                        Text {
                            anchors.leftMargin: 10
                            anchors.fill: parent
                            verticalAlignment: Text.AlignVCenter
                            text: layerList.count > 0 ? "Available inputs:" : "No available inputs"
                            font.pixelSize: 16
                        }
                    }
                }

                ListView {
                    id: layerList
                    width: parent.width
                    height: parent.height

                    highlight: layerHighlight
                    header: listHeader

                    delegate: layerItem
                    model: ctrl.layers.length
                    currentIndex: ctrl.activeLayer
                }
            }
        }
        Rectangle {
            Layout.fillWidth: true
            implicitHeight: 2
            color: "#34495e"
        }
        RowLayout {
            Layout.preferredHeight: 60
            Layout.fillWidth: true
            spacing: 0

            Item {
                Layout.minimumHeight: 60
                Layout.fillHeight: true
                Layout.fillWidth: true
                MouseArea {
                    visible: !view.text
                    anchors.fill: parent
                    onClicked: {
                        view.text = true
                    }
                }
                TextArea {
                    id: dioid
                    anchors.fill: parent
                    enabled: view.text
                    focus: view.text

                    textMargin: 5
                    textFormat: TextEdit.PlainText
                    font.family: "monospace"
                    font.pixelSize: 16
                    style: TextAreaStyle {
                        textColor: "#2b2b2b"
                        backgroundColor: "#ffffff"
                    }
                    function update() {
                        dioid.text = ctrl.dioid
                    }
                }
            }
            ColumnLayout {
                visible: view.text
                Button {
                    Layout.preferredWidth: 80
                    text: "Apply"
                    enabled: dioid.text.length > 0
                    onClicked: {
                        var ok = ctrl.setDioid(dioid.text)
                        if(ok) {
                            view.text = false
                            dioid.update()
                        }
                    }
                }
                Button {
                    Layout.preferredWidth: 80
                    text: "Cancel"
                    onClicked: {
                        view.text = false
                        dioid.update()
                    }
                }
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
        zoom: pinchArea.zoom
        drawShadows: !tglEraser.enabled

        onDrawShadowsChanged: {
            ctrl.flush()
        }

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
        id: coldstart
        interval: 100
        onTriggered: {
            cv.requestPaint()
        }
        repeat: false
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
