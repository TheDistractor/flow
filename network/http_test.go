package network

import (
	"net/http"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleHttpServer() {
	g := flow.NewGroup()
	g.Add("s", "HttpServer")
	g.Set("s.Handlers", flow.Tag{"/", http.FileServer(http.Dir("."))})
	g.Set("s.Start", ":12345")
	g.Run()
	_, err := http.Get("http://:12345/http.go")
	if err != nil {
		panic(err)
	}
	// Output:
	// Lost *url.URL: /http.go
}
