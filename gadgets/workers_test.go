package gadgets

import (
	"testing"
	"time"

	"github.com/jcw/flow"
)

func ExamplePrinter() {
	g := flow.NewCircuit()
	g.Add("p", "Printer")
	g.Feed("p.In", "hello")
	g.Run()
	// Output:
	// string: hello
}

func ExampleRepeater() {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Feed("r.Num", 3)
	g.Feed("r.In", "abc")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: abc
	// Lost string: abc
}

func ExampleCounter() {
	g := flow.NewCircuit()
	g.Add("c", "Counter")
	g.Feed("c.In", nil)
	g.Run()
	// Output:
	// Lost int: 1
}

func ExampleTimer() {
	g := flow.NewCircuit()
	g.Add("t1", "Timer")
	g.Add("t2", "Timer")
	g.Add("c", "Counter")
	g.Connect("t1.Out", "c.In", 0)
	g.Connect("t2.Out", "c.In", 0)
	g.Feed("t1.Rate", 100*time.Millisecond)
	g.Feed("t2.Rate", 200*time.Millisecond)
	g.Run()
	// Output:
	// Lost int: 2
}

func ExampleAllCircuitries() {
	g := flow.NewCircuit()
	g.Add("clock", "Clock")
	g.Add("counter", "Counter") // returns 0
	g.Add("pipe", "Pipe")
	g.Add("printer", "Printer")
	g.Add("repeater", "Repeater")
	g.Add("sink", "Sink")
	g.Add("timer", "Timer")
	g.Run()
	// Output:
	// Lost int: 0
}

func ExampleFanOut() {
	g := flow.NewCircuit()
	g.Add("f", "FanOut")
	g.Add("c", "Counter")
	g.Add("p", "Printer")
	g.Connect("f.Out:c", "c.In", 0)
	g.Connect("f.Out:p", "p.In", 0)
	g.Feed("f.In", "abc")
	g.Feed("f.In", "def")
	g.Run()
	// Output:
	// string: abc
	// string: def
	// Lost int: 2
}

func ExampleDelay() {
	g := flow.NewCircuit()
	g.Add("d", "Delay")
	g.Add("p", "Printer")
	g.Feed("d.Delay", "10ms")
	g.Feed("d.In", "abc")
	g.Feed("p.In", "def")
	g.Run()
	// Output:
	// string: def
	// Lost string: abc
}

func TestTimer(t *testing.T) {
	g := flow.NewCircuit()
	g.Add("t", "Timer")
	g.Feed("t.Rate", 100*time.Millisecond)
	g.Run()
}

func ExampleClock() {
	// The following example never ends.
	g := flow.NewCircuit()
	g.Add("c", "Clock")
	g.Feed("c.Rate", time.Second)
	g.Run()
}
