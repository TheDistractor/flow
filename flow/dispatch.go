package flow

import (
	"fmt"
)

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

// special marker sent through to determine when to switch
// TODO: relies on the marker's address, won't work through a remoted stream
var marker struct{}

type dispatchFront struct {
	Work
	In    Input
	Use   Input
	SwIn  Input
	Feeds map[string]Output
	Out   Output
	Rej   Output
}

func (w *dispatchFront) Run() {
	worker := ""
	for {
		select {
		case m := <-w.Use:
			if m != nil {
				// send a marker, will act on it when it comes back on SwIn
				w.Feeds[worker].Send(marker)
				<-w.SwIn
				// perform the switch, now that previous output has drained
				worker = m.(string)
				if w.Feeds[worker] == nil {
					if Registry[worker] == nil {
						w.Rej.Send(m) // report that no such worker was found
						worker = ""
					} else { // create, hook up, and launch the new worker
						fmt.Println("Dispatching to new worker:", worker)
						g := w.parent
						g.Add(worker, worker)
						g.Connect("front.Feeds:"+worker, worker+".In", 0)
						g.Connect(worker+".Out", "back.In", 0)
						g.Launch(worker)
					}
				}
			}

		case m := <-w.In:
			if m == nil {
				return
			}
			feed := w.Feeds[worker]
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
		if m == marker {
			w.SwOut.Send(m)
		} else {
			w.Out.Send(m)
		}
	}
}
