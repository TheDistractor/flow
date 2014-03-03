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
	m := NewMemo(v)
	m.Attr["dest"] = dest
	g.inbox = append(g.inbox, m)
}

// Input ports can receive memo's.
type Input <-chan *Memo

// Output ports are used to send memo's elsewhere.
type Output chan<- *Memo

// The worker is the basic unit of processing, shuffling memo's between ports.
type Worker interface {
	Run()
}

type Wire struct {
	conn chan *Memo
	refs int
	// To    string
	// Froms map[string]string
}

// A group is a collection of inter-connected workers.
type Group struct {
	inbox   []*Memo
	workers map[string]Worker
	inputs  map[string]*Wire
	outputs map[string]*Wire
}

// Initialise a new group.
func NewGroup() *Group {
	return &Group{
		workers: make(map[string]Worker),
		inputs:  make(map[string]*Wire),
		outputs: make(map[string]*Wire),
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
		w = &Wire{conn: make(chan *Memo, capacity)}
		g.inputs[to] = w
		tp := g.findPort(to)
		tp.Set(reflect.ValueOf(w.conn))
	}
	w.refs++
	cv := reflect.ValueOf(w.conn)
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

func outputChannels(w Worker) (results []reflect.Value) {
	forAllChannels(w, func(name string, value reflect.Value) {
		if value.Type().ChanDir()&reflect.SendDir != 0 {
			results = append(results, value)
		}
	})
	return
}

// Start up the group, and return when it is finished.
func (g *Group) Run() {
	done := make(chan struct{})
	sink := make(chan *Memo)

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
			channels := outputChannels(w)

			// set all unused output channels to "sink"
			for _, v := range channels {
				if v.IsNil() {
					v.Set(reflect.ValueOf(sink))
				}
			}

			// fmt.Println("start", n)
			w.Run()
			// fmt.Println("end", n)

			// close all output channels once last reference is gone
			for k, v := range g.outputs {
				if strings.HasPrefix(k, n+".") {
					v.refs--
					if v.refs == 0 {
						close(v.conn)
					}
				}
			}

			wait.Done()
		}(n, w)
	}

	// send out the initial memo's
	for _, v := range g.inbox {
		g.pushMemo(v, v.Attr["dest"].(string)) // TODO: not general enough
	}

	// wait until all workers have finished, as well as the sink reporter
	wait.Wait()
	close(sink)
	<-done
}

// Generic error checking, panics if e is not nil.
func Check(e error) {
	if e != nil {
		panic(e)
	}
}
