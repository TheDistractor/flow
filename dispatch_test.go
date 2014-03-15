package flow_test

import (
	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func ExampleDispatcher() {
	g := flow.NewCircuit()
	g.Add("d", "Dispatcher")
	g.Feed("d.In", "abc")
	g.Feed("d.In", flow.Tag{"<dispatch>", "Counter"})
	g.Feed("d.In", "def")
	g.Feed("d.In", "ghi")
	g.Feed("d.In", flow.Tag{"<dispatch>", ""})
	g.Feed("d.In", "jkl")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost flow.Tag: {<dispatched> Counter}
	// Lost flow.Tag: {<dispatched> }
	// Lost string: jkl
	// Lost int: 2
}
