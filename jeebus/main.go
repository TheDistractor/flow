// This application can exercise the "flow" package via a JSON config file.
package main

import (
	"fmt"
	"time"

	"github.com/jcw/flow/flow"

	_ "github.com/jcw/flow/database"
	_ "github.com/jcw/flow/decoders"
	_ "github.com/jcw/flow/javascript"
	_ "github.com/jcw/flow/network"
	_ "github.com/jcw/flow/rfdata"
	_ "github.com/jcw/flow/serial"
	_ "github.com/jcw/flow/workers"
)

func main() {
	defer flow.DontPanic()

	// started for each websocket connection with protocol type "jeebus"
	flow.Registry["WebSocket-jeebus"] = func() flow.Worker {
		trace := flow.Transformer(func(m flow.Memo) flow.Memo {
			fmt.Println("ws:", m)
			return m
		})
		g := flow.NewGroup()
		g.AddWorker("t", trace)
		g.Map("In", "t.In")
		g.Map("Out", "t.Out")
		return g
	}

	g := flow.NewGroup()
	g.Add("http", "HttpServer")
	g.Set("http.Handlers", flow.Tag{"/", "../jeebus/app"})
	g.Set("http.Handlers", flow.Tag{"/base/", "../jeebus/base"})
	g.Set("http.Handlers", flow.Tag{"/common/", "../jeebus/common"})
	g.Set("http.Handlers", flow.Tag{"/ws", "<websocket>"})
	g.Set("http.Start", ":3000")
	g.Run()

	println("listening on http://localhost:3000/")
	time.Sleep(1e6 * time.Hour)
}
