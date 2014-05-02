package tegview

import "gopkg.in/qml.v0"

type View struct {
	engine  *qml.Engine
	win     *qml.Window
	model   *teg
	control *Ctrl
	closed  chan struct{}
}

func NewView(engine *qml.Engine) *View {
	model := newTeg()
	engine.Context().SetVar("baseTeg", model)
	qml.RegisterTypes("TegView", 1, 0, []qml.TypeSpec{
		{
			Init: func(ctrl *Ctrl, obj qml.Object) {
				ctrl.events = make(chan interface{}, 100)
			},
		},
	})

	component, err := engine.LoadFile("project/qml/tegview.qml")
	if err != nil {
		panic(err)
	}

	win := component.CreateWindow(nil)
	control := win.Root().Property("ctrl").(*Ctrl)
	control.model = model

	view := &View{
		engine:  engine,
		win:     win,
		model:   model,
		closed:  make(chan struct{}),
		control: control,
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
	v.model.fakeData()
	v.model.updated()
	v.control.handleEvents()
	v.win.Show()
	/*
		f, err := os.Create("teg.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		<-v.closed
	*/
	return v.closed
}
