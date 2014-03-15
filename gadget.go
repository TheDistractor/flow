package flow

import (
	"reflect"
	"strings"

	"github.com/golang/glog"
)

// Gadget keeps track of internal details about a gadget.
type Gadget struct {
	gadget  Circuitry
	name    string
	circuit *Circuit
	alive   bool
	inputs  map[string]*wire
	outputs map[string]*wire
}

func (w *Gadget) initGadget(wi Circuitry, nm string, gr *Circuit) *Gadget {
	if w.circuit != nil {
		glog.Fatalln("gadget is already in use:", nm)
	}
	w.gadget = wi
	w.name = nm
	w.circuit = gr
	w.inputs = map[string]*wire{}
	w.outputs = map[string]*wire{}
	return w
}

func (w *Gadget) gadgetValue() reflect.Value {
	return reflect.ValueOf(w.gadget).Elem()
}

func (w *Gadget) pinValue(pin string) reflect.Value {
	pp := pinPart(pin)
	// if it's a circuit, look up mapped pins
	if g, ok := w.gadget.(*Circuit); ok {
		p := g.pinMap[pp]
		return g.gadgetOf(p).pinValue(p) // recursive
	}
	fv := w.gadgetValue().FieldByName(pp)
	if !fv.IsValid() {
		glog.Fatalln("pin not found:", pin)
	}
	return fv
}

func (w *Gadget) getInput(pin string, capacity int) *wire {
	c := w.inputs[pin]
	if c == nil {
		c = &wire{channel: make(chan Message, capacity), dest: w}
		w.inputs[pin] = c
	}
	if capacity > c.capacity {
		c.capacity = capacity
	}
	return c
}

func (w *Gadget) setOutput(pin string, c *wire) {
	ppfv := strings.Split(pin, ":")
	fp := w.pinValue(ppfv[0])
	if len(ppfv) == 1 {
		if !fp.IsNil() {
			glog.Fatalf("output already connected: %s.%s", w.name, pin)
		}
		setValue(fp, c)
	} else { // it's not an Output, so it must be a map[string]Output
		if fp.IsNil() {
			setValue(fp, map[string]Output{})
		}
		outputs := fp.Interface().(map[string]Output)
		if _, ok := outputs[ppfv[1]]; ok {
			glog.Fatalf("output already connected: %s.%s", w.name, pin)
		}
		outputs[ppfv[1]] = c
	}
	c.senders++
	w.outputs[pin] = c
}

func (w *Gadget) setupChannels() {
	// make sure all the inbox wires have also been set up
	for dest, messages := range w.circuit.inbox {
		if gadgetPart(dest) == w.name {
			w.getInput(dest, len(messages)) // will add wire to input map
		}
	}

	// set up and pre-fill all the input pins
	for p, c := range w.inputs {
		// create a channel with the proper capacity
		c.channel = make(chan Message, c.capacity)
		setValue(w.pinValue(p), c.channel)
		// fill it with messages from the inbox, if any
		for _, m := range w.circuit.inbox[p] {
			c.channel <- m
		}
		// close the channel if there is no other feed
		if c.senders == 0 {
			close(c.channel)
		}
	}

	// set dangling inputs to a null input and dangling outputs to a fake sink
	sink := &fakeSink{}
	null := make(chan Message)
	close(null)

	we := w.gadgetValue()
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

func (w *Gadget) isFinished() bool {
	for _, c := range w.inputs {
		if len(c.channel) > 0 {
			return false
		}
	}
	return true
}

func (w *Gadget) closeChannels() {
	for p, c := range w.inputs {
		c.channel = nil
		setValue(w.pinValue(p), c.channel)
	}
	for _, c := range w.outputs {
		c.Close()
	}
}

func (w *Gadget) sendTo(c *wire, v Message) {
	if !w.alive {
		w.launch()
	}

	c.channel <- v
}

func (w *Gadget) launch() {
	w.alive = true
	w.circuit.wait.Add(1)
	w.setupChannels()

	go func() {
		defer w.circuit.wait.Done()
		defer DontPanic()

		for {
			w.gadget.Run()
			if w.isFinished() {
				break
			}
		}

		w.closeChannels()
		w.alive = false
	}()
}

func setValue(val reflect.Value, any interface{}) {
	val.Set(reflect.ValueOf(any))
}
