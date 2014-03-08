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
	w.inputs = map[string]*connection{}
	w.outputs = map[string]*connection{}
	return w
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
		fmt.Println("port not found:", p)
	}
	return fv
}

func (w *Work) processInbox() {
	for dest, memos := range w.parent.inbox[w.name] {
		c := make(chan Memo, len(memos))
		dp := w.port(dest)
		dp.Set(reflect.ValueOf(c))
		for _, m := range memos {
			c <- m
		}
		close(c)
	}
}

func (w *Work) forAllPorts(f func(string, reflect.Value)) {
	wv := reflect.ValueOf(w.worker)
	we := wv.Elem()
	wt := we.Type()
	for i := 0; i < we.NumField(); i++ {
		fd := wt.Field(i)
		ft := fd.Type.Name()
		switch {
		case ft == "Input" || ft == "Output":
			f(ft, we.Field(i))
		case fd.Type.String() == "map[string]flow.Output":
			// TODO: hack, won't be sufficient when adding multi-in ports
			for k, _ := range we.Field(i).Interface().(map[string]Output) {
				f(k, we.Field(i))
			}
		}
	}
	return
}

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
			default:
				val.Interface().(map[string]Output)[typ] = sink
			}
		}
	})
}

func (w *Work) closeChannels() {
	for _, c := range w.outputs {
		c.Close()
	}
}

func (w *Work) launch() {
	w.parent.wait.Add(1)
	go func() {
		defer w.parent.wait.Done()

		w.processInbox()
		w.connectChannels()
		w.worker.Run()
		w.closeChannels()
	}()
}
