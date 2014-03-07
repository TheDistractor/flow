package decoders

import (
	"bytes"
	"encoding/binary"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-radioBlip"] = func() flow.Worker { return &RadioBlip{} }
}

// Decoder for the "radioBlip.ino" sketch. Registers as "Node-radioBlip".
type RadioBlip struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start decoding radioBlip packets
func (w *RadioBlip) Run() {
	for m := range w.In {
		if v, ok := m.([]byte); ok && len(v) >= 4 {
			buf := bytes.NewBuffer(v[1:])
			var ping uint32
			err := binary.Read(buf, binary.LittleEndian, &ping)
			if err != nil {
				panic(err)
			}

			result := map[string]int{
				"<reading>": 1,
				"ping":      int(ping),
				"age":       int(ping / (86400 / 64)),
			}

			if len(v) >= 8 {
				result["tag"] = int(v[5] & 0x7F)
				result["vpre"] = 50 + int(v[6])
				if v[5]&0x80 != 0 {
					// if high bit of id is set, this is a boost node
					// reporting its battery -  this is ratiometric
					// (proportional) w.r.t. the "vpre" just measured
					result["vbatt"] = result["vpre"] * int(v[7]) / 255
				} else if v[7] != 0 {
					// in the non-boost case, the second value is vcc
					// after the previous transmit -  this is always set,
					// except in the first transmission after power-up
					result["vpost"] = 50 + int(v[7])
				}
			}

			m = result
		}

		w.Out.Send(m)
	}
}
