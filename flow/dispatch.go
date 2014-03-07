package flow

import (
	"fmt"
)

func init() {
	Registry["Dispatcher"] = func() Worker {
		g := NewGroup()
		g.AddWorker("head", &dispatchHead{})
		g.AddWorker("tail", &dispatchTail{})
		g.Connect("head.Feeds:", "tail.In", 0)  // fallback for marker
		g.Connect("tail.Back", "head.Reply", 1) // must have room for reply
		g.Map("In", "head.In")
		g.Map("Rej", "head.Rej")
		g.Map("Out", "tail.Out")
		return g
	}
}

// A dispatcher sends memos to newly created workers, based on dispatch tags.
// These workers must have an In and an Out port. Their output is merged into
// a single Out port, the rest is sent to Rej. Registers as "Dispatcher".
type Dispatcher Group

// The implementation uses a group with dispatchHead and dispatchTail workers.
// Newly created workers are inserted "between" them, using Feeds as fanout.
// Switching needs special care to drain the preceding worker output first.

// special marker sent through to determine when to switch
// TODO: relies on the marker's address, won't work through a remoted stream
var marker struct{}

type dispatchHead struct {
	Work
	In    Input
	Reply Input
	Feeds map[string]Output
	Out   Output
	Rej   Output
}

func (w *dispatchHead) Run() {
	worker := ""
	for m := range w.In {
		if tag, ok := m.(Tag); ok && tag.Tag == "dispatch" {
			if tag.Val == worker {
				continue
			}

			// send a marker and act on it once it comes back on Reply
			println(fmt.Sprintln("send switch marker:", tag))
			w.Feeds[worker].Send(marker)
			println(fmt.Sprintln("wait for switch to:", tag))
			<-w.Reply // TODO: add a timeout?
			println(fmt.Sprintln("switching to:", tag))

			// perform the switch, now that previous output has drained
			worker = tag.Val.(string)
			if w.Feeds[worker] == nil {
				if Registry[worker] == nil {
					w.Rej.Send(tag) // report that no such worker was found
					worker = ""
				} else { // create, hook up, and launch the new worker
					println("Dispatching to new worker: " + worker)
					g := w.parent
					g.Add(worker, worker)
					g.Connect("head.Feeds:"+worker, worker+".In", 0)
					g.Connect(worker+".Out", "tail.In", 0)
					g.Launch(worker)
				}
			}

			// pass through a "consumed" dispatch tag
			m = Tag{"dispatched", worker}
		}

		feed := w.Feeds[worker]
		if feed == nil {
			feed = w.Rej
		}
		feed.Send(m)
	}
	println("dispatch head ends")
	for _, o := range w.Feeds {
		o.Close()
	}
	// w.Out.Close()
	// w.Rej.Close()
}

type dispatchTail struct {
	Work
	In   Input
	Back Output
	Out  Output
}

func (w *dispatchTail) Run() {
	for m := range w.In {
		if m == marker {
			println("switch marker seen")
			w.Back.Send(m)
			println("reply switch marker")
		} else {
			w.Out.Send(m)
		}
	}
	println("dispatch tail ends")
	w.Back.Close()
	w.Out.Close()
}
