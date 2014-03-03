package flow

import (
	"testing"
)

func TestSerial(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }
	// team := NewTeam()
	// team.Add("SerialIn", "jeelink1")
	// team.Add("SketchType", "jeelink2")
	// team.Add("Printer", "printer")
	// team.Connect("jeelink1.Out", "jeelink2.In", 0)
	// team.Connect("jeelink2.Out", "printer.In", 0)
	// team.Request("/dev/tty.usbserial-A900ad5m", "jeelink1.Port")
	// team.Run()
}
