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
    property var layers: ctrl.layers

    property real zoom: 1.0
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
                    view.zoom = R.limit(view.zoom + 0.2)
                }
            }

            XButton {
                imageSrc: "icons/magnifier-zoom-out.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: {
                    view.zoom = R.limit(view.zoom - 0.2)
                }
            }

            XButton {
                imageSrc: "icons/magnifier-zoom-fit.png"
                original: true
                bgColor: panelBtnBgColor
                bgPressedColor: panelBtnBgPressedColor
                onClicked: {
                    view.zoom = 1.0
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

    Plane {
        anchors.fill: parent
    }

    Ctrl {
        id: ctrl
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
        id: errorHide
        interval: 5000
        repeat: false
        onTriggered: {
            view.sane = true
            view.errorText = ""
        }
    }
}
