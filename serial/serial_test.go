package serial

import (
	"testing"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func TestSerial(t *testing.T) {
	t.Skip("skipping serial test, never ends and needs hardware.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("SerialIn", "s")
	g.Add("SketchType", "t")
	g.Add("Printer", "p")
	g.Connect("s.Out", "t.In", 0)
	g.Connect("t.Out", "p.In", 0)
	g.Set("s.Port", "/dev/tty.usbserial-A900ad5m")
	g.Run()
}
