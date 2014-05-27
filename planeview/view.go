package planeview

import (
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/xlab/teg-workshop/workspace"
	"gopkg.in/qml.v0"
)

const (
	Title = "I/O of %s"
)

var (
	EmptyImage = image.NewRGBA(image.Rect(0, 0, 10, 10))
)

type View struct {
	id      string
	control *Ctrl
	win     *qml.Window
	childs  chan workspace.Window
	closed  chan struct{}
	stop    chan struct{}
	updated chan int
	updates chan interface{}
	visible bool
}

func NewView(id string) *View {
	engine := qml.NewEngine()
	qml.RegisterTypes("PlaneCtrl", 1, 0, []qml.TypeSpec{
		{
			Init: func(ctrl *Ctrl, obj qml.Object) {
				ctrl.Layers = &Layers{}
				ctrl.ActiveLayer = -1
				ctrl.enabled = make(map[string]bool)
				ctrl.events = make(chan interface{}, 100)
				ctrl.actions = make(chan interface{}, 100)
				ctrl.errors = make(chan error, 100)
			},
		},
	})

	component, err := engine.LoadFile("project/qml/planeview.qml")
	if err != nil {
		panic(err)
	}

	win := component.CreateWindow(nil)
	control := win.Root().Property("ctrl").(*Ctrl)

	view := &View{
		id:      id,
		win:     win,
		childs:  make(chan workspace.Window, 10),
		closed:  make(chan struct{}),
		stop:    make(chan struct{}),
		updated: make(chan int, 100),
		control: control,
	}

	var once sync.Once
	quitcode := func() {
		view.visible = false
		view.control.stopHandling()
		close(view.stop)
		close(view.childs)
		close(view.closed)
		win.Destroy()
	}
	quit := func() {
		once.Do(quitcode)
	}
	win.On("closing", quit)
	engine.On("quit", quit)

	return view
}

func (v *View) SetModels(models []*Plane) {
	for i, m := range models {
		m.id = i
		m.updated = v.updated
	}
	if len(models) > 0 {
		v.control.models = models
		for i, m := range models {
			if m.ioId == v.control.Layers.active {
				v.control.ActiveLayer = i
				break
			}
		}
	} else {
		v.control.models = nil
		v.control.Layers = &Layers{}
		v.control.ActiveLayer = -1
		qml.Changed(v.control, &v.control.Layers)
		qml.Changed(v.control, &v.control.Updated)
		return
	}
	renderers := make([]*planeRenderer, 0, len(models))
	for i := 0; i < len(models); i++ {
		renderers = append(renderers, newPlaneRenderer(v.control))
	}
	layers := &Layers{
		renderers: renderers, Length: len(renderers),
		active: v.control.Layers.active,
	}
	v.control.Layers = layers
	qml.Changed(v.control, &v.control.Layers)
	qml.Changed(v.control, &v.control.Updated)
	if v.visible {
		for i := range models {
			v.updated <- i
		}
	}
}

func (v *View) SetTitle(text string) {
	v.control.Title = fmt.Sprintf(Title, text)
	qml.Changed(v.control, &v.control.Title)
}

func (v *View) Id() string {
	return "plane_" + v.id
}

func (v *View) SetError(err error) {
	v.control.Error(err)
}

func (v *View) Childs() chan workspace.Window {
	return v.childs
}

func (v *View) Show() chan struct{} {
	go func() {
		for {
			select {
			case <-v.stop:
				return
			case err := <-v.control.errors:
				if err != nil {
					log.Println(err)
					v.control.ErrorText = err.Error()
					qml.Changed(v.control, &v.control.ErrorText)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-v.stop:
				return
			case id := <-v.updated:
				if id >= len(v.control.models) {
					continue
				}
				m := v.control.models[id]
				v.control.Layers.renderers[id].process(m)
				qml.Changed(v.control, &v.control.Layers)
				qml.Changed(v.control, &v.control.Updated)
			}
		}
	}()

	v.visible = true
	v.control.handleEvents()
	v.win.Show()

	return v.closed
}

type Layers struct {
	active    string
	renderers []*planeRenderer
	Length    int
}

func (l *Layers) At(i int) *PlaneBuffer {
	if i < len(l.renderers) {
		return l.renderers[i].Screen
	}
	return newPlaneBuffer()
}
