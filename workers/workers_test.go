package workers

import (
	"testing"
	"time"

	"github.com/jcw/flow/flow"
)

func ExamplePrinter() {
	g := flow.NewGroup()
	g.Add("Printer", "p")
	g.Request("hello", "p.In")
	g.Run()
	// Output:
	// string: hello
}

func ExampleRepeater() {
	g := flow.NewGroup()
	g.Add("Repeater", "r")
	g.Add("Printer", "p")
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
	g.Add("Counter", "c")
	g.Add("Printer", "p")
	g.Connect("c.Out", "p.In", 0)
	g.Request(nil, "c.In")
	g.Run()
	// Output:
	// int: 1
}

func ExampleTimer() {
	g := flow.NewGroup()
	g.Add("Timer", "t1")
	g.Add("Timer", "t2")
	g.Add("Counter", "c")
	g.Add("Printer", "p")
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
	g.Add("Clock", "clock")
	g.Add("Counter", "counter") // returns 0
	g.Add("Pipe", "pipe")
	g.Add("Printer", "printer")
	g.Add("Repeater", "repeater")
	g.Add("Sink", "sink")
	g.Add("Timer", "timer")
	g.Run()
	// Output:
	// Lost output: 0
}

func TestTimer(t *testing.T) {
	g := flow.NewGroup()
	g.Add("Timer", "t")
	g.Add("Printer", "p")
	g.Connect("t.Out", "p.In", 0)
	g.Request(100*time.Millisecond, "t.Rate")
	g.Run()
}

func TestClock(t *testing.T) {
	t.Skip("skipping clock test, never ends.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("Clock", "c")
	g.Add("Printer", "p")
	g.Connect("c.Out", "p.In", 0)
	g.Request(time.Second, "c.Rate")
	g.Run()
}
