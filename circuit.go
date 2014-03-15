package flow

import (
	"fmt"
	"strings"
	"sync"

	"github.com/golang/glog"
)

// Initialise a new circuit.
func NewCircuit() *Circuit {
	return &Circuit{
		gadgets: map[string]*Gadget{},
		pinMap:  map[string]string{},
		inbox:   map[string][]Message{},
	}
}

// A circuit is a collection of inter-connected gadgets.
type Circuit struct {
	Gadget

	gadgets map[string]*Gadget   // gadgets added to this circuit
	pinMap  map[string]string    // pin label lookup map
	inbox   map[string][]Message // message feeds
	wait    sync.WaitGroup       // tracks number of running gadgets
}

// Add a named gadget to the circuit with a unique name.
func (c *Circuit) Add(name, gadget string) {
	fun := Registry[gadget]
	if fun == nil {
		fmt.Println("not found: ", gadget)
		return
	}
	c.AddCircuitry(name, fun())
}

// Add a gadget or circuit to the circuit with a unique name.
func (c *Circuit) AddCircuitry(name string, g Circuitry) {
	c.gadgets[name] = g.initGadget(g, name, c)
}

func (c *Circuit) gadgetOf(s string) *Gadget {
	// TODO: migth be useful for extending an existing circuit
	// if gadgetPart(s) == "" && c.pinMap[s] != "" {
	// 	s = c.pinMap[s] // unnamed gadgets can use the circuit's pin map
	// }
	g, ok := c.gadgets[gadgetPart(s)]
	if !ok {
		glog.Fatalln("gadget not found for:", s)
	}
	return g
}

// Connect an output pin with an input pin.
func (c *Circuit) Connect(from, to string, capacity int) {
	w := c.gadgetOf(to).getInput(pinPart(to), capacity)
	c.gadgetOf(from).setOutput(pinPart(from), w)
}

// Set up a message to feed to a gadget on startup.
func (c *Circuit) Feed(pin string, m Message) {
	c.inbox[pin] = append(c.inbox[pin], m)
}

// Start up the circuit, and return when it is finished.
func (c *Circuit) Run() {
	for _, g := range c.gadgets {
		g.launch()
	}
	c.wait.Wait()
}

// Label an external pin to map it to an internal one.
func (c *Circuit) Label(external, internal string) {
	if strings.Contains(external, ".") {
		glog.Fatalln("external pin should not include a dot:", external)
	}
	c.pinMap[external] = internal
}
