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
		inbox:   map[string]map[string][]Memo{},
	}
}

// A group is a collection of inter-connected workers.
type Group struct {
	Work
	workers map[string]*Work
	portMap map[string]string
	inbox   map[string]map[string][]Memo
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
	return g.workers[workerPart(s)]
}

func workerPart(s string) string {
	n := strings.IndexRune(s, '.')
	if n < 0 {
		n = len(s)
	}
	return s[:n]
}

func portPart(s string) string {
	n := strings.IndexRune(s, '.')
	return s[n+1:]
}

// Connect an output port with an input port.
func (g *Group) Connect(from, to string, capacity int) {
	tw := g.workerOf(to)
	c := tw.inputs[portPart(to)]
	if c == nil {
		c = &connection{channel: make(chan Memo, capacity)}
		tw.inputs[portPart(to)] = c
		tp := tw.port(portPart(to))
		if !tp.IsValid() {
			panic("cannot set to:" + to)
		}
		tp.Set(reflect.ValueOf(c.channel))
	}
	c.senders++

	fw := g.workerOf(from)
	ppfv := strings.Split(portPart(from), ":")
	fp := fw.port(ppfv[0])
	if len(ppfv) == 1 {
		if !fp.IsNil() {
			fmt.Println("output already connected:", from)
			// TODO: close the previous Output
		}
		cv := reflect.ValueOf(c)
		fp.Set(cv)
	} else { // it's not an Output, so it must be a map[string]Output
		if fp.IsNil() {
			fp.Set(reflect.ValueOf(map[string]Output{}))
		}
		// TODO: close the previous Output, if any
		fp.Interface().(map[string]Output)[ppfv[1]] = c
	}
	fw.outputs[portPart(from)] = c
}

// Set up a memo to be sent to a worker on startup.
func (g *Group) Set(port string, v Memo) {
	wp := workerPart(port)
	if _, ok := g.inbox[wp]; !ok {
		g.inbox[wp] = map[string][]Memo{}
	}
	pp := portPart(port)
	g.inbox[wp][pp] = append(g.inbox[wp][pp], v)
}

// Start up the group, and return when it is finished.
func (g *Group) Run() {
	for _, w := range g.workers {
		w.launch()
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
func (g *Group) launch(name string) {
	g.workers[name].launch()
}
