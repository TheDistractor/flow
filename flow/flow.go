package flow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
)

// Version of this package.
var Version = "0.0.1"

// The registry is the factory for all known types of workers.
var Registry = map[string]func() Worker{}

// Memo's are the generic type sent to, between, and from workers.
type Memo interface{}

// Get the type of a memo, using reflection.
// TODO: not used
// func Type(m Memo) string {
// 	return reflect.TypeOf(m).String()
// }

// Input ports are used to receive memo's.
type Input <-chan Memo

// Output ports are used to send memo's elsewhere.
type Output interface {
	Send(v Memo)
	Close()
}

// The worker is the basic unit of processing, shuffling memo's between ports.
type Worker interface {
	Run()

	initWork(Worker, string, *Group) *Work
}

// Work keeps track of internal details about a worker.
type Work struct {
	worker  Worker
	name    string
	parent  *Group
	inbox   map[string][]Memo
	inputs  map[string]*connection
	outputs map[string]*connection
}

func (w *Work) initWork(wi Worker, nm string, gr *Group) *Work {
	if w.parent != nil {
		panic("worker is already in use: " + nm)
	}
	w.worker = wi
	w.name = nm
	w.parent = gr
	w.inbox = map[string][]Memo{}
	w.inputs = map[string]*connection{}
	w.outputs = map[string]*connection{}
	return w
}

// Return the group of this worker.
func (w *Work) MyGroup() *Group {
	return w.parent
}

// Return the name of this worker.
func (w *Work) MyName() string {
	return w.name
}

func (w *Work) port(p string) reflect.Value {
	wp := reflect.ValueOf(w.worker)
	wv := wp.Elem()
	fv := wv.FieldByName(p)
	if !fv.IsValid() {
		// maybe it's a group with mapped ports
		if g, ok := w.worker.(*Group); ok {
			if p, ok := g.portMap[p]; ok {
				fw := g.workerOf(p)
				return fw.port(portPart(p)) // recursive
			}
		}
		fmt.Println("port not found: " + p)
	}
	return fv
}

func (w *Work) addToInbox(port string, value Memo) {
	w.inbox[port] = append(w.inbox[port], value)
}

func (w *Work) processInbox() {
	for dest, memos := range w.inbox {
		c := make(chan Memo, len(memos))
		dp := w.port(dest)
		dp.Set(reflect.ValueOf(c))
		for _, m := range memos {
			c <- m
		}
		close(c)
		delete(w.inbox, dest)
	}
}

func (w *Work) forAllPorts(f func(string, reflect.Value)) {
	wv := reflect.ValueOf(w.worker)
	we := wv.Elem()
	wt := we.Type()
	for i := 0; i < we.NumField(); i++ {
		fd := wt.Field(i)
		ft := fd.Type.Name()
		switch ft {
		case "Input", "Output":
			f(ft, we.Field(i))
		}
	}
	return
}

// use a fake sink for every output port not connected to anything else
type fakeSink struct{}

func (c *fakeSink) Send(m Memo) {
	fmt.Printf("Lost %T: %v\n", m, m)
}

func (c *fakeSink) Close() {}

func (w *Work) connectChannels() {
	sink := &fakeSink{}
	null := make(chan Memo)
	close(null)

	w.forAllPorts(func(typ string, val reflect.Value) {
		if val.IsNil() {
			switch typ {
			case "Input":
				val.Set(reflect.ValueOf(null))
			case "Output":
				val.Set(reflect.ValueOf(sink))
			}
		}
	})
}

func (w *Work) closeChannels() {
	for _, c := range w.outputs {
		c.Close()
	}
	// TODO: cleanup, to allow re-running
	// w.forAllPorts(func(typ string, val reflect.Value) {
	// 	if !val.IsNil() {
	// 		val.Set(reflect.ValueOf(nil))
	// 	}
	// })
}

func (w *Work) Launch() {
	w.parent.wait.Add(1)
	go func() {
		defer w.parent.wait.Done()

		w.processInbox()
		w.connectChannels()
		w.worker.Run()
		w.closeChannels()
	}()
}

type connection struct {
	channel chan Memo
	senders int
}

// Send a memo through an output port.
func (c *connection) Send(v Memo) {
	c.channel <- v
}

func (c *connection) Close() {
	c.senders--
	if c.senders == 0 {
		close(c.channel)
	}
}

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
		// TODO: refcount needs to be lowered
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

type config struct {
	Workers     []struct{ Type, Name string }
	Connections []struct{ From, To string }
	Requests    []struct{ Data, To string }
}

// Load a group from a JSON description in a string.
func (g *Group) LoadString(s string) error {
	var conf config
	err := json.Unmarshal([]byte(s), &conf)
	if err == nil {
		for _, w := range conf.Workers {
			g.Add(w.Name, w.Type)
		}
		for _, c := range conf.Connections {
			g.Connect(c.From, c.To, 0)
		}
		for _, r := range conf.Requests {
			g.Set(r.To, r.Data)
		}
	}
	return err
}

// Load a group from a JSON description in a file.
func (g *Group) LoadFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = g.LoadString(string(data))
	}
	return err
}

// A transformer processes each memo through a supplied function.
func Transformer(f func(Memo) Memo) Worker {
	return &transformer{fun: f}
}

type transformer struct {
	Work
	In  Input
	Out Output

	fun func(Memo) Memo
}

func (w *transformer) Run() {
	for m := range w.In {
		w.Out.Send(w.fun(m))
	}
}
