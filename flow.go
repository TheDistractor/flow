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
	Type string
	Attr map[string]interface{}
}

// Create a new memo from an arbitrary value and register its type.
func NewMemo(v interface{}) *Memo {
	return &Memo{v, reflect.TypeOf(v).String(), make(map[string]interface{})}
}

// Requests are memo's which need to be sent to a worker on startup.
func (t *Team) Request(v interface{}, dest string) {
	m := NewMemo(v)
	m.Attr["dest"] = dest
	t.inbox = append(t.inbox, m)
}

// Input ports can receive memo's.
type Input <-chan *Memo

// Output ports are used to send memo's elsewhere.
type Output chan<- *Memo

// The worker is the basic unit of processing, shuffling memo's between ports.
type Worker interface {
	Run()
}

// Pipes are workers with an "In" and an "Out" port.
type Pipe struct {
	Worker
	In  Input
	Out Output
}

// A team is a collection of inter-connected workers.
type Team struct {
	inbox   []*Memo
	workers map[string]Worker
}

// Initialise a new team.
func NewTeam() *Team {
	return &Team{
		workers: make(map[string]Worker),
	}
}

// Add a worker to the team, with a unique name.
func (t *Team) Add(component, name string) {
	fun := Registry[component]
	if fun == nil {
		fmt.Println("not found: ", component)
	}
	t.workers[name] = fun()
}

func (t *Team) findPort(name string) reflect.Value {
	segments := strings.Split(name, ".")
	worker := t.workers[segments[0]]
	wp := reflect.ValueOf(worker)
	wv := wp.Elem()
	fv := wv.FieldByName(segments[1])
	if !fv.IsValid() {
		fmt.Println("port not found: " + name)
	}
	return fv
}

// Connect an output port with an input port.
func (t *Team) Connect(from, to string, capacity int) {
	fp := t.findPort(from)
	tp := t.findPort(to)
	if !fp.IsNil() || !tp.IsNil() {
		fmt.Println("ports already set?", fp, tp)
	}
	c := make(chan *Memo, capacity)
	fp.Set(reflect.ValueOf(c))
	tp.Set(reflect.ValueOf(c))
}

func (t *Team) pushMemo(m *Memo, dest string) {
	dp := t.findPort(dest)
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

// Start up the team, and return when it is finished.
func (t *Team) Run() {
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
	wait.Add(len(t.workers))

	for _, w := range t.workers {
		go func(w Worker) {
			channels := outputChannels(w)

			// set all unused output channels to "sink"
			for _, v := range channels {
				if v.IsNil() {
					v.Set(reflect.ValueOf(sink))
				}
			}

			w.Run()

			// close all output channels, except "sink"
			for _, v := range channels {
				if !v.IsNil() && v.Interface() != Output(sink) {
					close(v.Interface().(Output))
				}
			}

			wait.Done()
		}(w)
	}

	// send out the initial memo's
	for _, v := range t.inbox {
		t.pushMemo(v, v.Attr["dest"].(string)) // TODO: not general enough
	}

	// wait until all workers have finished, as well as the sink reporter
	wait.Wait()
	close(sink)
	<-done
}

// Generic error checking.
func Check(e error) {
	if e != nil {
		panic(e)
	}
}
