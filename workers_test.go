package flow

import (
	"testing"
	"time"
)

func ExampleGraph() {
	team := NewTeam()
	team.Add("Printer", "printer")
	team.Request("hello", "printer.In")
	team.Run()
	// Output:
	// string: hello
}

func TestTimer(t *testing.T) {
	team := NewTeam()
	team.Add("Timer", "timer1")
	team.Add("Printer", "printer")
	team.Connect("timer1.Out", "printer.In", 0)
	team.Request(100*time.Millisecond, "timer1.Rate")
	team.Run()
}

func TestDualTimer(t *testing.T) {
    t.Skip("shared channels not working yet.")
	team := NewTeam()
	team.Add("Timer", "timer1")
	team.Add("Timer", "timer2")
	team.Add("Printer", "printer")
	team.Connect("timer1.Out", "printer.In", 0)
	team.Connect("timer2.Out", "printer.In", 0)
	team.Request(100*time.Millisecond, "timer1.Rate")
	team.Request(200*time.Millisecond, "timer2.Rate")
	team.Run()
}

func TestClock(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }
	// The following test code never ends, uncomment to try it out:
	//
    // team := NewTeam()
    // team.Add("Clock", "clock1")
    // team.Add("Printer", "printer")
    // team.Connect("clock1.Out", "printer.In", 0)
    // team.Request(time.Second, "clock1.Rate")
    // team.Run()
}
