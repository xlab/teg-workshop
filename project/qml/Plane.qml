import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Controls.Styles 1.1
import QtQuick.Layouts 1.1
import QtQuick.Dialogs 1.1
import PlaneCtrl 1.0
import 'planerender.js' as R

ColumnLayout {
    spacing: 0

    Timer {
        id: coldstart
        interval: 100
        onTriggered: {
            cv.requestPaint()
        }
        repeat: false
    }

    property string panelBtnFgColor: "black"
    property string panelBtnBgColor: "#15000000"
    property string panelBtnFgPressedColor: "white"
    property string panelBtnBgPressedColor: "#3498db"

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

            focus: !view.text

            Keys.onPressed: {
                ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
                ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false

                if(ctrl.modifierKeyControl && event.key === Qt.Key_L) {
                    tglLock.enabled = !tglLock.enabled
                } else if(ctrl.modifierKeyControl && event.key === Qt.Key_W) {
                    Qt.quit()
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

            Item {
                anchors.top: parent.top
                anchors.left: parent.left
                anchors.leftMargin: 15
                anchors.topMargin: 15
                z: 15
                XButton {
                    id: starAct
                    imageSrc: "icons/star.png"
                    original: true
                    bgColor: panelBtnBgColor
                    bgPressedColor: panelBtnBgPressedColor
                    onClicked: ctrl.star()
                }

                XButton {
                    id: applyAct
                    imageSrc: "icons/tick.png"
                    original: true
                    bgColor: panelBtnBgColor
                    bgPressedColor: panelBtnBgPressedColor
                    onClicked: ctrl.fix()
                }
            }

            Repeater {
                z: 1
                id: layerStack
                anchors.fill: parent
                property var updated: ctrl.updated
                onUpdatedChanged: {
                    if(layers.length < 1) {
                        view.text = false
                    }

                    dioid.update()
                    cv.requestPaint()

                    var selected = ctrl.definedSelected()
                    var temporary = ctrl.hasTemporary()
                    starAct.visible = selected && !temporary
                    applyAct.visible = temporary
                }

                model: layers.length
                PlaneLayer {
                    anchors.fill: parent
                    screen: layers.at(index)
                    layerId: index

                    function setEnabled(){
                        visible = ctrl.enabledAt(index)
                    }

                    canvasSize.width: cv.canvasSize.width
                    canvasSize.height: cv.canvasSize.height
                    canvasWindow.width: cv.canvasWindow.width
                    canvasWindow.height: cv.canvasWindow.height
                    canvasWindow.x: cv.canvasWindow.x
                    canvasWindow.y: cv.canvasWindow.y
                }

                function repaint() {
                    for(var i = 0; i < layerStack.count; i++){
                        layerStack.itemAt(i).setEnabled()
                        layerStack.itemAt(i).requestPaint()
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
                    layerStack.repaint()
                }

                onPaint: {
                    var ctx = cv.getContext("2d")
                    R.background(ctx, region, pinchArea.zoom,
                                 cv.canvasSize.width/2, cv.canvasSize.height/2,
                                 cv.canvasWindow.width/2, cv.canvasWindow.height/2,
                                 cv.canvasWindow.x, cv.canvasWindow.y)
                }

                onCanvasWindowChanged: {
                    ctrl.canvasWindowX = canvasWindow.x
                    ctrl.canvasWindowY = canvasWindow.y
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
                    ctrl.canvasWidth = canvasSize.width
                    ctrl.canvasHeight = canvasSize.height
                    coldstart.start()
                }
            }

            PinchArea {
                z: 10
                id: pinchArea
                anchors.fill: parent
                property real zoom: view.zoom
                property real initialZoom
                onZoomChanged: {
                    ctrl.zoom = view.zoom
                    cv.requestPaint()
                    ctrl.flush()
                }

                Behavior on zoom {
                    PropertyAnimation {
                        duration: 50
                    }
                }

                onPinchStarted: {
                    initialZoom = view.zoom
                }

                onPinchUpdated: {
                    view.zoom = R.limit(initialZoom * pinch.scale)
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
                            if(ctrl.enabledAt(index)) {
                                layerList.currentIndex = index
                                ctrl.setActive(index)
                            } else {
                                layerList.currentIndex = -1
                                ctrl.setActive(-1)
                            }
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
                                property var updated: ctrl.updated
                                onUpdatedChanged: {
                                    text = ctrl.labelAt(index)
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
                model: layers.length
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
                    if(layers.length > 0) {
                        view.text = true
                    }
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
                textColor: "#2b2b2b"
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
