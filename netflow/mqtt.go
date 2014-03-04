package netflow

import (
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jcw/flow/flow"
	"github.com/jeffallen/mqtt"
)

func init() {
	flow.Registry["MqttSub"] = func() flow.Worker { return &MqttSub{} }
	flow.Registry["MqttPub"] = func() flow.Worker { return &MqttPub{} }
}

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
		if err != nil {
			panic(err)
		}
		client := mqtt.NewClientConn(sock)
		err = client.Connect("", "")
		if err != nil {
			panic(err)
		}

		if topic, ok := <-w.Topic; ok {
			client.Subscribe([]proto.TopicQos{{
				Topic: topic.(string),
				Qos:   proto.QosAtMostOnce,
			}})
			for m := range client.Incoming {
				payload := []byte(m.Payload.(proto.BytesPayload))
				w.Out <- []string{m.TopicName, string(payload)}
			}
		}
	}
}

type MqttPub struct {
	flow.Work
	Port flow.Input
	In   flow.Input
}

// Start publishing to MQTT.
func (w *MqttPub) Run() {
	if port, ok := <-w.Port; ok {
		sock, err := net.Dial("tcp", port.(string))
		if err != nil {
			panic(err)
		}
		client := mqtt.NewClientConn(sock)
		err = client.Connect("", "")
		if err != nil {
			panic(err)
		}

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
