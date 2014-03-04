package flow_test

import "github.com/jcw/flow/flow"

func ExampleNewGroup() {
	g := flow.NewGroup()
	g.Run()
}
