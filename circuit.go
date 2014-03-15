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

	gadgets map[string]*Gadget
	pinMap  map[string]string
	inbox   map[string][]Message
	wait    sync.WaitGroup
}

// Add a named gadget to the circuit with a unique name.
func (g *Circuit) Add(name, gadget string) {
	fun := Registry[gadget]
	if fun == nil {
		fmt.Println("not found: ", gadget)
		return
	}
	g.AddCircuitry(name, fun())
}

// Add a gadget or circuit to the circuit with a unique name.
func (g *Circuit) AddCircuitry(name string, w Circuitry) {
	g.gadgets[name] = w.initGadget(w, name, g)
}

func (g *Circuit) gadgetOf(s string) *Gadget {
	// TODO: migth be useful for extending an existing circuit
	// if gadgetPart(s) == "" && g.pinMap[s] != "" {
	// 	s = g.pinMap[s] // unnamed gadgets can use the circuit's pin map
	// }
	w, ok := g.gadgets[gadgetPart(s)]
	if !ok {
		glog.Fatalln("gadget not found for:", s)
	}
	return w
}

// Connect an output pin with an input pin.
func (g *Circuit) Connect(from, to string, capacity int) {
	c := g.gadgetOf(to).getInput(pinPart(to), capacity)
	g.gadgetOf(from).setOutput(pinPart(from), c)
}

// Set up a message to be sent to a gadget on startup.
func (g *Circuit) Feed(pin string, v Message) {
	g.inbox[pin] = append(g.inbox[pin], v)
}

// Start up the circuit, and return when it is finished.
func (g *Circuit) Run() {
	for _, w := range g.gadgets {
		w.launch()
	}
	g.wait.Wait()
}

// Map an external pin to an internal one.
func (g *Circuit) Label(external, internal string) {
	if strings.Contains(external, ".") {
		glog.Fatalln("external pin should not include a dot:", external)
	}
	g.pinMap[external] = internal
}
