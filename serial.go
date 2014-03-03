package flow

import (
	"bufio"
	"strings"

	"github.com/chimera/rs232"
)

func init() {
	Registry["SerialIn"] = func() Worker { return new(SerialIn) }
	Registry["SketchType"] = func() Worker { return new(SketchType) }
}

// Line-oriented serial input port, opened once the Port input is set.
type SerialIn struct {
	Worker
	Port Input
	Out  Output
}

// Start processing incoming text lines from the serial interface.
func (w *SerialIn) Run() {
	port := <-w.Port

	opt := rs232.Options{BitRate: 57600, DataBits: 8, StopBits: 1}
	dev, err := rs232.Open(port.Val.(string), opt)
	Check(err)

	scanner := bufio.NewScanner(dev)
	for scanner.Scan() {
		w.Out <- NewMemo(scanner.Text())
	}
}

// SketchType looks for lines of the form "[name...]" in the input stream.
// These are turned into "Sketch" tokens, the rest is passed through as is.
type SketchType Pipe

// Start transforming the "[name...]" markers in the input stream.
func (w *SketchType) Run() {
	for m := range w.In {
		if s, ok := m.Val.(string); ok {
			if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
				tag := s[1:strings.IndexAny(s, ".]")]
				type Sketch string
				m = NewMemo(Sketch(tag))
				m.Attr["version"] = s
			}
		}
		w.Out <- m
	}
}
