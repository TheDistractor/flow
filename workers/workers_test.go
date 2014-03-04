package workers

import (
	"testing"
	"time"

	"github.com/jcw/flow/flow"
)

func ExamplePrinter() {
	g := flow.NewGroup()
	g.Add("p", "Printer")
	g.Request("hello", "p.In")
	g.Run()
	// Output:
	// string: hello
}

func ExampleRepeater() {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("p", "Printer")
	g.Connect("r.Out", "p.In", 0)
	g.Request(3, "r.Num")
	g.Request("abc", "r.In")
	g.Run()
	// Output:
	// string: abc
	// string: abc
	// string: abc
}

func ExampleCounter() {
	g := flow.NewGroup()
	g.Add("c", "Counter")
	g.Add("p", "Printer")
	g.Connect("c.Out", "p.In", 0)
	g.Request(nil, "c.In")
	g.Run()
	// Output:
	// int: 1
}

func ExampleTimer() {
	g := flow.NewGroup()
	g.Add("t1", "Timer")
	g.Add("t2", "Timer")
	g.Add("c", "Counter")
	g.Add("p", "Printer")
	g.Connect("t1.Out", "c.In", 0)
	g.Connect("t2.Out", "c.In", 0)
	g.Connect("c.Out", "p.In", 0)
	g.Request(100*time.Millisecond, "t1.Rate")
	g.Request(200*time.Millisecond, "t2.Rate")
	g.Run()
	// Output:
	// int: 2
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
	// Lost output: 0
}

func TestTimer(t *testing.T) {
	g := flow.NewGroup()
	g.Add("t", "Timer")
	g.Add("p", "Printer")
	g.Connect("t.Out", "p.In", 0)
	g.Request(100*time.Millisecond, "t.Rate")
	g.Run()
}

func TestClock(t *testing.T) {
	t.Skip("skipping clock test, never ends.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("c", "Clock")
	g.Add("p", "Printer")
	g.Connect("c.Out", "p.In", 0)
	g.Request(time.Second, "c.Rate")
	g.Run()
}
