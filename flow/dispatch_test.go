package flow_test

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleDispatcher() {
	g := flow.NewGroup()
	g.Add("d", "Dispatcher")
	g.Set("d.In", "abc")
	g.Set("d.In", flow.Tag{"<dispatch>", "Counter"})
	g.Set("d.In", "def")
	g.Set("d.In", "ghi")
	g.Set("d.In", flow.Tag{"<dispatch>", ""})
	g.Set("d.In", "jkl")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost flow.Tag: {<dispatched> Counter}
	// Lost flow.Tag: {<dispatched> }
	// Lost string: jkl
	// Lost int: 2
}
