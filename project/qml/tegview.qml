import QtQuick 2.0
import QtQuick.Controls 1.1
import 'tegrender.js' as Renderer

Rectangle {
    width: 700
    height: 600
    color: "#ecf0f1"

    Button {
        id: btn0
        z: 10
        anchors.left: parent.left
        anchors.bottom: parent.bottom
        text: "+"
        onClicked: cv.zoom += 0.2
    }

    Button {
        id: btn1
        z: 10
        anchors.leftMargin: 20
        anchors.left: btn0.right
        anchors.bottom: parent.bottom
        text: "-"
        onClicked: cv.zoom -= 0.2
    }

    Button {
        id: btn2
        z: 10
        anchors.leftMargin: 20
        anchors.left: btn1.right
        anchors.bottom: parent.bottom
        text: "R"
        onClicked: {
            for(var p in model.places) {
                model.places[p].counter = rand(0,12)
                model.places[p].timer = rand(0,12)
                cv.requestPaint()
            }
        }

        function rand(min, max) {
            return Math.floor(Math.random() * (max - min + 1)) + min;
        }
    }

    Canvas {
        id: cv
        property real zoom: 1.0
        anchors.fill: parent
        tileSize.height: Math.min(cv.height, cv.width)
        tileSize.width: Math.min(cv.height, cv.width)
        canvasSize.height: 16536 * scale
        canvasSize.width: 16536 * scale
        canvasWindow.width: width
        canvasWindow.height: height
        onPaint: Renderer.render(cv, region, zoom, model)


        Component.onCompleted: {
            canvasWindow.x = canvasSize.width / 2
            canvasWindow.y = canvasSize.height / 2
            coldstart.start()
        }

        onZoomChanged: {
            cv.requestPaint()
        }

        MouseArea {
            id: drag
            anchors.fill: parent
            property int cx0
            property int cy0
            property int x0
            property int y0
            onPressed: {
                cx0 = cv.canvasWindow.x
                cy0 = cv.canvasWindow.y
                x0 = mouse.x
                y0 = mouse.y
            }
            onMouseXChanged: {
                if(x0 != mouseX) {
                    cv.canvasWindow.x = cx0 + (x0 - mouseX)
                    cv.requestPaint()
                }
            }
            onMouseYChanged: {
                if(y0 != mouseY) {
                    cv.canvasWindow.y = cy0 + (y0 - mouseY)
                    cv.requestPaint()
                }
            }
        }

        Timer {
            id: coldstart
            interval: 10
            onTriggered: cv.requestPaint()
            repeat: false
        }
    }

    Item {
        id: model
        property var places: {
            "id_1": {"control":{"x":0-100, "y":0-100}, "place": true, "x": 0, "y": 0, "selected": true, "counter": 2, "timer": 3},
            "id_2": {"place": true, "x": 60, "y": 60, "selected": true, "counter": 2, "timer": 3},
            "id_3": {"place": true, "x": 120, "y": 120, "selected": false, "counter": 2, "timer": 3},
            "id_4": {"place": true, "x": 180, "y": 180, "selected": false, "counter": 2, "timer": 3},
            "id_5": {"place": true, "x": 240, "y": 240, "selected": false, "counter": 2, "timer": 3},
            "id_6": {"place": true, "x": 300, "y": 300, "selected": true, "counter": 2, "timer": 3},
            "id_7": {"place": true, "x": 360, "y": 360, "selected": false, "counter": 2, "timer": 3},
            "id_8": {"place": true, "x": 0+100, "y": 0, "selected": true, "counter": 2, "timer": 3},
            "id_9": {"place": true, "x": 60+100, "y": 60, "selected": true, "counter": 2, "timer": 3},
            "id_10": {"place": true, "x": 120+100, "y": 120, "selected": false, "counter": 2, "timer": 3},
            "id_11": {"place": true, "x": 180+100, "y": 180, "selected": false, "counter": 2, "timer": 3},
            "id_12": {"place": true, "x": 240+100, "y": 240, "selected": false, "counter": 2, "timer": 3},
            "id_13": {"place": true, "x": 300+100, "y": 300, "selected": true, "counter": 2, "timer": 3},
            "id_14": {"place": true, "x": 360+100, "y": 360, "selected": false, "counter": 2, "timer": 3,
                "label": "Rearrangulate\nexterior #1"},
        }
        property var transitions: {
            "id_1": {"transition": true, "x": 0-200, "y": 0, "selected": true},
            "id_2": {"transition": true, "x": 60-200, "y": 60, "selected": true},
            "id_3": {"control":{"x":120-200+100, "y":120+100}, "out": 3, "transition": true, "x": 120-200, "y": 120, "selected": false},
            "id_4": {"transition": true, "x": 180-200, "y": 180, "selected": false},
            "id_5": {"transition": true, "x": 240-200, "y": 240, "selected": false},
            "id_6": {"transition": true, "x": 300-200, "y": 300, "selected": true, "counter": 2,  "label": "Rearrangulate\nexterior #1"},
            "id_7": {"transition": true, "x": 360-200, "y": 360, "selected": false, "counter": 2, "label": "Rearrangulate\nexterior #2", "horizontal": true},
        }
        property var arcs: {
            "id_1": { "start": {"type":"place", "id":"id_1"},
                "end": {"type":"transition", "id":"id_3"}, "index": 1 },
        }
    }
}
