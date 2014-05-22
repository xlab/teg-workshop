import QtQuick 2.0
import QtGraphicalEffects 1.0

Item {
    id: toggle

    property string imageSrc
    property int fontSize: 32
    property string text
    property string fontFamily: "Helvetica Neue"
    property string fgColor: "black"
    property string bgColor: "transparent"
    property string fgPressedColor
    property string bgPressedColor
    property bool pressed: false

    signal toggled()
    property bool enabled

    // Only emit toggled(),
    // do not actually toggle 'enabled' property
    property bool onlySignal

    width: 32
    height: 32
    property int radius: 8
    opacity: enabled ? 1 : 0.3

    Rectangle {
        radius: toggle.radius
        smooth: false
        id: background
        anchors.fill: parent
        color: if (bgPressedColor && pressed) {
                   return bgPressedColor
               } else { return toggle.bgColor }
        opacity: (pressed && !bgPressedColor) ? 0.5 : 1
    }

    Image {
        z: 2
        id: img
        visible: toggle.imageSrc
        anchors.fill: parent
        fillMode: Image.Pad
        source: toggle.imageSrc
    }

    Text {
        z: 3
        anchors.centerIn: parent
        font.family: toggle.fontFamily
        font.pointSize: toggle.fontSize
        color: pressed ? fgPressedColor : fgColor
        text: toggle.text
        font.capitalization: Font.AllUppercase
    }

    MouseArea {
        anchors.fill: parent
        onClicked: {
            if(!toggle.onlySignal) {
                toggle.enabled = !toggle.enabled
            }
            toggled()
        }
        onPressed: toggle.pressed = true
        onReleased: toggle.pressed = false
    }
}
