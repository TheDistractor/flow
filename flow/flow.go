package flow

// Version of this package.
var Version = "0.1.0"

// The registry is the factory for all known types of workers.
var Registry = map[string]func() Worker{}

// Memo's are the generic type sent to, between, and from workers.
type Memo interface{}

// Input ports are used to receive memo's.
type Input <-chan Memo

// Output ports are used to send memo's elsewhere.
type Output interface {
	Send(v Memo)
	Close()
}

// The worker is the basic unit of processing, shuffling memo's between ports.
type Worker interface {
	Run()

	initWork(Worker, string, *Group) *Work
}

// A transformer processes each memo through a supplied function.
func Transformer(f func(Memo) Memo) Worker {
	return &transformer{fun: f}
}

type transformer struct {
	Work
	In  Input
	Out Output

	fun func(Memo) Memo
}

func (w *transformer) Run() {
	for m := range w.In {
		w.Out.Send(w.fun(m))
	}
}
