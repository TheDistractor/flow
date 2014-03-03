package flow

import (
	"testing"
	"time"
)

func ExampleGraph() {
	g := NewGroup()
	g.Add("Printer", "printer")
	g.Request("hello", "printer.In")
	g.Run()
	// Output:
	// string: hello
}

func ExampleRepeatCount() {
	g := NewGroup()
	g.Add("Repeater", "rep1")
	g.Add("Counter", "cnt1")
	g.Add("Printer", "printer")
	g.Connect("rep1.Out", "cnt1.In", 0)
	g.Connect("cnt1.Out", "printer.In", 0)
	g.Request(3, "rep1.Num")
	g.Request(nil, "rep1.In")
	g.Run()
	// Output:
	// int: 3
}

func ExampleNoCount() {
	g := NewGroup()
	g.Add("Counter", "cnt1")
	g.Add("Printer", "printer")
	g.Connect("cnt1.Out", "printer.In", 0)
	// cnt1.In is not connected, acts same as if closed right away
	g.Run()
	// Output:
	// int: 0
}

func TestTimer(t *testing.T) {
	g := NewGroup()
	g.Add("Timer", "timer1")
	g.Add("Printer", "printer")
	g.Connect("timer1.Out", "printer.In", 0)
	g.Request(100*time.Millisecond, "timer1.Rate")
	g.Run()
}

func TestDualTimer(t *testing.T) {
	g := NewGroup()
	g.Add("Timer", "timer1")
	g.Add("Timer", "timer2")
	g.Add("Printer", "printer")
	g.Connect("timer1.Out", "printer.In", 0)
	g.Connect("timer2.Out", "printer.In", 0)
	g.Request(100*time.Millisecond, "timer1.Rate")
	g.Request(200*time.Millisecond, "timer2.Rate")
	g.Run()
}

func TestClock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// The following test code never ends, uncomment to try it out:
	//
	// g := NewGroup()
	// g.Add("Clock", "clock1")
	// g.Add("Printer", "printer")
	// g.Connect("clock1.Out", "printer.In", 0)
	// g.Request(time.Second, "clock1.Rate")
	// g.Run()
}
