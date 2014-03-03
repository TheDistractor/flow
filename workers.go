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

type Sink struct {
	Worker
	In Input
}

func (w *Sink) Run() {
	for _ = range w.In {
	}
}

type Printer struct {
	Worker
	In Input
}

func (w *Printer) Run() {
	for m := range w.In {
		fmt.Printf("%s: %v\n", m.Type, m.Val)
	}
}

type Timer struct {
	Worker
	Rate Input
	Out  Output
}

func (w *Timer) Run() {
	rate := <-w.Rate
	t := <-time.After(rate.Val.(time.Duration))
	w.Out <- NewMemo(t)
}

type Clock struct {
	Worker
	Rate Input
	Out  Output
}

func (w *Clock) Run() {
	rate := <-w.Rate
	t := time.NewTicker(rate.Val.(time.Duration))
	for m := range t.C {
		w.Out <- NewMemo(m)
	}
}
