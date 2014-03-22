package gadgets

import (
	"os"
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
	// hello
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
	// abc
	// def
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
	// def
	// Lost string: abc
}

func ExampleTimeStamp() {
	g := flow.NewCircuit()
	g.Add("t", "TimeStamp")
	g.Run()
	g.Feed("t.In", "abc")
	// produces two lines, the timestamp followed by the "abc" string
}

func ExampleReadFileText() {
	g := flow.NewCircuit()
	g.Add("r", "ReadFileText")
	g.Feed("r.In", "example.json")
	g.Run()
	// Output:
	// Lost flow.Tag: {<open> example.json}
	// Lost string: {
	// Lost string:     "a": 123,
	// Lost string:     "b": [3,4,5],
	// Lost string:     "c": true
	// Lost string: }
	// Lost flow.Tag: {<close> example.json}
}

func ExampleReadFileJSON() {
	g := flow.NewCircuit()
	g.Add("r", "ReadFileJSON")
	g.Feed("r.In", "example.json")
	g.Run()
	// Output:
	// Lost flow.Tag: {<file> example.json}
	// Lost map[string]interface {}: map[a:123 b:[3 4 5] c:true]
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
	g.Feed("c.Rate", "1s")
	g.Run()
}

func ExampleEnvVar() {
	os.Setenv("FOO", "bar!")

	g := flow.NewCircuit()
	g.Add("e", "EnvVar")
	g.Feed("e.In", "FOO")
	g.Feed("e.In", flow.Tag{"FOO", "def"})
	g.Feed("e.In", flow.Tag{"BLAH", "abc"})
	g.Run()
	// Output:
	// Lost string: bar!
	// Lost string: bar!
	// Lost string: abc
}
