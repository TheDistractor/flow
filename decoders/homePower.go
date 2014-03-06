package decoders

import (
	"bytes"
	"encoding/binary"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-homePower"] = func() flow.Worker { return &HomePower{} }
}

// Decoder for the "homePower.ino" sketch.
type HomePower struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start decoding homePower packets
func (w *HomePower) Run() {
	active := false
	var vec, prev [6]uint16
	for m := range w.In {
		switch v := m.(type) {

		case string:
			active = v == "<Node-homePower>"

		case []byte:
			if active {
				buf := bytes.NewBuffer(v[1:])
				err := binary.Read(buf, binary.LittleEndian, &vec)
				if err != nil {
					panic(err)
				}
				results := map[string]int{"<reading>": 1}
				if vec[0] != prev[0] {
					results["c1"] = int(vec[0])
					results["p1"] = time2watt(int(vec[1]))
				}
				if vec[2] != prev[2] {
					results["c2"] = int(vec[2])
					results["p2"] = time2watt(int(vec[3]))
				}
				if vec[4] != prev[4] {
					results["c3"] = int(vec[4])
					results["p3"] = time2watt(int(vec[5]))
				}
				copy(prev[:], vec[:])
				if len(results) == 1 {
					continue
				}
				m = results
			}

		default:
			active = false
		}

		w.Out.Send(m)
	}
}

func time2watt(t int) int {
	if t > 60000 {
		t = 1000 * (t - 60000)
	}
	if t > 0 {
		t = 18000000 / t
	}
	return t
}
