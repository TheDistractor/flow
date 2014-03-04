// Interface to serial port devices.
package serial

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/chimera/rs232"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["SerialIn"] = func() flow.Worker { return &SerialIn{} }
	flow.Registry["SketchType"] = func() flow.Worker { return &SketchType{} }
	flow.Registry["RFpacket"] = func() flow.Worker { return &RFpacket{} }
}

// Line-oriented serial input port, opened once the Port input is set.
type SerialIn struct {
	flow.Work
	Port flow.Input
	Out  flow.Output
}

// Start processing incoming text lines from the serial interface.
func (w *SerialIn) Run() {
	if port, ok := <-w.Port; ok {
		opt := rs232.Options{BitRate: 57600, DataBits: 8, StopBits: 1}
		dev, err := rs232.Open(port.(string), opt)
		check(err)

		scanner := bufio.NewScanner(dev)
		for scanner.Scan() {
			w.Out <- scanner.Text()
		}
	}
}

// SketchType looks for lines of the form "[name...]" in the input stream.
// These are turned into "Sketch" tokens, the rest is passed through as is.
type SketchType struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// This type is inserted as marker before each "[name...]" line.
type Sketch string

// Start transforming the "[name...]" markers in the input stream.
func (w *SketchType) Run() {
	for m := range w.In {
		if s, ok := m.(string); ok {
			if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
				tag := s[1:strings.IndexAny(s, ".]")]
				w.Out <- Sketch(tag)
			}
		}
		w.Out <- m
	}
}

// Generic error checking, panics if e is not nil.
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Comvert lines starting with "OK " into binary packets.
type RFpacket struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// This type is used for each line which has valid packet data.
type Packet struct {
	id   byte
	rssi int
	data []byte
}

// Start converting lines into binary packets.
func (w *RFpacket) Run() {
	for m := range w.In {
		if s, ok := m.(string); ok {
			if strings.HasPrefix(s, "OK ") {
				s = strings.TrimSpace(s[3:])
				var rssi int

				// convert the line of decimal byte values to a byte buffer
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
