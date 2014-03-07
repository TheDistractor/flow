package flow

import (
	"fmt"
	"reflect"
)

// Work keeps track of internal details about a worker.
type Work struct {
	worker  Worker
	name    string
	parent  *Group
	inbox   map[string][]Memo
	inputs  map[string]*connection
	outputs map[string]*connection
}

func (w *Work) initWork(wi Worker, nm string, gr *Group) *Work {
	if w.parent != nil {
		panic("worker is already in use: " + nm)
	}
	w.worker = wi
	w.name = nm
	w.parent = gr
	w.inbox = map[string][]Memo{}
	w.inputs = map[string]*connection{}
	w.outputs = map[string]*connection{}
	return w
}

// Return the group of this worker.
func (w *Work) MyGroup() *Group {
	return w.parent
}

// Return the name of this worker.
func (w *Work) MyName() string {
	return w.name
}

func (w *Work) port(p string) reflect.Value {
	wp := reflect.ValueOf(w.worker)
	wv := wp.Elem()
	fv := wv.FieldByName(p)
	if !fv.IsValid() {
		// maybe it's a group with mapped ports
		if g, ok := w.worker.(*Group); ok {
			if p, ok := g.portMap[p]; ok {
				fw := g.workerOf(p)
				return fw.port(portPart(p)) // recursive
			}
		}
		fmt.Println("port not found: " + p)
	}
	return fv
}

func (w *Work) addToInbox(port string, value Memo) {
	w.inbox[port] = append(w.inbox[port], value)
}

func (w *Work) processInbox() {
	for dest, memos := range w.inbox {
		c := make(chan Memo, len(memos))
		dp := w.port(dest)
		dp.Set(reflect.ValueOf(c))
		for _, m := range memos {
			c <- m
		}
		close(c)
		delete(w.inbox, dest)
	}
}

func (w *Work) forAllPorts(f func(string, reflect.Value)) {
	wv := reflect.ValueOf(w.worker)
	we := wv.Elem()
	wt := we.Type()
	for i := 0; i < we.NumField(); i++ {
		fd := wt.Field(i)
		ft := fd.Type.Name()
		switch ft {
		case "Input", "Output":
			f(ft, we.Field(i))
		}
	}
	return
}

// use a fake sink for every output port not connected to anything else
type fakeSink struct{}

func (c *fakeSink) Send(m Memo) {
	fmt.Printf("Lost %T: %v\n", m, m)
}

func (c *fakeSink) Close() {}

func (w *Work) connectChannels() {
	sink := &fakeSink{}
	null := make(chan Memo)
	close(null)

	w.forAllPorts(func(typ string, val reflect.Value) {
		if val.IsNil() {
			switch typ {
			case "Input":
				val.Set(reflect.ValueOf(null))
			case "Output":
				val.Set(reflect.ValueOf(sink))
			}
		}
	})
}

func (w *Work) closeChannels() {
	for _, c := range w.outputs {
		c.Close()
	}
}

func (w *Work) Launch() {
	w.parent.wait.Add(1)
	go func() {
		defer w.parent.wait.Done()

		w.processInbox()
		w.connectChannels()
		w.worker.Run()
		w.closeChannels()
	}()
}
