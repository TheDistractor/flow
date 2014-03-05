package main

import "github.com/jcw/flow/flow"

func Example() {
	g := flow.NewGroup()
	g.Add("clock", "Clock")
	g.Add("counter", "Counter") // will return 0 when not hooked up
	g.Add("pipe", "Pipe")
	g.Add("printer", "Printer")
	g.Add("repeater", "Repeater")
	g.Add("serial-in", "SerialIn")
	g.Add("sink", "Sink")
	g.Add("timer", "Timer")
	g.Run()
	// Output:
	// Lost int: 0
}
