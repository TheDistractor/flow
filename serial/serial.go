// Interface to serial port devices.
package serial

import (
	"bufio"
	"strings"
	"time"

	"github.com/chimera/rs232"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["TimeStamp"] = func() flow.Worker { return &TimeStamp{} }
	flow.Registry["SerialIn"] = func() flow.Worker { return &SerialIn{} }
	flow.Registry["SketchType"] = func() flow.Worker { return &SketchType{} }
}

// Insert a timestamp before each message. Registers as "TimeStamp".
type TimeStamp struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start inserting timestamps.
func (w *TimeStamp) Run() {
	for m := range w.In {
		w.Out.Send(time.Now())
		w.Out.Send(m)
	}
}

// Line-oriented serial input port, opened once the Port input is set.
type SerialIn struct {
	flow.Work
	Port flow.Input
	Out  flow.Output
}

// Start processing incoming text lines from the serial interface.
// Registers as "SerialIn".
func (w *SerialIn) Run() {
	if port, ok := <-w.Port; ok {
		opt := rs232.Options{BitRate: 57600, DataBits: 8, StopBits: 1}
		dev, err := rs232.Open(port.(string), opt)
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(dev)
		for scanner.Scan() {
			w.Out.Send(scanner.Text())
		}
	}
}

// SketchType looks for lines of the form "[name...]" in the input stream.
// These then cause a corresponding worker to be loaded dynamically.
// Registers as "SketchType".
type SketchType struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start transforming the "[name...]" markers in the input stream.
func (w *SketchType) Run() {
	for m := range w.In {
		if s, ok := m.(string); ok {
			if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
				tag := "Sketch-" + s[1:strings.IndexAny(s, ".]")]
				m = flow.Tag{"<dispatch>", tag}
			}
		}
		w.Out.Send(m)
	}
}
