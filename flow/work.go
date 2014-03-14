package flow

import (
	"reflect"
	"strings"
	"sync"

	"github.com/golang/glog"
)

// Work keeps track of internal details about a worker.
type Work struct {
	worker Worker
	name   string
	group  *Group
	alive  bool

	mutex   sync.Mutex
	inputs  map[string]*connection
	outputs map[string]*connection
}

func (w *Work) initWork(wi Worker, nm string, gr *Group) *Work {
	if w.group != nil {
		glog.Fatalln("worker is already in use:", nm)
	}
	w.worker = wi
	w.name = nm
	w.group = gr
	w.inputs = map[string]*connection{}
	w.outputs = map[string]*connection{}
	return w
}

func (w *Work) workerValue() reflect.Value {
	return reflect.ValueOf(w.worker).Elem()
}

func (w *Work) portValue(port string) reflect.Value {
	pp := portPart(port)
	// if it's a group, look up mapped ports
	if g, ok := w.worker.(*Group); ok {
		p := g.portMap[pp]
		return g.workerOf(p).portValue(p) // recursive
	}
	fv := w.workerValue().FieldByName(pp)
	if !fv.IsValid() {
		glog.Fatalln("port not found:", port)
	}
	return fv
}

func (w *Work) getInput(port string, capacity int) *connection {
	c := w.inputs[port]
	if c == nil {
		c = &connection{channel: make(chan Memo, capacity), dest: w}
		w.inputs[port] = c
	}
	if capacity > c.capacity {
		c.capacity = capacity
	}
	return c
}

func (w *Work) setOutput(port string, c *connection) {
	ppfv := strings.Split(port, ":")
	fp := w.portValue(ppfv[0])
	if len(ppfv) == 1 {
		if !fp.IsNil() {
			glog.Fatalf("output already connected: %s.%s", w.name, port)
		}
		fp.Set(reflect.ValueOf(c))
	} else { // it's not an Output, so it must be a map[string]Output
		if fp.IsNil() {
			fp.Set(reflect.ValueOf(map[string]Output{}))
		}
		outputs := fp.Interface().(map[string]Output)
		if _, ok := outputs[ppfv[1]]; ok {
			glog.Fatalf("output already connected: %s.%s", w.name, port)
		}
		outputs[ppfv[1]] = c
	}
	c.senders++
	w.outputs[port] = c
}

func (w *Work) setupChannels() {
	we := w.workerValue()
	sink := &fakeSink{}
	null := make(chan Memo)
	close(null)

	// make sure all the inbox connections have also been set up
	for dest, memos := range w.group.inbox {
		if workerPart(dest) == w.name {
			w.getInput(dest, len(memos)) // will add connection to input map
		}
	}

	// connect all dangling outputs to a fake sink
	for i := 0; i < we.NumField(); i++ {
		fe := we.Field(i)
		if fe.Type().String() == "flow.Output" && fe.IsNil() {
			fe.Set(reflect.ValueOf(sink))
		}
	}

	// set up and pre-fill all the input ports
	for p, c := range w.inputs {
		// create a channel with the proper capacity
		c.channel = make(chan Memo, c.capacity)
		w.portValue(p).Set(reflect.ValueOf(c.channel))
		// fill it with memos from the inbox, if any
		for _, m := range w.group.inbox[p] {
			c.channel <- m
		}
		// close the channel if there is no other feed
		if c.senders == 0 {
			close(c.channel)
		}
	}

	// set all remaining inputs to a dummy null input
	for i := 0; i < we.NumField(); i++ {
		fe := we.Field(i)
		if fe.Type().String() == "flow.Input" && fe.IsNil() {
			fe.Set(reflect.ValueOf(null))
		}
	}
}

func (w *Work) closeChannels() {
	for _, c := range w.outputs {
		c.Close()
	}
	for p, c := range w.inputs {
		c.channel = nil
		w.portValue(p).Set(reflect.ValueOf(c.channel))
	}
}

func (w *Work) launch() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.alive {
		return
	}

	w.alive = true
	w.group.wait.Add(1)
	w.setupChannels()

	go func() {
		defer w.group.wait.Done()
		defer DontPanic()

		w.worker.Run()

		w.mutex.Lock()
		defer w.mutex.Unlock()
		w.alive = false

		w.closeChannels()
	}()
}
