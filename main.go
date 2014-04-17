package main

import (
	"github.com/xlab/teg-workshop/tegview"
	"gopkg.in/qml.v0"
)

func main() {
	qml.Init(nil)
	engine := qml.NewEngine()
	view := tegview.NewView(engine)
	<-view.Show()
}
