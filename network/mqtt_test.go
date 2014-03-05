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
	g.Add("sub", "MqttSub")
	g.Set("sub.Port", ":1883")
	g.Set("sub.Topic", "#")
	g.Run()
}

func TestMqttPub(t *testing.T) {
	t.Skip("skipping mqtt publish test, needs MQTT.")
	// The following test code never ends, comment out the above to try it out
	g := flow.NewGroup()
	g.Add("pub", "MqttPub")
	g.Set("pub.Port", ":1883")
	g.Set("pub.In", []string{"Hello", "world"})
	g.Run()
}
