// Interface to MQTT as client and as server.
package network

import (
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jcw/flow/flow"
	"github.com/jeffallen/mqtt"
)

func init() {
	flow.Registry["MqttSub"] = func() flow.Worker { return &MqttSub{} }
	flow.Registry["MqttPub"] = func() flow.Worker { return &MqttPub{} }
	flow.Registry["MqttServer"] = func() flow.Worker { return &MqttServer{} }
}

// MqttSub can subscribe to MQTT. Registers as "MqttSub".
type MqttSub struct {
	flow.Work
	Port  flow.Input
	Topic flow.Input
	Out   flow.Output
}

// Start listening and subscribing to MQTT.
func (w *MqttSub) Run() {
	if port, ok := <-w.Port; ok {
		sock, err := net.Dial("tcp", port.(string))
		flow.Check(err)
		client := mqtt.NewClientConn(sock)
		err = client.Connect("", "")
		flow.Check(err)

		if topic, ok := <-w.Topic; ok {
			client.Subscribe([]proto.TopicQos{{
				Topic: topic.(string),
				Qos:   proto.QosAtMostOnce,
			}})
			for m := range client.Incoming {
				payload := []byte(m.Payload.(proto.BytesPayload))
				w.Out.Send([]string{m.TopicName, string(payload)})
			}
		}
	}
}

// MqttPub can publish to MQTT. Registers as "MqttPub".
type MqttPub struct {
	flow.Work
	Port flow.Input
	In   flow.Input
}

// Start publishing to MQTT.
func (w *MqttPub) Run() {
	if port, ok := <-w.Port; ok {
		sock, err := net.Dial("tcp", port.(string))
		flow.Check(err)
		client := mqtt.NewClientConn(sock)
		err = client.Connect("", "")
		flow.Check(err)

		if m, ok := <-w.In; ok {
			msg := m.([]string)
			client.Publish(&proto.Publish{
				Header:    proto.Header{Retain: msg[0][0] == '/'},
				TopicName: msg[0],
				Payload:   proto.BytesPayload(msg[1]),
			})
		}
	}
}

// MqttServer is an embedded MQTT server. Registers as "MqttServer".
type MqttServer struct {
	flow.Work
	Port flow.Input
}

// Start the MQTT server.
func (w *MqttServer) Run() {
	if port, ok := <-w.Port; ok {
		listener, err := net.Listen("tcp", port.(string))
		flow.Check(err)
		server := mqtt.NewServer(listener)
		server.Start()
		<-server.Done
	}
}
