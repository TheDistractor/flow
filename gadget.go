package flow

import (
	"reflect"
	"strings"

	"github.com/golang/glog"
)

// Gadget keeps track of internal details about a gadget.
type Gadget struct {
	circuitry Circuitry        // pointer to self as a Circuitry object
	name      string           // name of this gadget in the circuit
	owner     *Circuit         // owning circuit
	alive     bool             // true while running
	inputs    map[string]*wire // inbound wires
	outputs   map[string]*wire // outbound wires
}

func (g *Gadget) initGadget(cy Circuitry, nm string, ow *Circuit) *Gadget {
	if g.owner != nil {
		glog.Fatalln("gadget is already in use:", nm)
	}
	g.circuitry = cy
	g.name = nm
	g.owner = ow
	g.inputs = map[string]*wire{}
	g.outputs = map[string]*wire{}
	return g
}

func (g *Gadget) gadgetValue() reflect.Value {
	return reflect.ValueOf(g.circuitry).Elem()
}

func (g *Gadget) pinValue(pin string) reflect.Value {
	pp := pinPart(pin)
	// if it's a circuit, look up mapped pins
	if g, ok := g.circuitry.(*Circuit); ok {
		p := g.labels[pp]
		return g.gadgetOf(p).circuitry.pinValue(p) // recursive
	}
	fv := g.gadgetValue().FieldByName(pp)
	if !fv.IsValid() {
		glog.Fatalln("pin not found:", pin)
	}
	return fv
}

func (g *Gadget) getInput(pin string, capacity int) *wire {
	c := g.inputs[pin]
	if c == nil {
		c = &wire{channel: make(chan Message, capacity), dest: g}
		g.inputs[pin] = c
	}
	if capacity > c.capacity {
		c.capacity = capacity
	}
	return c
}

func (g *Gadget) setOutput(pin string, c *wire) {
	ppfv := strings.Split(pin, ":")
	fp := g.circuitry.pinValue(ppfv[0])
	if len(ppfv) == 1 {
		if !fp.IsNil() {
			glog.Fatalf("output already connected: %s.%s", g.name, pin)
		}
		setValue(fp, c)
	} else { // it's not an Output, so it must be a map[string]Output
		if fp.IsNil() {
			setValue(fp, map[string]Output{})
		}
		outputs := fp.Interface().(map[string]Output)
		if _, ok := outputs[ppfv[1]]; ok {
			glog.Fatalf("output already connected: %s.%s", g.name, pin)
		}
		outputs[ppfv[1]] = c
	}
	c.senders++
	g.outputs[pin] = c
}

func (g *Gadget) setupChannels() {
	// make sure all the feed wires have also been set up
	for dest, msgs := range g.owner.feeds {
		if gadgetPart(dest) == g.name {
			g.getInput(dest, len(msgs)) // will add wire to the inputs map
		}
	}

	// set up and pre-fill all the input pins
	for pin, wire := range g.inputs {
		// create a channel with the proper capacity
		wire.channel = make(chan Message, wire.capacity)
		setValue(g.circuitry.pinValue(pin), wire.channel)
		// fill it with messages from the feed inbox, if any
		for _, msg := range g.owner.feeds[pin] {
			wire.channel <- msg
		}
		// close the channel if there is no other feed
		if wire.senders == 0 {
			close(wire.channel)
		}
	}

	// set dangling inputs to a null input and dangling outputs to a fake sink
	gadget := g.gadgetValue()
	for i := 0; i < gadget.NumField(); i++ {
		field := gadget.Field(i)
		switch field.Type().String() {
		case "flow.Input":
			if field.IsNil() {
				null := make(chan Message)
				close(null)
				setValue(field, null)
			}
		case "flow.Output":
			if field.IsNil() {
				setValue(field, &fakeSink{})
			}
		}
	}
}

func (g *Gadget) isFinished() bool {
	for _, wire := range g.inputs {
		if len(wire.channel) > 0 {
			return false
		}
	}
	return true
}

func (g *Gadget) closeChannels() {
	for pin, wire := range g.inputs {
		wire.channel = nil
		setValue(g.circuitry.pinValue(pin), wire.channel)
	}
	for _, wire := range g.outputs {
		wire.Disconnect()
	}
}

func (g *Gadget) sendTo(w *wire, v Message) {
	if !g.alive {
		g.launch()
	}

	w.channel <- v
}

func (g *Gadget) launch() {
	g.alive = true
	g.owner.wait.Add(1)
	g.setupChannels()

	go func() {
		defer g.owner.wait.Done()
		defer DontPanic()

		for {
			g.circuitry.Run()
			if g.isFinished() {
				break
			}
		}

		g.closeChannels()
		g.alive = false
	}()
}

func setValue(value reflect.Value, any interface{}) {
	value.Set(reflect.ValueOf(any))
}
