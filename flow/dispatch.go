package flow

func init() {
	Registry["Dispatcher"] = func() Worker {
		g := NewGroup()
		g.AddWorker("front", &dispatchFront{})
		g.AddWorker("back", &dispatchBack{})
		g.Connect("front.Feeds:", "back.In", 0)  // fallback for marker
		g.Connect("back.SwOut", "front.SwIn", 1) // must have room for reply
		g.Map("In", "front.In")
		g.Map("Use", "front.Use")
		g.Map("Rej", "front.Rej")
		g.Map("Out", "back.Out")
		return g
	}
}

// A dispatcher sends memos to newly created workers, as set in the use port.
// These workers must have an In and an Out port. Their output is merged into
// a single Out port, the rest is sent to Rej. Registers as "Dispatcher".
type Dispatcher Group

// The implementation uses a group with dispatchFront and dispatchBack workers.
// Newly created workers are inserted "between" them, using Feeds as fanout.
// Switching needs special care to drain the preceding worker output first.

type marker string // special marker sent through to determine when to switch

type dispatchFront struct {
	Work
	In    Input
	Use   Input
	SwIn  Input
	Feeds map[string]Output
	Out   Output
	Rej   Output

	worker  string
}

func (w *dispatchFront) Run() {
	useChan := w.Use
	for {
		select {
		case m := <-useChan:
			useChan = nil // suspend
			if m != nil {
				// send a marker, will act on it when it comes back on SwIn
				w.Feeds[w.worker].Send(marker(m.(string)))
			}

		case m := <-w.SwIn:
			useChan = w.Use // resume
			w.worker = string(m.(marker))
			sw := w.worker
			if w.Feeds[sw] == nil {
				if Registry[sw] == nil {
					w.Rej.Send(m) // report that no such worker was found
					w.worker = ""
				} else { // create, hook up, and launch the new worker
					g := w.parent
					g.Add(sw, sw)
					g.Connect("front.Feeds:"+sw, sw+".In", 0)
					g.Connect(sw+".Out", "back.In", 0)
					g.Launch(sw)
				}
			}

		case m := <-w.In:
			if m == nil {
				return
			}
			feed := w.Feeds[w.worker]
			if feed == nil {
				feed = w.Rej
			}
			feed.Send(m)
		}
	}
}

type dispatchBack struct {
	Work
	In    Input
	SwOut Output
	Out   Output
}

func (w *dispatchBack) Run() {
	for m := range w.In {
		if _, ok := m.(marker); ok {
			w.SwOut.Send(m)
		} else {
			w.Out.Send(m)
		}
	}
}
