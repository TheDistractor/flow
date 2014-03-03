package workers

import (
	"fmt"
	"time"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Sink"] = func() flow.Worker { return &Sink{} }
	flow.Registry["Pipe"] = func() flow.Worker { return &Pipe{} }
	flow.Registry["Repeater"] = func() flow.Worker { return &Repeater{} }
	flow.Registry["Counter"] = func() flow.Worker { return &Counter{} }
	flow.Registry["Printer"] = func() flow.Worker { return &Printer{} }
	flow.Registry["Timer"] = func() flow.Worker { return &Timer{} }
	flow.Registry["Clock"] = func() flow.Worker { return &Clock{} }
}

// A sink eats up all the memos it receives.
type Sink struct {
	flow.Worker
	In flow.Input
}

// Start reading memos and discard them.
func (w *Sink) Run() {
	for _ = range w.In {
	}
}

// Pipes are workers with an "In" and an "Out" port.
type Pipe struct {
	flow.Worker
	In  flow.Input
	Out flow.Output
}

// Start passing through memos.
func (w *Pipe) Run() {
	for m := range w.In {
		w.Out <- m
	}
}

// Repeaters are pipes which repeat each memo a number of times.
type Repeater struct {
	flow.Worker
	In  flow.Input
	Out flow.Output
	Num flow.Input
}

// Start repeating incoming memos.
func (w *Repeater) Run() {
	num := <-w.Num
	n := num.Val.(int)
	for m := range w.In {
		for i := 0; i < n; i++ {
			w.Out <- m
		}
	}
}

// A counter reports the number of memos it has received.
type Counter struct {
	flow.Worker
	In    flow.Input
	Out   flow.Output
	count int
}

// Start counting incoming memos.
func (w *Counter) Run() {
	for _ = range w.In {
		w.count++
	}
	w.Out <- flow.NewMemo(w.count)
}

// Printers report the memos sent to them as output.
type Printer struct {
	flow.Worker
	In flow.Input
}

// Start printing incoming memos.
func (w *Printer) Run() {
	for m := range w.In {
		fmt.Printf("%s: %v\n", m.Type(), m.Val)
	}
}

// A timer sends out one memo after the time set by the Rate port.
type Timer struct {
	flow.Worker
	Rate flow.Input
	Out  flow.Output
}

// Start the timer, sends one memo when it expires.
func (w *Timer) Run() {
	rate := <-w.Rate
	t := <-time.After(rate.Val.(time.Duration))
	w.Out <- flow.NewMemo(t)
}

// A clock sends out memos at a fixed rate, as set by the Rate port.
type Clock struct {
	flow.Worker
	Rate flow.Input
	Out  flow.Output
}

// Start sending out periodic memos, once the rate is known.
func (w *Clock) Run() {
	rate := <-w.Rate
	t := time.NewTicker(rate.Val.(time.Duration))
	for m := range t.C {
		w.Out <- flow.NewMemo(m)
	}
}
