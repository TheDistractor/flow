package decoders

import (
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-ookRelay"] = func() flow.Worker { return &OokRelay{} }
}

var ookDecoders = []string{
	"Dcf", "Viso", "Emx", "Ksx", "Fsx", "Orsc", "Cres", "Kaku",
	"Xrf", "Hez", "Elro", "?11", "?12", "?13", "?14", "?15",
}

// Decoder for the "ookRelay.ino" sketch. Registers as "Node-ookRelay".
type OokRelay struct {
	flow.Work
	In   flow.Input
	Type flow.Output
	Out  flow.Output
}

// Start decoding ookRelay packets
func (w *OokRelay) Run() {
	active := false
	for m := range w.In {
		switch v := m.(type) {

		case string:
			active = v == "<Node-ookRelay>"

		case []byte:
			if active {
				offset := 1
				for offset < len(v) {
					typ := int(v[offset] & 0x0F)
					size := int(v[offset] >> 4)
					offset++

					// insert a new decoder request
					w.Type.Send("Node-ook" + ookDecoders[typ])
					w.Out.Send("<Node-ook" + ookDecoders[typ] + ">")
					w.Out.Send(v[offset : offset+size])

					offset += size
				}
				continue
			}

		default:
			active = false
		}

		w.Out.Send(m)
	}
}
