package flow

import (
	"testing"
)

func TestSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// The following test code never ends, uncomment to try it out:
	//
	// g := NewGroup()
	// g.Add("SerialIn", "jeelink1")
	// g.Add("SketchType", "jeelink2")
	// g.Add("Printer", "printer")
	// g.Connect("jeelink1.Out", "jeelink2.In", 0)
	// g.Connect("jeelink2.Out", "printer.In", 0)
	// g.Request("/dev/tty.usbserial-A900ad5m", "jeelink1.Port")
	// g.Run()
}
