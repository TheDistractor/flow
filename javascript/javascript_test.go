package javascript

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleJavaScript() {
	g := flow.NewGroup()
	g.Add("js", "JavaScript")
	g.Set("js.Cmd", `console.log("Hello from Otto!");`)
	g.Run()
	// Output:
	// Hello from Otto!
}

func ExampleJavaScript_2() {
	cmd := `
		console.log("Hello from Otto!");
		function onIn(v) {
			console.log("Got:", v);
			emitOut(3 * v)
		}
	`
	g := flow.NewGroup()
	g.Add("js", "JavaScript")
	g.Add("p", "Printer")
	g.Connect("js.Out", "p.In", 0)
	g.Set("js.Cmd", cmd)
	g.Set("js.In", 123)
	g.Run()
	// Output:
	// Hello from Otto!
	// Got: 123
	// float64: 369
}
