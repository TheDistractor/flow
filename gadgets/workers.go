// Basic collection of pre-defined gadgets.
package gadgets

import (
	"bufio"
	"fmt"
	"os"
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
	flow.Registry["Forever"] = func() flow.Circuitry { return &Forever{} }
	flow.Registry["Delay"] = func() flow.Circuitry { return &Delay{} }
	flow.Registry["TimeStamp"] = func() flow.Circuitry { return &TimeStamp{} }
	flow.Registry["ReadTextFile"] = func() flow.Circuitry { return &ReadTextFile{} }
	flow.Registry["EnvVar"] = func() flow.Circuitry { return &EnvVar{} }
}

// A sink eats up all the messages it receives. Registers as "Sink".
type Sink struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start reading messages and discard them.
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

// Start passing through messages.
func (w *Pipe) Run() {
	for m := range w.In {
		w.Out.Send(m)
	}
}

// Repeaters are pipes which repeat each message a number of times.
// Registers as "Repeater".
type Repeater struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
	Num flow.Input
}

// Start repeating incoming messages.
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

// A counter reports the number of messages it has received.
// Registers as "Counter".
type Counter struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output

	count int
}

// Start counting incoming messages.
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

// Printers report the messages sent to them as output. Registers as "Printer".
type Printer struct {
	flow.Gadget
	In flow.Input
}

// Start printing incoming messages.
func (w *Printer) Run() {
	for m := range w.In {
		fmt.Printf("%+v\n", m)
	}
}

// A timer sends out one message after the time set by the Rate pin.
// Registers as "Timer".
type Timer struct {
	flow.Gadget
	Rate flow.Input
	Out  flow.Output
}

// Start the timer, sends one message when it expires.
func (w *Timer) Run() {
	if rate, ok := <-w.Rate; ok {
		t := <-time.After(rate.(time.Duration))
		w.Out.Send(t)
	}
}

// A clock sends out messages at a fixed rate, as set by the Rate pin.
// Registers as "Clock".
type Clock struct {
	flow.Gadget
	Rate flow.Input
	Out  flow.Output
}

// Start sending out periodic messages, once the rate is known.
func (w *Clock) Run() {
	if r, ok := <-w.Rate; ok {
		rate, err := time.ParseDuration(r.(string))
		flow.Check(err)
		t := time.NewTicker(rate)
		defer t.Stop()
		for m := range t.C {
			w.Out.Send(m)
		}
	}
}

// A fanout sends out messages to each of its outputs, which is set up as map.
// Registers as "FanOut".
type FanOut struct {
	flow.Gadget
	In  flow.Input
	Out map[string]flow.Output
}

// Start sending out messages to all output pins (does not make copies of them).
func (w *FanOut) Run() {
	for m := range w.In {
		for _, o := range w.Out {
			o.Send(m)
		}
	}
}

// Forever does just what the name says: run forever (and do nothing at all)
type Forever struct {
	flow.Gadget
}

// Start running (nearly) forever.
func (w *Forever) Run() {
	time.Sleep(1e6 * time.Hour)
}

// Send data out after a certain delay.
type Delay struct {
	flow.Gadget
	In    flow.Input
	Delay flow.Input
	Out   flow.Output
}

func (g *Delay) Run() {
	delay, _ := time.ParseDuration((<-g.Delay).(string))
	for m := range g.In {
		time.Sleep(delay)
		g.Out.Send(m)
	}
}

// Insert a timestamp before each message. Registers as "TimeStamp".
type TimeStamp struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start inserting timestamps.
func (w *TimeStamp) Run() {
	for m := range w.In {
		w.Out.Send(time.Now())
		w.Out.Send(m)
	}
}

// ReadTextFile takes strings and replaces them by the lines of that file.
// Inserts <open> and <close> tags before doing so. Registers as "ReadTextFile".
type ReadTextFile struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start picking up strings and injecting the text lines instead.
func (w *ReadTextFile) Run() {
	for m := range w.In {
		if name, ok := m.(string); ok {
			file, err := os.Open(name)
			flow.Check(err)
			scanner := bufio.NewScanner(file)
			w.Out.Send(flow.Tag{"<open>", name})
			for scanner.Scan() {
				w.Out.Send(scanner.Text())
			}
			w.Out.Send(flow.Tag{"<close>", name})
		} else {
			w.Out.Send(m)
		}
	}
}

// Lookup an environment variable, with optional default.
type EnvVar struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start lookup up environment variables.
func (g *EnvVar) Run() {
	for m := range g.In {
		switch v := m.(type) {
		case string:
			m = os.Getenv(v)
		case flow.Tag:
			if s := os.Getenv(v.Tag); s != "" {
				m = s
			} else {
				m = v.Msg
			}
		}
		g.Out.Send(m)
	}
}
