package javascript

import (
	"github.com/jcw/flow/flow"
)

func ExampleJavaScript() {
	g := flow.NewGroup()
	g.Add("js", "JavaScript")
	g.Set("js.Cmd", `console.log("Hello from Otto!");`)
	g.Run()
	// Output:
	// Hello from Otto!
}
