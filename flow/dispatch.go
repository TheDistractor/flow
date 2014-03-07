package flow

func init() {
	Registry["Dispatcher"] = func() Worker { return &Dispatcher{} }
}

// A dispatcher sends memos to new workers, as determined by the use port.
// These workers must have an In and an Out port. The rest is sent to Rej.
type Dispatcher struct {
	Work
	In  Input
	Use Input
	Out Output
	Rej Output

	feeds   map[string]Output
	replies Input
	worker  string
}

type marker string // special marker sent through to determine when to switch

// Start dispatching to the worker named in the Use port.
func (w *Dispatcher) Run() {
	useChan := w.Use
	for {
		select {
		case m := <-useChan:
			if m == nil {
				useChan = nil // stop listening
			} else if feed, ok := w.feeds[w.worker]; ok {
				useChan = nil                    // suspended
				go feed.Send(marker(m.(string))) // will eventually fire
			} else {
				w.switchToWorker(m.(string))
			}

		case m := <-w.In:
			if m == nil {
				for _, f := range w.feeds {
					f.Close()
				}
				// let feeds drain, ends when workers close replies port
			} else if feed, ok := w.feeds[w.worker]; ok {
				feed.Send(m)
			} else {
				w.Rej.Send(m)
			}

		case m, ok := <-w.replies:
			if !ok {
				return
			}
			if m, ok := m.(marker); ok {
				// the marker came back, switch to the requested worker
				w.switchToWorker(string(m))
			} else {
				w.Out.Send(m)
			}
		}
	}
}

func (w *Dispatcher) switchToWorker(name string) {
	if Registry[name] == nil {
		w.Rej.Send(marker(name)) // report that no such worker was found
		name = ""
	}
	w.worker = name
	if name != "" && w.feeds[name] == nil {
		g := w.parent
		g.Add(name, name)
		g.Connect(w.name+".feeds."+name, name+".In", 0)
		g.Connect(name+".Out", w.name+".replies", 0)
		// w.feeds[name] = ...
	}
}
