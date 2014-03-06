package decoders

import (
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-roomNode"] = func() flow.Worker { return &RoomNode{} }
}

// Decoder for the "roomNode.ino" sketch. Registers as "Node-roomNode".
type RoomNode struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start decoding roomNode packets
func (w *RoomNode) Run() {
	active := false
	for m := range w.In {
		switch v := m.(type) {

		case string:
			active = v == "<Node-roomNode>"

		case []byte:
			if active {
				m = map[string]int{
					"<reading>": 1,
					"temp":      (int(v[3]) + int(v[4])<<8) & 0x3FF,
					"humi":      int(v[2] >> 1),
					"light":     int(v[1]),
					"moved":     int(v[2] & 1),
				}
			}

		default:
			active = false
		}

		w.Out.Send(m)
	}
}
