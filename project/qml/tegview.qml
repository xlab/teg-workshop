import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import TegView 1.0
import 'tegrender.js' as Renderer

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
            ToolButton {
                text: "Rand"
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
                checked: true
                text: "View"
            }
            RadioButton {
                id: editbtn
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

        focus: true
        Keys.onPressed: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            ctrl.modifierKeyAlt = (event.modifiers & Qt.AltModifier) ? true : false
        }
        Keys.onReleased: {
            ctrl.modifierKeyShift = (event.modifiers & Qt.ShiftModifier) ? true : false
            ctrl.modifierKeyControl = (event.modifiers & Qt.ControlModifier) ? true : false
            ctrl.modifierKeyAlt = (event.modifiers & Qt.AltModifier) ? true : false
        }

        MouseArea {
            id: drag
            anchors.fill: parent
            property int cx0
            property int cy0
            property int x0
            property int y0

            onPressed: {
                if(viewMode) {
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
                    if (viewMode) {
                        cv.canvasWindow.x = cx0 + (x0 - mouse.x)
                        cv.canvasWindow.y = cy0 + (y0 - mouse.y)
                        cv.requestPaint()
                    } else if(editMode) {
                        ctrl.mouseMoved(mouse.x, mouse.y)
                    }
                }
            }

            onReleased: {
                if (editMode) {
                    ctrl.mouseReleased(mouse.x, mouse.y)
                }
            }
        }

        Timer {
            id: coldstart
            interval: 1000
            onTriggered: cv.requestPaint()
            repeat: false
        }
    }

    Item {
        id: model
        property var updated: tegModel.updated
        property var placeSpecs: tegModel.placeSpecs
        property var transitionSpecs: tegModel.transitionSpecs
        property var arcSpecs: tegModel.arcSpecs
        property var places: []
        property var transitions: []
        property var arcs: []

        onUpdatedChanged: {
            cv.requestPaint()
        }

        function preparePlaceSpec(spec) {
            return {
                "x": spec.x, "y": spec.y,
                "place": true, "selected": spec.selected,
                "counter": spec.counter, "timer": spec.timer,
                "control": {"x": spec.control.x, "y": spec.control.y},
                "label": spec.label
            }
        }

        function prepareTransitionSpec(spec) {
            return {
                "x": spec.x, "y": spec.y, "in": spec.in, "out": spec.out,
                "transition": true, "selected": spec.selected, "horizontal": spec.horizontal,
                //"control": {"x": spec.control.x, "y": spec.control.y},
                "label": spec.label
            }
        }

        onPlaceSpecsChanged: {
            // console.log("Generating new places hashmap, count=" + placeSpecs.length)
            for(var i = 0; i < placeSpecs.length; ++i){
                var spec = placeSpecs.value(i)
                places[i] = preparePlaceSpec(spec)
            }
        }

        onTransitionSpecsChanged: {
            // console.log("Generating new transitions hashmap, count=" + transitionSpecs.length)
            for(var i = 0; i < transitionSpecs.length; ++i){
                var spec = transitionSpecs.value(i)
                transitions[i] = prepareTransitionSpec(spec)
            }
        }

        onArcSpecsChanged: {
            // console.log("Generating new arcs hashmap, count=" + arcSpecs.length)
            for(var i = 0; i < arcSpecs.length; ++i){
                var a = arcSpecs.value(i)
                arcs[i] = {
                    "place": preparePlaceSpec(a.place), "transition": prepareTransitionSpec(a.transition),
                    "control": a.place.control, "index": a.index, "inbound": a.inbound,
                }
            }
        }
    }
}

