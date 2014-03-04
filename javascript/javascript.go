package javascript

import (
	"github.com/robertkrimen/otto"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["JavaScript"] = func() flow.Worker { return &JavaScript{} }
}

type JavaScript struct {
	flow.Work
	Cmd flow.Input
	Out flow.Output
	engine *otto.Otto
}

// Start running the JavaScript engine
func (w *JavaScript) Run() {
	w.engine = otto.New()
	for cmd := range w.Cmd {
		result, err := w.engine.Run(cmd.(string))
		if err != nil {
			panic(err)
		}
		if !result.IsUndefined() {
			w.Out <- result
		}
	}
}
