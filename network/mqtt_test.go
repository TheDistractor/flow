package network

import (
	"testing"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func TestMqttSub(t *testing.T) {
	t.Skip("skipping mqtt subscribe test, needs MQTT and never ends.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("s", "MqttSub")
	g.Set("s.Port", ":1883")
	g.Set("s.Topic", "#")
	g.Run()
}

func TestMqttPub(t *testing.T) {
	t.Skip("skipping mqtt publish test, needs MQTT.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("p", "MqttPub")
	g.Set("p.Port", ":1883")
	g.Set("p.In", []string{"Hello", "world"})
	g.Run()
}
