package flow_test

import (
	"strings"
	
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleDispatcher() {
	flow.Registry["upper"] = func() flow.Worker {
		return flow.Transformer(func(m flow.Memo) flow.Memo {
			if s, ok := m.(string); ok {
				m = strings.ToUpper(s)
			}
			return m
		})
	} 
	
	g := flow.NewGroup()
	g.Add("d", "Dispatcher")
	g.Set("d.In", "abc")
	g.Set("d.In", flow.Tag{"dispatch", "upper"})
	g.Set("d.In", "def")
	g.Set("d.In", "ghi")
	g.Set("d.In", flow.Tag{"dispatch", ""})
	g.Set("d.In", "jkl")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost flow.Tag: {dispatched upper}
	// Lost string: DEF
	// Lost string: GHI
	// Lost flow.Tag: {dispatched }
	// Lost string: jkl
}
