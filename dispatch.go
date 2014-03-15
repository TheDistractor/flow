package flow

func init() {
	Registry["Dispatcher"] = func() Circuitry {
		g := NewCircuit()
		g.AddCircuitry("head", &dispatchHead{})
		g.AddCircuitry("tail", &dispatchTail{})
		g.Connect("head.Feeds:", "tail.In", 0)  // keeps tail alive
		g.Connect("tail.Back", "head.Reply", 1) // must have room for reply
		g.Label("In", "head.In")
		g.Label("Rej", "head.Rej")
		g.Label("Out", "tail.Out")
		return g
	}
}

// A dispatcher sends messages to newly created gadgets, based on dispatch tags.
// These gadgets must have an In and an Out pin. Their output is merged into
// a single Out pin, the rest is sent to Rej. Registers as "Dispatcher".
type Dispatcher Circuit

// The implementation uses a circuit with dispatchHead and dispatchTail gadgets.
// Newly created gadgets are inserted "between" them, using Feeds as fanout.
// Switching needs special care to drain the preceding gadget output first.

type dispatchHead struct {
	Gadget
	In    Input
	Reply Input
	Feeds map[string]Output
	Rej   Output
}

func (w *dispatchHead) Run() {
	gadget := ""
	for m := range w.In {
		if tag, ok := m.(Tag); ok && tag.Tag == "<dispatch>" {
			if tag.Msg == gadget {
				continue
			}

			// send (unique!) marker and act on it once it comes back on Reply
			w.Feeds[gadget].Send(Tag{"<marker>", w.circuit})
			<-w.Reply // TODO: add a timeout?

			// perform the switch, now that previous output has drained
			gadget = tag.Msg.(string)
			if w.Feeds[gadget] == nil {
				if Registry[gadget] == nil {
					w.Rej.Send(tag) // report that no such gadget was found
					gadget = ""
				} else { // create, hook up, and launch the new gadget
					println("Dispatching to new gadget: " + gadget)
					g := w.circuit
					g.Add(gadget, gadget)
					g.Connect("head.Feeds:"+gadget, gadget+".In", 0)
					g.Connect(gadget+".Out", "tail.In", 0)
					g.gadgets[gadget].launch()
				}
			}

			// pass through a "consumed" dispatch tag
			w.Feeds[""].Send(Tag{"<dispatched>", gadget})
			continue
		}

		feed := w.Feeds[gadget]
		if feed == nil {
			feed = w.Rej
		}
		feed.Send(m)
	}
}

type dispatchTail struct {
	Gadget
	In   Input
	Back Output
	Out  Output
}

func (w *dispatchTail) Run() {
	for m := range w.In {
		if tag, ok := m.(Tag); ok && tag.Tag == "<marker>" && tag.Msg == w.circuit {
			w.Back.Send(m)
		} else {
			w.Out.Send(m)
		}
	}
}
