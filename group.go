package flow

import (
	"fmt"
	"strings"
	"sync"

	"github.com/golang/glog"
)

// Initialise a new group.
func NewGroup() *Group {
	return &Group{
		workers: map[string]*Work{},
		portMap: map[string]string{},
		inbox:   map[string][]Memo{},
	}
}

// A group is a collection of inter-connected workers.
type Group struct {
	Work

	workers map[string]*Work
	portMap map[string]string
	inbox   map[string][]Memo
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
	// TODO: migth be useful for extending an existing group
	// if workerPart(s) == "" && g.portMap[s] != "" {
	// 	s = g.portMap[s] // unnamed workers can use the group's port map
	// }
	w, ok := g.workers[workerPart(s)]
	if !ok {
		glog.Fatalln("worker not found for:", s)
	}
	return w
}

// Connect an output port with an input port.
func (g *Group) Connect(from, to string, capacity int) {
	c := g.workerOf(to).getInput(portPart(to), capacity)
	g.workerOf(from).setOutput(portPart(from), c)
}

// Set up a memo to be sent to a worker on startup.
func (g *Group) Set(port string, v Memo) {
	g.inbox[port] = append(g.inbox[port], v)
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
		glog.Fatalln("external port should not include a dot:", external)
	}
	g.portMap[external] = internal
}
