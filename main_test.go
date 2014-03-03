package main

import "github.com/jcw/flow/flow"

func ExampleLoadAll() {
	g := flow.NewGroup()
	// TODO: these crash because they don't check for closed input channels
	// g.Add("Clock", "clock")
	// g.Add("Repeater", "repeater")
	// g.Add("SerialIn", "serial-in")
	// g.Add("Timer", "timer")
	g.Add("Counter", "counter")
	g.Add("Printer", "printer")
	g.Add("Sink", "sink")
	g.Run()
	// Output:
	// Lost output: 0
}
