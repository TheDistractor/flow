package flow

import (
	"fmt"
	"time"
)

func init() {
	Registry["Sink"] = func() Worker { return &Sink{} }
	Registry["Pipe"] = func() Worker { return &Pipe{} }
	Registry["Repeater"] = func() Worker { return &Repeater{} }
	Registry["Counter"] = func() Worker { return &Counter{} }
	Registry["Printer"] = func() Worker { return &Printer{} }
	Registry["Timer"] = func() Worker { return &Timer{} }
	Registry["Clock"] = func() Worker { return &Clock{} }
}

// A sink eats up all the memos it receives.
type Sink struct {
	Worker
	In Input
}

// Start reading memos and discard them.
func (w *Sink) Run() {
	for _ = range w.In {
	}
}

// Pipes are workers with an "In" and an "Out" port.
type Pipe struct {
	Worker
	In  Input
	Out Output
}

// Start passing through memos.
func (w *Pipe) Run() {
	for m := range w.In {
		w.Out <- m
	}
}

// Repeaters are pipes which repeat each memo a number of times.
type Repeater struct {
	Worker
	In    Input
	Out   Output
	Num Input
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
	Worker
	In    Input
	Out   Output
	count int
}

// Start counting incoming memos.
func (w *Counter) Run() {
	for _ = range w.In {
		w.count++
	}
	w.Out <- NewMemo(w.count)
}

// Printers report the memos sent to them as output.
type Printer struct {
	Worker
	In Input
}

// Start printing incoming memos.
func (w *Printer) Run() {
	for m := range w.In {
		fmt.Printf("%s: %v\n", m.Type(), m.Val)
	}
}

// A timer sends out one memo after the time set by the Rate port.
type Timer struct {
	Worker
	Rate Input
	Out  Output
}

// Start the timer, sends one memo when it expires.
func (w *Timer) Run() {
	rate := <-w.Rate
	t := <-time.After(rate.Val.(time.Duration))
	w.Out <- NewMemo(t)
}

// A clock sends out memos at a fixed rate, as set by the Rate port.
type Clock struct {
	Worker
	Rate Input
	Out  Output
}

// Start sending out periodic memos, once the rate is known.
func (w *Clock) Run() {
	rate := <-w.Rate
	t := time.NewTicker(rate.Val.(time.Duration))
	for m := range t.C {
		w.Out <- NewMemo(m)
	}
}
