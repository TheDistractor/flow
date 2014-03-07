package decoders

import (
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-ookDcf"] = func() flow.Worker { return &OokDcf{} }
}

// Decoder for the "ookDcf.ino" sketch. Registers as "Node-ookDcf".
type OokDcf struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start decoding ookDcf packets
func (w *OokDcf) Run() {
	for m := range w.In {
		if v, ok := m.([]byte); ok && len(v) >= 6 {
			date := ((2000+int(v[0]))*100+int(v[1]))*100 + int(v[2])
			m = map[string]int{
				"<reading>": 1,
				"date":      date,
				"tod":       int(v[3])*100 + int(v[4]),
				"dst":       int(v[5]),
			}
		}

		w.Out.Send(m)
	}
}
