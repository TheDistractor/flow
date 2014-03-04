package netflow

import (
	"testing"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func TestMqttSub(t *testing.T) {
	t.Skip("skipping mqtt subscription test, never ends.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("s", "MqttSub")
	g.Add("p", "Printer")
	g.Connect("s.Out", "p.In", 10)
	g.Set("s.Port", ":1883")
	g.Set("s.Topic", "#")
	g.Run()
}
