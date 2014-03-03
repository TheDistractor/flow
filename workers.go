package flow

import (
	"fmt"
	"time"
)

func init() {
	Registry["Sink"] = func() Worker { return new(Sink) }
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

// A pass-through simply copies each message from its In to Out ports.
type PassThrough Pipe

// Start passing through messages.
func (w *PassThrough) Run() {
	for m := range w.In {
		w.Out <- m
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
		fmt.Printf("%s: %v\n", m.Type, m.Val)
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
