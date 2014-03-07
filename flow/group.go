package flow

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Initialise a new group.
func NewGroup() *Group {
	return &Group{
		workers: map[string]*Work{},
		portMap: map[string]string{},
	}
}

// A group is a collection of inter-connected workers.
type Group struct {
	Work
	workers map[string]*Work
	portMap map[string]string
	wait    sync.WaitGroup
}

// Add a named worker to the group with a unique name.
func (g *Group) Add(name, worker string) {
	fun := Registry[worker]
	if fun == nil {
		fmt.Println("not found: ", worker)
		return
	}
	g.AddWorker(name, fun())
}

// Add a worker or workgroup to the group with a unique name.
func (g *Group) AddWorker(name string, w Worker) {
	g.workers[name] = w.initWork(w, name, g)
}

func (g *Group) workerOf(s string) *Work {
	if n := strings.IndexRune(s, '.'); n > 0 {
		s = s[:n]
	}
	return g.workers[s]
}

func portPart(s string) string {
	n := strings.IndexRune(s, '.')
	return s[n+1:]
}

// Connect an output port with an input port.
func (g *Group) Connect(from, to string, capacity int) {
	fw := g.workerOf(from)
	fp := fw.port(portPart(from))
	if !fp.IsNil() {
		fmt.Println("output already connected:", from)
		// TODO: refcount needs to be lowered if it's a *connection
	}
	tw := g.workerOf(to)
	c := tw.inputs[portPart(to)]
	if c == nil {
		c = &connection{channel: make(chan Memo, capacity)}
		tw.inputs[portPart(to)] = c
		tp := tw.port(portPart(to))
		tp.Set(reflect.ValueOf(c.channel))
	}
	c.senders++
	cv := reflect.ValueOf(c)
	fp.Set(cv)
	fw.outputs[portPart(from)] = c
}

// Set up a memo which needs to be sent to a worker on startup.
func (g *Group) Set(port string, v Memo) {
	w := g.workerOf(port)
	w.addToInbox(portPart(port), v)
}

// Start up the group, and return when it is finished.
func (g *Group) Run() {
	for _, w := range g.workers {
		w.Launch()
	}
	g.wait.Wait()
}

// Map an external port to an internal one.
func (g *Group) Map(external, internal string) {
	if strings.Contains(external, ".") {
		panic("external port should not include a dot: " + external)
	}
	g.portMap[external] = internal
}

// Launch a dynamically added Worker
func (g *Group) Launch(name string) {
	g.workers[name].Launch()
}
