// Driver and decoders for RF12/RF69 packet data.
package rfdata

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["Sketch-RF12demo"] = func() flow.Worker { return &RF12demo{} }
}

type RF12demo struct {
	flow.Work
	In  flow.Input
	Out flow.Output
	Rej flow.Output
}

// Start converting lines into binary packets.
func (w *RF12demo) Run() {
	if m, ok := <-w.In; ok {
		config := parseConfigLine(m.(string))
		w.Out.Send(config)
		for m = range w.In {
			if s, ok := m.(string); ok {
				if strings.HasPrefix(s, "OK ") {
					data, rssi := convertToBytes(s)
					info := map[string]int{
						"<node>": int(data[0] & 0x1F),
						"rssi":   rssi,
					}
					w.Out.Send(info)
					w.Out.Send(data)
				} else {
					w.Rej.Send(m)
				}
			} else {
				w.Out.Send(m) // not a string
			}
		}
	}
}

// Parse lines of the form "[RF12demo.12] _ i31* g5 @ 868 MHz c1 q1"
var re = regexp.MustCompile(`\.(\d+)] . i(\d+)\*? g(\d+) @ (\d+) MHz`)

func parseConfigLine(s string) map[string]int {
	m := re.FindStringSubmatch(s)
	v, _ := strconv.Atoi(m[1])
	i, _ := strconv.Atoi(m[2])
	g, _ := strconv.Atoi(m[3])
	b, _ := strconv.Atoi(m[4])
	return map[string]int{"<RF12demo>": v, "band": b, "group": g, "id": i}
}

func convertToBytes(s string) ([]byte, int) {
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
	return buf.Bytes(), rssi
}
