package main

import (
	"log"
	"os"

	"github.com/xlab/teg-workshop/tegview"
	"gopkg.in/qml.v0"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln("runtime error: " + err.Error())
		os.Exit(1)
	}
}

func run() error {
	qml.Init(nil)
	engine := qml.NewEngine()
	view := tegview.NewView(engine)
	view.Show()
	return <-view.Closed()
}
