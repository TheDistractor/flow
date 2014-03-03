package flow

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// The registry is the factory for all known types of workers.
var Registry = make(map[string]func() Worker)

// Memo's are the basic type sent to, between, and from workers.
type Memo struct {
	Val  interface{}
	Attr map[string]interface{}
}

// Create a new memo from an arbitrary value and register its type.
func NewMemo(v interface{}) *Memo {
	return &Memo{v, make(map[string]interface{})}
}

// Get the type of the value of a memo, using reflection.
func (m *Memo) Type() string {
	return reflect.TypeOf(m.Val).String()
}

// Requests are memo's which need to be sent to a worker on startup.
func (g *Group) Request(v interface{}, dest string) {
	g.inbox[dest] = NewMemo(v)
}

// Input ports can receive memo's.
type Input <-chan *Memo

// Output ports are used to send memo's elsewhere.
type Output chan<- *Memo

// The worker is the basic unit of processing, shuffling memo's between ports.
type Worker interface {
	Run()
}

type connection struct {
	channel chan *Memo
	senders int
}

// A group is a collection of inter-connected workers.
type Group struct {
	inbox   map[string]*Memo
	workers map[string]Worker
	inputs  map[string]*connection
	outputs map[string]*connection
}

// Initialise a new group.
func NewGroup() *Group {
	return &Group{
		inbox:   make(map[string]*Memo),
		workers: make(map[string]Worker),
		inputs:  make(map[string]*connection),
		outputs: make(map[string]*connection),
	}
}

// Add a worker to the group, with a unique name.
func (g *Group) Add(component, name string) {
	fun := Registry[component]
	if fun == nil {
		fmt.Println("not found: ", component)
	}
	g.workers[name] = fun()
}

func (g *Group) findPort(name string) reflect.Value {
	segments := strings.Split(name, ".")
	worker := g.workers[segments[0]]
	wp := reflect.ValueOf(worker)
	wv := wp.Elem()
	fv := wv.FieldByName(segments[1])
	if !fv.IsValid() {
		fmt.Println("port not found: " + name)
	}
	return fv
}

// Connect an output port with an input port.
func (g *Group) Connect(from, to string, capacity int) {
	fp := g.findPort(from)
	if !fp.IsNil() {
		fmt.Println("from port already set: ", from)
	}
	w := g.inputs[to]
	if w == nil {
		w = &connection{channel: make(chan *Memo, capacity)}
		g.inputs[to] = w
		tp := g.findPort(to)
		tp.Set(reflect.ValueOf(w.channel))
	}
	w.senders++
	cv := reflect.ValueOf(w.channel)
	fp.Set(cv)
	g.outputs[from] = w
}

func (g *Group) pushMemo(m *Memo, dest string) {
	dp := g.findPort(dest)
	c := make(chan *Memo, 1)
	dp.Set(reflect.ValueOf(c))
	c <- m
	close(c)
}

func forAllChannels(w Worker, f func(string, reflect.Value)) {
	wv := reflect.ValueOf(w)
	we := wv.Elem()
	wt := we.Type()
	for i := 0; i < we.NumField(); i++ {
		if fd := wt.Field(i); fd.Name != "" && fd.Type.Kind() == reflect.Chan {
			f(fd.Name, we.Field(i))
		}
	}
	return
}

// Start up the group, and return when it is finished.
func (g *Group) Run() {
	done := make(chan struct{})
	sink := make(chan *Memo)
	null := make(chan *Memo)
	close(null)

	// report all memo's sent to the sink, for debugging
	go func() {
		for m := range sink {
			fmt.Println("Lost output:", m.Val)
		}
		close(done)
	}()

	var wait sync.WaitGroup
	wait.Add(len(g.workers))

	for n, w := range g.workers {
		go func(n string, w Worker) {
			// fmt.Println(" go start", n)

			// connect unused inputs to "null" and unused outputs to "sink"
			forAllChannels(w, func(_ string, v reflect.Value) {
				if v.IsNil() {
					c := null
					if v.Type().ChanDir()&reflect.SendDir != 0 {
						c = sink
					}
					v.Set(reflect.ValueOf(c))
				}
			})

			// fmt.Println("run start", n)
			w.Run()
			// fmt.Println("run end", n)

			// close all output channels once last reference is gone
			for k, v := range g.outputs {
				if strings.HasPrefix(k, n+".") {
					v.senders--
					if v.senders == 0 {
						close(v.channel)
					}
				}
			}

			wait.Done()
			// fmt.Println(" go end", n)
		}(n, w)
	}

	// send out the initial memo's
	for k, v := range g.inbox {
		g.pushMemo(v, k)
	}

	// wait until all workers have finished, as well as the sink reporter
	wait.Wait()
	close(sink)
	<-done
}
