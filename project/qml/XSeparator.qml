import QtQuick 2.0

Item {
    id: sep
    implicitWidth: 16
    anchors.top: parent.top
    anchors.bottom: parent.bottom

    property real space: 3
    property string color: "#999"
    Rectangle {
        width: sep.space
        height: parent.height
        anchors.centerIn: parent
        color: sep.color
    }
}
