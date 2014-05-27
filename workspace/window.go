package workspace

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrTooMany   = errors.New("workspace: too many opened windows")
	ErrNotUnique = errors.New("workspace: target window already opened")
)

type Group struct {
	sync.WaitGroup
	ids   idMap
	count int32
}

func NewGroup() *Group {
	return &Group{
		ids: idMap{ids: make(map[string]bool, 10)},
	}
}

type idMap struct {
	sync.RWMutex
	ids map[string]bool
}

func (i *idMap) add(id string) {
	i.Lock()
	i.ids[id] = true
	i.Unlock()
}

func (i *idMap) exists(id string) (ok bool) {
	i.RLock()
	_, ok = i.ids[id]
	i.RUnlock()
	return
}

func (i *idMap) remove(id string) {
	i.Lock()
	delete(i.ids, id)
	i.Unlock()
}

func (g *Group) AddWindow(w Window) (err error) {
	if atomic.LoadInt32(&g.count) >= 10 {
		return ErrTooMany
	}
	if g.ids.exists(w.Id()) {
		return ErrNotUnique
	}
	g.ids.add(w.Id())
	atomic.AddInt32(&g.count, 1)
	g.Add(1)
	go func() {
		for ch := range w.Childs() {
			if err := g.AddWindow(ch); err != nil {
				w.SetError(err)
			}
		}
	}()
	go func() {
		<-w.Show()
		g.ids.remove(w.Id())
		atomic.AddInt32(&g.count, -1)
		g.Done()
	}()
	return
}

type Window interface {
	Id() string
	Show() chan struct{}
	Childs() chan Window
	SetError(err error)
}
