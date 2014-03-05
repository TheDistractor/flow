// Interface to serial port devices.
package serial

import (
	"bufio"
	"strings"

	"github.com/chimera/rs232"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["SerialIn"] = func() flow.Worker { return &SerialIn{} }
	flow.Registry["SketchType"] = func() flow.Worker { return &SketchType{} }
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
// These then cause a corresponding worker to be loaded dynamically.
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
				// FIXME: this code is a horrible hack
				if _, ok := flow.Registry[tag]; ok {
					// must create a new workgroup to insert the new worker
					wg := flow.NewGroup()
					wg.Add("sketch", tag)
					wg.Map("In", "sketch.In")
					// dynamically insert this new group and connect to it
					g := w.MyGroup()
					g.AddWorker("(sketch)", wg)
					g.Connect(w.MyName() + ".Out", "(sketch).In", 0)
					// start the new group running
					go wg.Run()
				}
			}
			w.Out <- m
		}
	}
}

// Generic error checking, panics if e is not nil.
func check(e error) {
	if e != nil {
		panic(e)
	}
}
