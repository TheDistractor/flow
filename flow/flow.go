package flow

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/golang/glog"
)

// Version of this package.
var Version = "0.1.0"

// The registry is the factory for all known types of workers.
var Registry = map[string]func() Worker{}

// Memos are the generic type sent to, between, and from workers.
type Memo interface{}

// A tag allows adding a descriptive string to a memo.
type Tag struct {
	Tag string
	Val Memo
}

// Input ports are used to receive memos.
type Input <-chan Memo

// Output ports are used to send memos elsewhere.
type Output interface {
	Send(v Memo) // Send a memo through an output port.
	Close()      // Detach the port, close channel when last one is gone.
}

// The worker is the basic unit of processing, shuffling memos between ports.
type Worker interface {
	Run()

	initWork(Worker, string, *Group) *Work
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
	if m, ok := <-w.In; ok {
		w.Out.Send(w.fun(m))
	}
}

// A connection is a ref-counted Input, it's closed when the count drops to 0.
type connection struct {
	channel  chan Memo
	senders  int
	capacity int
	dest     *Work
}

func (c *connection) Send(v Memo) {
	c.dest.launch()
	// TODO: there's still a race condition if c.dest dies here
	c.channel <- v
}

func (c *connection) Close() {
	c.senders--
	if c.senders == 0 && c.channel != nil {
		close(c.channel)
	}
}

// Use a fake sink for every output port not connected to anything else.
type fakeSink struct{}

func (c *fakeSink) Send(m Memo) {
	fmt.Printf("Lost %T: %v\n", m, m)
}

func (c *fakeSink) Close() {}

// extract "a" from "a.b", panics if there's no dot in the string
func workerPart(s string) string {
	n := strings.IndexRune(s, '.')
	return s[:n]
}

// extract "b" from "a.b", also works if only "b" is given
func portPart(s string) string {
	n := strings.IndexRune(s, '.')
	return s[n+1:]
}

// Utility to check for errors and panic if the arg is not nil.
func Check(err interface{}) {
	if err != nil {
		glog.Fatal(err)
	}
}

// Call this as "defer flow.DontPanic()" for a concise stack trace on panics.
func DontPanic() {
	// generate a nice stack trace, see https://code.google.com/p/gonicetrace/
	if e := recover(); e != nil {
		fmt.Fprintf(os.Stderr, "\nPANIC: %v\n", e)
		for skip := 1; skip < 20; skip++ {
			pc, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			if strings.HasSuffix(file, ".go") {
				name := runtime.FuncForPC(pc).Name()
				name = name[strings.LastIndex(name, "/")+1:]
				fmt.Fprintf(os.Stderr, "%s:%d %s()\n", file, line, name)
			}
		}
		glog.Fatal("EXIT")
	}
}
