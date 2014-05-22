import QtQuick 2.0
import QtGraphicalEffects 1.0

Item {
    id: button

    property string imageSrc
    property bool original: false
    property int fontSize: 32
    property string text
    property string fontFamily: "Helvetica Neue"

    property string fgColor: "black"
    property string bgColor: "transparent"
    property string fgPressedColor
    property string bgPressedColor
    property bool pressed: false
    signal clicked()

    width: 32
    height: 32
    property int radius: 8
    opacity: enabled ? 1 : 0.3

    Rectangle {
        radius: button.radius
        smooth: false
        id: background
        anchors.fill: parent
        color: if (bgPressedColor && pressed) {
                   return bgPressedColor
               } else { return button.bgColor }
        opacity: (pressed && !bgPressedColor) ? 0.5 : 1
    }

    Image {
        z: 2
        id: img
        visible: original
        anchors.fill: parent
        fillMode: Image.Pad
        source: button.imageSrc
    }

    ColorOverlay {
        z: 2
        source: img
        visible: button.imageSrc && !original
        anchors.fill: parent
        color: pressed ? fgPressedColor : fgColor
    }

    Text {
        z: 3
        visible: !button.imageSrc
        anchors.centerIn: parent
        font.family: button.fontFamily
        font.pointSize: button.fontSize
        color: pressed ? fgPressedColor : fgColor
        text: button.text
        font.capitalization: Font.AllUppercase
    }

    MouseArea {
        anchors.fill: parent
        onClicked: button.clicked()
        onPressed: button.pressed = true
        onReleased: button.pressed = false
    }
}
