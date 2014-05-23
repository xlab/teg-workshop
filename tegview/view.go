package tegview

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/xlab/teg-workshop/workspace"
	"gopkg.in/qml.v0"
)

const (
	DefaultTitle = "Untitled document"
)

type View struct {
	id       string
	model    *teg
	control  *Ctrl
	win      *qml.Window
	renderer *tegRenderer
	childs   chan workspace.Window
	closed   chan struct{}
	stop     chan struct{}
}

type actionNewWindow struct{ title string }
type actionOpenFile struct{ name string }
type actionSaveFile struct{ name string }
type actionEditGroup struct {
	model *teg
	label string
}

func NewView() *View {
	engine := qml.NewEngine()
	model := newTeg()
	renderer := newTegRenderer(nil)
	engine.Context().SetVar("tegRenderer", renderer)
	qml.RegisterTypes("TegView", 1, 0, []qml.TypeSpec{
		{
			Init: func(ctrl *Ctrl, obj qml.Object) {
				ctrl.model = model
				ctrl.events = make(chan interface{}, 100)
				ctrl.actions = make(chan interface{}, 100)
				ctrl.errors = make(chan error, 100)

				renderer.ctrl = ctrl

				ctrl.Title = DefaultTitle
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
		win:      win,
		model:    model,
		renderer: renderer,
		childs:   make(chan workspace.Window, 10),
		closed:   make(chan struct{}),
		stop:     make(chan struct{}),
		control:  control,
		id:       model.id,
	}

	win.On("closing", func() {
		view.control.stopHandling()
		view.model.deselectAll()
		view.model.updateParentGroups()
		close(view.stop)
		view.model = nil
		view.control.model = nil
		view.control = nil
		close(view.childs)
		close(view.closed)
	})

	return view
}

type Test struct {
	Texts []string
}

func (v *View) setModel(model *teg) {
	v.id = model.id
	v.model = model
	v.control.model = model
}

func (v *View) setTitle(text string) {
	if len(text) > 0 {
		v.control.Title = text
	} else {
		v.control.Title = DefaultTitle
	}
	qml.Changed(v.control, &v.control.Title)
}

func (v *View) FakeModel() {
	v.model.fakeData()
}

func (v *View) saveModel(path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(v.model); err != nil {
		return
	}
	return
}

func (v *View) loadModel(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err = dec.Decode(v.model); err != nil {
		return
	}
	return
}

func (v *View) newWindow(title string) (view *View, err error) {
	view = NewView()
	view.setTitle(title)
	return
}

func (v *View) openFile(name string) (err error) {
	view, err := v.newWindow(path.Base(name))
	if err != nil {
		return
	}
	if err = view.loadModel(name); err != nil {
		return
	}
	v.childs <- view
	return
}

func (v *View) editGroup(model *teg, label string) (err error) {
	view, err := v.newWindow(label)
	if err != nil {
		return
	}
	view.setModel(model)
	view.model.update()
	v.childs <- view
	return
}

func (v *View) saveFile(name string) (err error) {
	err = v.saveModel(name)
	return
}

func (v *View) Id() string {
	return v.id
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
			case act := <-v.control.actions:
				switch act.(type) {
				case actionNewWindow:
					if win, err := v.newWindow(act.(actionNewWindow).title); err != nil {
						v.control.Error(err)
					} else {
						v.childs <- win
					}
				case actionOpenFile:
					if err := v.openFile(act.(actionOpenFile).name); err != nil {
						v.control.Error(err)
					}
				case actionSaveFile:
					if err := v.saveFile(act.(actionSaveFile).name); err != nil {
						v.control.Error(err)
					}
				case actionEditGroup:
					info := act.(actionEditGroup)
					if err := v.editGroup(info.model, info.label); err != nil {
						v.control.Error(err)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-v.stop:
				return
			case <-v.renderer.ready:
				qml.Changed(v.renderer, &v.renderer.Screen)
			case <-v.model.updated:
				v.renderer.task <- nil
			}
		}
	}()
	go func() {
		for {
			select {
			case <-v.stop:
				return
			default:
				v.renderer.process(v.model)
			}
		}
	}()

	v.model.update()
	v.control.handleEvents()
	v.win.Show()

	return v.closed
}
