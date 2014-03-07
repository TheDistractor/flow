package decoders

import (
	"bytes"
	"encoding/binary"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-smaRelay"] = func() flow.Worker { return &SmaRelay{} }
}

// Decoder for the "smaRelay.ino" sketch. Registers as "Node-smaRelay".
type SmaRelay struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start decoding smaRelay packets.
func (w *SmaRelay) Run() {
	var vec, prev [7]uint16
	for m := range w.In {
		if v, ok := m.([]byte); ok && len(v) >= 12 {
			buf := bytes.NewBuffer(v[1:])
			err := binary.Read(buf, binary.LittleEndian, &vec)
			if err != nil {
				panic(err)
			}
			result := map[string]int{
				"<reading>": 1,
				"acw":       int(vec[2]),
				"dcv1":      int(vec[3]),
				"dcv2":      int(vec[4]),
			}
			if vec[0] != prev[0] {
				result["yield"] = int(vec[0])
			}
			if vec[1] != prev[1] {
				result["total"] = int(vec[1])
			}
			if vec[2] != 0 {
				result["dcw1"] = int(vec[5])
				result["dcw2"] = int(vec[6])
			}
			copy(prev[:], vec[:])
			m = result
		}

		w.Out.Send(m)
	}
}
