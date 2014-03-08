package workers

import (
	"testing"
	"time"

	"github.com/jcw/flow/flow"
)

func ExamplePrinter() {
	g := flow.NewGroup()
	g.Add("p", "Printer")
	g.Set("p.In", "hello")
	g.Run()
	// Output:
	// string: hello
}

func ExampleRepeater() {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Set("r.Num", 3)
	g.Set("r.In", "abc")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: abc
	// Lost string: abc
}

func ExampleCounter() {
	g := flow.NewGroup()
	g.Add("c", "Counter")
	g.Set("c.In", nil)
	g.Run()
	// Output:
	// Lost int: 1
}

func ExampleTimer() {
	g := flow.NewGroup()
	g.Add("t1", "Timer")
	g.Add("t2", "Timer")
	g.Add("c", "Counter")
	g.Connect("t1.Out", "c.In", 0)
	g.Connect("t2.Out", "c.In", 0)
	g.Set("t1.Rate", 100*time.Millisecond)
	g.Set("t2.Rate", 200*time.Millisecond)
	g.Run()
	// Output:
	// Lost int: 2
}

func ExampleAllWorkers() {
	g := flow.NewGroup()
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
	g := flow.NewGroup()
	g.Add("fo", "FanOut")
	g.Add("c", "Counter")
	g.Add("p", "Printer")
	g.Connect("fo.Out:c", "c.In", 0)
	g.Connect("fo.Out:p", "p.In", 0)
	g.Set("fo.In", "abc")
	g.Set("fo.In", "def")
	g.Run()
	// Output:
	// string: abc
	// string: def
	// Lost int: 2
}

func TestTimer(t *testing.T) {
	g := flow.NewGroup()
	g.Add("t", "Timer")
	g.Set("t.Rate", 100*time.Millisecond)
	g.Run()
}

func TestClock(t *testing.T) {
	t.Skip("skipping clock test, never ends.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("c", "Clock")
	g.Set("c.Rate", time.Second)
	g.Run()
}
