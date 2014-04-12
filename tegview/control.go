package tegview

import "gopkg.in/qml.v0"

type View struct {
	engine *qml.Engine
	win    *qml.Window
	closed chan error
}

func NewView(engine *qml.Engine) *View {
	component, err := engine.LoadFile("project/qml/tegview.qml")
	if err != nil {
		panic(err)
	}

	view := &View{
		engine: engine,
		win:    component.CreateWindow(nil),
		closed: make(chan error, 2),
	}
	engine.On("quit", view.close)
	view.win.On("closing", view.close)

	return view
}

func (v *View) Show() {
	v.win.Show()
}

func (v *View) Closed() <-chan error {
	return v.closed
}

func (v *View) close() {
	v.closed <- nil
}
