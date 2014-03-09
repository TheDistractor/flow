package network

import (
	"net/http"
	"time"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleHttpServer() {
	g := flow.NewGroup()
	g.Add("p", "Pipe")
	g.Add("t", "Timer")
	g.Add("s", "HttpServer")
	g.Connect("p.Out", "s.Start", 0)
	g.Connect("t.Out", "s.Start", 0)

	// set up a file server for all URLs
	g.Set("s.Handlers", flow.Tag{"/", http.FileServer(http.Dir("."))})

	// will start the HTTP server on port 12345
	g.Set("p.In", ":12345")

	// will stop it again 100 ms later
	g.Set("t.Rate", 100*time.Millisecond)

	// send a GET request to the server after 50 ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, err := http.Get("http://:12345/http.go")
		if err != nil {
			panic(err)
		}
	}()

	g.Run()
	// Output:
	// Lost *url.URL: /http.go
}
