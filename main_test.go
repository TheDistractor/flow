package main

import "github.com/jcw/flow/flow"

func ExampleWorker() {
	g := flow.NewGroup()
	g.Add("Clock", "clock")
	g.Add("Counter", "counter") // returns 0
	g.Add("Pipe", "pipe")
	g.Add("Printer", "printer")
	g.Add("Repeater", "repeater")
	g.Add("SerialIn", "serial-in")
	g.Add("Sink", "sink")
	g.Add("Timer", "timer")
	g.Run()
	// Output:
	// Lost output: 0
}
