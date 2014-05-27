package tegview

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"sync"

	"github.com/xlab/teg-workshop/planeview"
	"github.com/xlab/teg-workshop/workspace"
	"gopkg.in/qml.v0"
	"gopkg.in/xlab/clipboard.v1"
)

const (
	DefaultTitle = "Untitled document"
)

type View struct {
	id        string
	model     *teg
	control   *Ctrl
	win       *qml.Window
	renderer  *tegRenderer
	planeView *planeview.View
	childs    chan workspace.Window
	closed    chan struct{}
	stop      chan struct{}
}

type actionNewWindow struct{ title string }
type actionOpenFile struct{ name string }
type actionSaveFile struct{ name string }
type actionEditGroup struct {
	model *teg
	label string
}
type actionPlaneView struct {
	models    []*planeview.Plane
	id, title string
}

func NewView() *View {
	engine := qml.NewEngine()
	model := newTeg()
	renderer := newTegRenderer(nil)
	engine.Context().SetVar("tegRenderer", renderer)
	qml.RegisterTypes("TegCtrl", 1, 0, []qml.TypeSpec{
		{
			Init: func(ctrl *Ctrl, obj qml.Object) {
				ctrl.model = model
				ctrl.events = make(chan interface{}, 100)
				ctrl.actions = make(chan interface{}, 100)
				ctrl.errors = make(chan error, 100)
				ctrl.clip = clipboard.New(engine)

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

	var once sync.Once
	quitcode := func() {
		view.control.stopHandling()
		view.model.deselectAll()
		view.model.updateParentGroups()
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

func (v *View) setModel(model *teg) {
	v.id = model.id
	v.model = model
	v.control.model = model
}

func (v *View) SetTitle(text string) {
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

func (v *View) newWindow(title string) (view *View) {
	view = NewView()
	view.SetTitle(title)
	return
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

func (v *View) openFile(name string) (err error) {
	view := v.newWindow(path.Base(name))
	if err = view.loadModel(name); err != nil {
		return
	}
	v.childs <- view
	return
}

func (v *View) editGroup(model *teg, label string) {
	view := v.newWindow(label)
	view.setModel(model)
	view.model.update()
	v.childs <- view
	return
}

func (v *View) saveFile(name string) (err error) {
	err = v.saveModel(name)
	if err != nil {
		return
	}
	v.SetTitle(path.Base(name))
	return
}

func (v *View) updatePlaneViewer() {

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
					view := v.newWindow(act.(actionNewWindow).title)
					v.childs <- view
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
					v.editGroup(info.model, info.label)
				case actionPlaneView:
					info := act.(actionPlaneView)
					view := planeview.NewView(info.id)
					v.planeView = view
					view.SetModels(info.models)
					view.SetTitle(info.title)
					v.childs <- view
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
			case <-v.model.updatedInfo:
				if v.planeView != nil {
					infos := make([]*planeview.Plane, 0, len(v.model.infos))
					for _, p := range v.model.infos {
						infos = append(infos, p)
					}
					v.planeView.SetModels(infos)
				}
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
