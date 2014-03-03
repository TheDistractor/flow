package flow

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Version of this package.
var Version = "0.0.1"

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

// Input ports can receive memo's.
type Input <-chan *Memo

// Output ports are used to send memo's elsewhere.
type Output chan<- *Memo

// The worker is the basic unit of processing, shuffling memo's between ports.
type Worker interface {
	Run()

	initWork(Worker, string, *Group) *Work
}

// Work keeps some information about each worker.
type Work struct {
	worker  Worker
	name    string
	parent  *Group
	inbox   map[string]*Memo
	inputs  map[string]*connection
	outputs map[string]*connection
}

func (w *Work) initWork(wi Worker, nm string, gr *Group) *Work {
	w.worker = wi
	w.name = nm
	w.parent = gr
	w.inbox = make(map[string]*Memo)
	w.inputs = make(map[string]*connection)
	w.outputs = make(map[string]*connection)
	return w
}

func (w *Work) port(name string) reflect.Value {
	wp := reflect.ValueOf(w.worker)
	wv := wp.Elem()
	fv := wv.FieldByName(portPart(name))
	if !fv.IsValid() {
		fmt.Println("port not found: " + name)
	}
	return fv
}

type connection struct {
	channel chan *Memo
	senders int
}

// A group is a collection of inter-connected workers.
type Group struct {
	workers map[string]*Work
}

// Initialise a new group.
func NewGroup() *Group {
	return &Group{make(map[string]*Work)}
}

// Add a worker to the group, with a unique name.
func (g *Group) Add(worker, name string) {
	fun := Registry[worker]
	if fun == nil {
		fmt.Println("not found: ", worker)
	}
	w := fun()
	g.workers[name] = w.initWork(w, name, g)
}

func (g *Group) workerOf(s string) *Work {
	n := strings.IndexRune(s, '.')
	return g.workers[s[:n]]
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
		fmt.Println("from port already set: ", from)
	}
	tw := g.workerOf(to)
	c := tw.inputs[portPart(to)]
	if c == nil {
		c = &connection{channel: make(chan *Memo, capacity)}
		tw.inputs[portPart(to)] = c
		tp := tw.port(to)
		tp.Set(reflect.ValueOf(c.channel))
	}
	c.senders++
	cv := reflect.ValueOf(c.channel)
	fp.Set(cv)
	fw.outputs[portPart(from)] = c
}

// Requests are memo's which need to be sent to a worker on startup.
func (g *Group) Request(v interface{}, dest string) {
	w := g.workerOf(dest)
	w.inbox[portPart(dest)] = NewMemo(v)
}

func forAllChannels(w *Work, f func(string, reflect.Value)) {
	wv := reflect.ValueOf(w.worker)
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
		go func(n string, w *Work) {
			// fmt.Println(" go start", n)

			// send out the initial memo's
			for dest, memo := range w.inbox {
				dp := w.port(dest)
				c := make(chan *Memo, 1)
				dp.Set(reflect.ValueOf(c))
				c <- memo
				close(c)
			}

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
			w.worker.Run()
			// fmt.Println("run end", n)

			// close all output channels once last reference is gone
			for _, v := range w.outputs {
				v.senders--
				if v.senders == 0 {
					close(v.channel)
				}
			}

			wait.Done()
			// fmt.Println(" go end", n)
		}(n, w)
	}

	// wait until all workers have finished, as well as the sink reporter
	wait.Wait()
	close(sink)
	<-done
}
