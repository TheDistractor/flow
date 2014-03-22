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
		p := g.pinMap[pp]
		return g.gadgetOf(p).pinValue(p) // recursive
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
	fp := g.pinValue(ppfv[0])
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
	// make sure all the inbox wires have also been set up
	for dest, messages := range g.owner.inbox {
		if gadgetPart(dest) == g.name {
			g.getInput(dest, len(messages)) // will add wire to input map
		}
	}

	// set up and pre-fill all the input pins
	for p, c := range g.inputs {
		// create a channel with the proper capacity
		c.channel = make(chan Message, c.capacity)
		setValue(g.pinValue(p), c.channel)
		// fill it with messages from the inbox, if any
		for _, m := range g.owner.inbox[p] {
			c.channel <- m
		}
		// close the channel if there is no other feed
		if c.senders == 0 {
			close(c.channel)
		}
	}

	// set dangling inputs to a null input and dangling outputs to a fake sink
	we := g.gadgetValue()
	for i := 0; i < we.NumField(); i++ {
		fe := we.Field(i)
		switch fe.Type().String() {
		case "flow.Input":
			if fe.IsNil() {
				null := make(chan Message)
				close(null)
				setValue(fe, null)
			}
		case "flow.Output":
			if fe.IsNil() {
				setValue(fe, &fakeSink{})
			}
		}
	}
}

func (g *Gadget) isFinished() bool {
	for _, c := range g.inputs {
		if len(c.channel) > 0 {
			return false
		}
	}
	return true
}

func (g *Gadget) closeChannels() {
	for p, c := range g.inputs {
		c.channel = nil
		setValue(g.pinValue(p), c.channel)
	}
	for _, c := range g.outputs {
		c.Disconnect()
	}
}

func (g *Gadget) sendTo(c *wire, v Message) {
	if !g.alive {
		g.launch()
	}

	c.channel <- v
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

func setValue(val reflect.Value, any interface{}) {
	val.Set(reflect.ValueOf(any))
}
