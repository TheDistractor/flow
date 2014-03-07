package decoders

import (
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Node-p1scanner"] = func() flow.Worker { return &P1scanner{} }
}

// Decoder for the "p1scanner.ino" sketch. Registers as "Node-p1scanner".
type P1scanner struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// see http://jeelabs.org/2012/12/01/extracting-data-from-p1-packets/
var params = []string{
	"", "use1", "use2", "gen1", "gen2", "mode", "usew", "genw", "", "gas",
}

// Start decoding p1scanner packets.
func (w *P1scanner) Run() {
	prev := []int{}
	for m := range w.In {
		if v, ok := m.([]byte); ok {
			vec := []int{}
			val := 0
			for _, b := range v[1:] {
				val = val<<7 + int(b&0x7F)
				if b&0x80 != 0 {
					vec = append(vec, val)
					val = 0
				}
			}
			// only report values which have actually changed
			// for usew and genw, only report the one that is active
			result := map[string]int{"<reading>": 1}
			if len(vec) >= 11 && vec[0] == 1 {
				for i, s := range params {
					switch s {
					case "":
						// skip
					case "usew", "genw":
						if vec[i] != 0 || i >= len(prev) || vec[i] != prev[i] {
							result[s] = vec[i]
						}
					default:
						if i >= len(prev) || vec[i] != prev[i] {
							result[s] = vec[i]
						}
					}
				}
			}
			prev = vec
			m = result
		}

		w.Out.Send(m)
	}
}
