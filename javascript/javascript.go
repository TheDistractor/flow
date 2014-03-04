// An embedded JavaScript engine as worker.
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
		engine, result, err := otto.Run(cmd.(string))
		if err != nil {
			panic(err)
		}
		if !result.IsUndefined() {
			w.Out <- result
		}
		// define a callback for send memos to Out
		engine.Set("emitOut", func(call otto.FunctionCall) otto.Value {
			out, err := call.Argument(0).Export()
			if err != nil {
				panic(err)
			}
			w.Out <- out
			return otto.UndefinedValue()
		})
		// enter the processing loop
		for in := range w.In {
			engine.Call("onIn", nil, in)
		}
	}
}
