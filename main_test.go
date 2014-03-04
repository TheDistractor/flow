package main

import "github.com/jcw/flow/flow"

func ExampleWorker() {
	g := flow.NewGroup()
	g.Add("clock", "Clock")
	g.Add("counter", "Counter") // returns 0
	g.Add("pipe", "Pipe")
	g.Add("printer", "Printer")
	g.Add("repeater", "Repeater")
	g.Add("serial-in", "SerialIn")
	g.Add("sink", "Sink")
	g.Add("timer", "Timer")
	g.Run()
	// Output:
	// Lost output: 0
}
