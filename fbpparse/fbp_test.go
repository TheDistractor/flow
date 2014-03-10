package fbpparse

import (
	"github.com/jcw/flow/flow"
)

func ExampleFbpParser() {
	g := flow.NewGroup()
	g.Add("f", "FbpParser")
	// see https://github.com/noflo/fbp/blob/master/spec/fbp.coffee
	g.Set("f.In", `
	    '8003' -> LISTEN WebServer(HTTP/Server) REQUEST -> IN Profiler(HTTP/Profiler) OUT -> IN Authentication(HTTP/BasicAuth)
	    Authentication() OUT -> IN GreetUser(HelloController) OUT -> IN WriteResponse(HTTP/WriteResponse) OUT -> IN Send(HTTP/SendResponse)
	    'hello.jade' -> SOURCE ReadTemplate(ReadFile) OUT -> TEMPLATE Render(Template)
	    GreetUser() DATA -> OPTIONS Render() OUT -> STRING WriteResponse()
		INPORT=ABC.DEF:123
	`)
	g.Run()
	// Output:
	// Lost bool: true
}
