package javascript

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleJavaScript() {
	g := flow.NewGroup()
	g.Add("js", "JavaScript")
	g.Add("p", "Printer")
	g.Connect("js.Out", "p.In", 0)
	g.Request(`console.log("Hello from Otto!");`, "js.Cmd")
	g.Run()
	// Output:
	// Hello from Otto!
}
