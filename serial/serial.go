package serial

import (
	"bufio"
	"strings"

	"github.com/chimera/rs232"
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["SerialIn"] = func() flow.Worker { return &SerialIn{} }
	flow.Registry["SketchType"] = func() flow.Worker { return &SketchType{} }
}

// Line-oriented serial input port, opened once the Port input is set.
type SerialIn struct {
	flow.Worker
	Port flow.Input
	Out  flow.Output
}

// Start processing incoming text lines from the serial interface.
func (w *SerialIn) Run() {
	port := <-w.Port

	opt := rs232.Options{BitRate: 57600, DataBits: 8, StopBits: 1}
	dev, err := rs232.Open(port.Val.(string), opt)
	check(err)

	scanner := bufio.NewScanner(dev)
	for scanner.Scan() {
		w.Out <- flow.NewMemo(scanner.Text())
	}
}

// SketchType looks for lines of the form "[name...]" in the input stream.
// These are turned into "Sketch" tokens, the rest is passed through as is.
type SketchType struct {
	flow.Worker
	In  flow.Input
	Out flow.Output
}

// Start transforming the "[name...]" markers in the input stream.
func (w *SketchType) Run() {
	for m := range w.In {
		if s, ok := m.Val.(string); ok {
			if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
				tag := s[1:strings.IndexAny(s, ".]")]
				type Sketch string
				m = flow.NewMemo(Sketch(tag))
				m.Attr["version"] = s
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
