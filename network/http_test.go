package network

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleHttpServer() {
	g := flow.NewGroup()
	g.Add("s", "HttpServer")
	g.Set("s.Handlers", flow.Tag{"/blah/", "../flow"})
	g.Set("s.Start", ":12345")
	g.Run()
	res, _ := http.Get("http://:12345/blah/flow.go")
	body, _ := ioutil.ReadAll(res.Body)
	data, _ := ioutil.ReadFile("../flow/flow.go")
	fmt.Println(string(body) == string(data))
	// Output:
	// Lost *url.URL: /blah/flow.go
	// true
}
