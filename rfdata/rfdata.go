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
		w.Out.Send(*config)
		for m = range w.In {
			if s, ok := m.(string); ok {
				if strings.HasPrefix(s, "OK ") {
					w.Out.Send(convertToPacket(s))
				} else {
					w.Rej.Send(m)
				}
			}
		}
	}
}

// Parse lines of the form "[RF12demo.12] _ i31* g5 @ 868 MHz c1 q1"
var re = regexp.MustCompile(` i(\d+)\*? g(\d+) @ (\d+) MHz`)

func parseConfigLine(s string) *map[string]int {
	m := re.FindStringSubmatch(s)
	b, _ := strconv.Atoi(m[3])
	g, _ := strconv.Atoi(m[2])
	i, _ := strconv.Atoi(m[1])
	return &map[string]int{"b": b, "g": g, "i": i}
}

// This type is used for each line which has valid packet data.
type Packet struct {
	Id   byte   // node ID, 1..31
	Rssi int    // RSSI signal level, if present in input
	Data []byte // binary packet contents, including the header byte
}

func convertToPacket(s string) *Packet {
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

	return &Packet{b[0] & 0x1F, rssi, b}
}
