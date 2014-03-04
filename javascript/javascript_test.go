package javascript

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/javascript"
)

func ExampleJavaScript() {
	g := flow.NewGroup()
	g.Add("js", "JavaScript")
	g.Request(`console.log("Hello from Otto!");`, "js.Cmd")
	g.Run()
	// Output:
	// Hello from Otto!
}
