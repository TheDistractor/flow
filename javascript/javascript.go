// Embedded JavaScript engine.
package javascript

import (
	"github.com/jcw/flow/flow"
	"github.com/robertkrimen/otto"
)

func init() {
	flow.Registry["JavaScript"] = func() flow.Worker { return &JavaScript{} }
}

type JavaScript struct {
	flow.Work
	In  flow.Input
	Cmd flow.Input
	Out flow.Output
}

// Start running the JavaScript engine.
func (w *JavaScript) Run() {
	if cmd, ok := <-w.Cmd; ok {
		// initial setup
		engine := otto.New()

		// define a callback for send memos to Out
		engine.Set("emitOut", func(call otto.FunctionCall) otto.Value {
			out, err := call.Argument(0).Export()
			if err != nil {
				panic(err)
			}
			w.Out <- out
			return otto.UndefinedValue()
		})

		// process the command input
		if _, err := engine.Run(cmd.(string)); err != nil {
			panic(err)
		}

		// only start the processing loop if the "onIn" handler exists
		value, err := engine.Get("onIn")
		if err == nil && value.IsFunction() {
			for in := range w.In {
				engine.Call("onIn", nil, in)
			}
		}
	}
}
