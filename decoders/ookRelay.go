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
	Out  flow.Output
}

// Start decoding ookRelay packets
func (w *OokRelay) Run() {
	for m := range w.In {
		if v, ok := m.([]byte); ok {
			offset := 1
			for offset < len(v) {
				typ := int(v[offset] & 0x0F)
				size := int(v[offset] >> 4)
				offset++

				// insert a new decoder request
				tag := "Node-ook" + ookDecoders[typ]
				w.Out.Send(&flow.Tag{"dispatch", tag})
				w.Out.Send(v[offset : offset+size])

				offset += size
			}
		} else {
			w.Out.Send(m)
		}
	}
}
