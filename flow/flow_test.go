package flow_test

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleNewGroup() {
	g := flow.NewGroup()
	g.Run()
}

func ExampleGroup_Map() {
	// new workgroup to convert the input to upper case and repeat it 3 times
	wg := flow.NewGroup()
	wg.Add("u", "ToUpper")
	wg.Add("r", "Repeater")
	wg.Connect("u.Out", "r.In", 0)
	wg.Request(3, "r.Num")
	wg.Map("MyIn", "u.In")
	wg.Map("MyOut", "r.Out")

	g := flow.NewGroup()
	g.AddWorker("wg", wg)
	g.Add("p", "Printer")
	g.Connect("wg.MyOut", "p.In", 0)
	g.Request("abc", "wg.MyIn")

	g.Run()

	// Output:
	// string: ABC
	// string: ABC
	// string: ABC
}
