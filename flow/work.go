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

	mutex   sync.Mutex
	alive   bool
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
		setValue(fp, c)
	} else { // it's not an Output, so it must be a map[string]Output
		if fp.IsNil() {
			setValue(fp, map[string]Output{})
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
	// make sure all the inbox connections have also been set up
	for dest, memos := range w.group.inbox {
		if workerPart(dest) == w.name {
			w.getInput(dest, len(memos)) // will add connection to input map
		}
	}

	// set up and pre-fill all the input ports
	for p, c := range w.inputs {
		// create a channel with the proper capacity
		c.channel = make(chan Memo, c.capacity)
		setValue(w.portValue(p), c.channel)
		// fill it with memos from the inbox, if any
		for _, m := range w.group.inbox[p] {
			c.channel <- m
		}
		// close the channel if there is no other feed
		if c.senders == 0 {
			close(c.channel)
		}
	}

	// set dangling inputs to a null input and dangling outputs to a fake sink
	sink := &fakeSink{}
	null := make(chan Memo)
	close(null)

	we := w.workerValue()
	for i := 0; i < we.NumField(); i++ {
		fe := we.Field(i)
		switch fe.Type().String() {
		case "flow.Input":
			if fe.IsNil() {
				setValue(fe, null)
			}
		case "flow.Output":
			if fe.IsNil() {
				setValue(fe, sink)
			}
		}
	}
}

func (w *Work) isFinished() bool {
	for _, c := range w.inputs {
		if len(c.channel) > 0 {
			return false
		}
	}
	return true
}

func (w *Work) closeChannels() {
	for p, c := range w.inputs {
		c.channel = nil
		setValue(w.portValue(p), c.channel)
	}
	for _, c := range w.outputs {
		c.Close()
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

		for {
			w.worker.Run()
			if w.isFinished() {
				break
			}
		}

		w.mutex.Lock()
		defer w.mutex.Unlock()
		w.alive = false

		w.closeChannels()
	}()
}

func setValue(val reflect.Value, any interface{}) {
	val.Set(reflect.ValueOf(any))
}
