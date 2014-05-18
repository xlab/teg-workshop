package tegview

import (
	"runtime"

	"gopkg.in/qml.v0"
)

type View struct {
	engine   *qml.Engine
	win      *qml.Window
	model    *teg
	renderer *tegRenderer
	control  *Ctrl
	closed   chan struct{}
}

func NewView(engine *qml.Engine) *View {
	model := newTeg()
	renderer := newTegRenderer(nil)
	engine.Context().SetVar("tegRenderer", renderer)
	qml.RegisterTypes("TegView", 1, 0, []qml.TypeSpec{
		{
			Init: func(ctrl *Ctrl, obj qml.Object) {
				ctrl.model = model
				ctrl.events = make(chan interface{}, 100)
				renderer.ctrl = ctrl
			},
		},
	})

	component, err := engine.LoadFile("project/qml/tegview.qml")
	if err != nil {
		panic(err)
	}

	win := component.CreateWindow(nil)
	control := win.Root().Property("ctrl").(*Ctrl)

	view := &View{
		engine:   engine,
		win:      win,
		model:    model,
		renderer: renderer,
		closed:   make(chan struct{}),
		control:  control,
	}

	win.On("closing", func() {
		close(view.closed)
	})

	return view
}

type Test struct {
	Texts []string
}

func (v *View) Show() (closed chan struct{}) {
	runtime.GOMAXPROCS(2)
	v.model.fakeData()

	go func() {
		for {
			select {
			case <-v.renderer.ready:
				qml.Changed(v.renderer, &v.renderer.Screen)
			case <-v.model.updated:
				v.renderer.task <- nil
			}
		}
	}()
	go func() {
		v.renderer.process(v.model)
	}()

	v.model.update()
	v.control.handleEvents()
	v.win.Show()

	return v.closed
}
