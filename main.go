package main

import (
	"runtime"

	"github.com/xlab/teg-workshop/tegview"
	"github.com/xlab/teg-workshop/workspace"
	"gopkg.in/qml.v0"
)

var group = workspace.NewGroup()

func main() {
	qml.Init(nil)
	runtime.GOMAXPROCS(2)
	root := tegview.NewView()
	root.FakeModel()
	group.AddWindow(root)
	group.Wait()
}
