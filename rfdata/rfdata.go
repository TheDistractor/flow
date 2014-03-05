package rfdata

import (
	"bytes"
	"strings"
	"strconv"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Sketch-RF12demo"] = func() flow.Worker { return &RF12demo{} }
}

type RF12demo struct {
	flow.Work
	In flow.Input
	Out flow.Output
}

// This type is used for each line which has valid packet data.
type Packet struct {
	id   byte
	rssi int
	data []byte
}

// Start converting lines into binary packets.
func (w *RF12demo) Run() {
	if m, ok := <- w.In; ok {
		println(m.(string))
		for m = range w.In {
			if s, ok := m.(string); ok {
				if strings.HasPrefix(s, "OK ") {
					s = strings.TrimSpace(s[3:])
					var rssi int

					// convert a line of decimal byte values to a byte buffer
					var buf bytes.Buffer
					for _, v := range strings.Split(s, " ") {
						if strings.HasPrefix(v, "(") {
							rssi, _ = strconv.Atoi(v[1 : len(v)-1])
						} else {
							n, _ := strconv.Atoi(v)
							buf.WriteByte(byte(n))
						}
					}
					b := buf.Bytes()

					m = &Packet{b[0] & 0x1F, rssi, b}
				}
			}
			w.Out <- m
		}
	}
}
