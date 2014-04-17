// Basic collection of pre-defined gadgets.
package gadgets

import (
	"github.com/jcw/flow"
)

func init() {
	flow.Registry["Pipe"] = func() flow.Circuitry { return new(Pipe) }
}


// Pipes are gadgets with an "In" and an "Out" pin. Registers as "Pipe".
type Pipe struct {
	flow.Gadget
	In  flow.Input
	Out flow.Output
}

// Start passing through messages.
func (w *Pipe) Run() {
	for m := range w.In {
		w.Out.Send(m)
	}
}

