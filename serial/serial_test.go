package serial

import (
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/workers"
)

func TestSerial(t *testing.T) {
	t.Skip("skipping serial test, never ends and needs hardware.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("SerialIn", "jeelink1")
	g.Add("SketchType", "jeelink2")
	g.Add("Printer", "printer")
	g.Connect("jeelink1.Out", "jeelink2.In", 0)
	g.Connect("jeelink2.Out", "printer.In", 0)
	g.Request("/dev/tty.usbserial-A900ad5m", "jeelink1.Port")
	g.Run()
}
