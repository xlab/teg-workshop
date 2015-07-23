package main

import (
	"runtime"

	"github.com/xlab/teg-workshop/tegview"
	"github.com/xlab/teg-workshop/workspace"
	"gopkg.in/qml.v1"
)

var group = workspace.NewGroup()

func init() {
	runtime.GOMAXPROCS(4)
}

func main() {
	qml.Run(func() error {
		root := tegview.NewView()
		group.AddWindow(root)
		group.Wait()
		return nil
	})
}
