// Basic collection of pre-defined gadgets.
package gadgets

import (
	"fmt"
	"time"

	"github.com/jcw/flow"
)

func init() {
	flow.Registry["Sink"] = func() flow.Circuitry { return &Sink{} }
	flow.Registry["Pipe"] = func() flow.Circuitry { return &Pipe{} }
	flow.Registry["Repeater"] = func() flow.Circuitry { return &Repeater{} }
	flow.Registry["Counter"] = func() flow.Circuitry { return &Counter{} }
	flow.Registry["Printer"] = func() flow.Circuitry { return &Printer{} }
	flow.Registry["Timer"] = func() flow.Circuitry { return &Timer{} }
	flow.Registry["Clock"] = func() flow.Circuitry { return &Clock{} }
	flow.Registry["FanOut"] = func() flow.Circuitry { return &FanOut{} }
}

// A sink eats up all the memos it receives. Registers as "Sink".
type Sink struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start reading memos and discard them.
func (w *Sink) Run() {
	w.Out.Close()
	for _ = range w.In {
	}
}

// Pipes are gadgets with an "In" and an "Out" pin. Registers as "Pipe".
type Pipe struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start passing through memos.
func (w *Pipe) Run() {
	for m := range w.In {
		w.Out.Send(m)
	}
}

// Repeaters are pipes which repeat each memo a number of times.
// Registers as "Repeater".
type Repeater struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
	Num flow.Input
}

// Start repeating incoming memos.
func (w *Repeater) Run() {
	if num, ok := <-w.Num; ok {
		n := num.(int)
		for m := range w.In {
			count := n
			if _, ok = m.(flow.Tag); ok {
				count = 1 // don't repeat tags, just pass them through
			}
			for i := 0; i < count; i++ {
				w.Out.Send(m)
			}
		}
	}
}

// A counter reports the number of memos it has received.
// Registers as "Counter".
type Counter struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output

	count int
}

// Start counting incoming memos.
func (w *Counter) Run() {
	for m := range w.In {
		if _, ok := m.(flow.Tag); ok {
			w.Out.Send(m) // don't count tags, just pass them through
		} else {
			w.count++
		}
	}
	w.Out.Send(w.count)
}

// Printers report the memos sent to them as output. Registers as "Printer".
type Printer struct {
	flow.Gadget
	In flow.Input
}

// Start printing incoming memos.
func (w *Printer) Run() {
	for m := range w.In {
		fmt.Printf("%T: %v\n", m, m)
	}
}

// A timer sends out one memo after the time set by the Rate pin.
// Registers as "Timer".
type Timer struct {
	flow.Gadget
	Rate flow.Input
	Out  flow.Output
}

// Start the timer, sends one memo when it expires.
func (w *Timer) Run() {
	if rate, ok := <-w.Rate; ok {
		t := <-time.After(rate.(time.Duration))
		w.Out.Send(t)
	}
}

// A clock sends out memos at a fixed rate, as set by the Rate pin.
// Registers as "Clock".
type Clock struct {
	flow.Gadget
	Rate flow.Input
	Out  flow.Output
}

// Start sending out periodic memos, once the rate is known.
func (w *Clock) Run() {
	if rate, ok := <-w.Rate; ok {
		t := time.NewTicker(rate.(time.Duration))
		for m := range t.C {
			w.Out.Send(m)
		}
	}
}

// A fanout sends out memos to each of its outputs, which is set up as map.
// Registers as "FanOut".
type FanOut struct {
	flow.Gadget
	In  flow.Input
	Out map[string]flow.Output
}

// Start sending out memos to all output pins (does not make copies of them).
func (w *FanOut) Run() {
	for m := range w.In {
		for _, o := range w.Out {
			o.Send(m)
		}
	}
}
