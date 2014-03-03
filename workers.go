package flow

import (
	"fmt"
	"time"
)

func init() {
	Registry["Sink"] = func() Worker { return new(Sink) }
	Registry["Pipe"] = func() Worker { return new(Pipe) }
	Registry["Repeater"] = func() Worker { return new(Repeater) }
	Registry["Printer"] = func() Worker { return new(Printer) }
	Registry["Timer"] = func() Worker { return new(Timer) }
	Registry["Clock"] = func() Worker { return new(Clock) }
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

// Start passing through messages.
func (w *Pipe) Run() {
	for m := range w.In {
		w.Out <- m
	}
}

// Repeaters are pipes which repeat each message a number of times
type Repeater struct {
	Pipe
	Num Input
}

// Start repeating incoming messages.
func (w *Repeater) Run() {
	num := <-w.Num
	n := num.Val.(int)
	for m := range w.In {
		for i := 0; i < n; i++ {
			w.Out <- m
		}
	}
}

// Printers report the messages sent to them as output.
type Printer struct {
	Worker
	In Input
}

// Start printing incoming messages.
func (w *Printer) Run() {
	for m := range w.In {
		fmt.Printf("%s: %v\n", m.Type(), m.Val)
	}
}

// A timer send out one message after the time set by the Rate port.
type Timer struct {
	Worker
	Rate Input
	Out  Output
}

// Start the timer, sends one message when it expires.
func (w *Timer) Run() {
	rate := <-w.Rate
	t := <-time.After(rate.Val.(time.Duration))
	w.Out <- NewMemo(t)
}

// A clock sends out messages at a fixed rate, as set by the Rate port.
type Clock struct {
	Worker
	Rate Input
	Out  Output
}

// Start sending out periodic messages, once the rate is known.
func (w *Clock) Run() {
	rate := <-w.Rate
	t := time.NewTicker(rate.Val.(time.Duration))
	for m := range t.C {
		w.Out <- NewMemo(m)
	}
}
