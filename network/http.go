package network

import (
	"net/http"
	"time"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["HttpServer"] = func() flow.Worker { return &HttpServer{} }
}

type HttpServer struct {
	flow.Work
	Handlers flow.Input
	Start    flow.Input
	Out      flow.Output
}

type flowHandler struct {
	h http.Handler
	s *HttpServer
}

func (fh *flowHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fh.s.Out.Send(req.URL)
	fh.h.ServeHTTP(w, req)
}

func (w *HttpServer) Run() {
	mux := http.NewServeMux() // don't use default to allow multiple instances
	for m := range w.Handlers {
		tag := m.(flow.Tag)
		mux.Handle(tag.Tag, &flowHandler{tag.Val.(http.Handler), w})
	}
	m := <-w.Start
	go func() {
		// will stay running until an error is returned or the app ends
		panic(http.ListenAndServe(m.(string), mux))
	}()
	// TODO: this is a hack to make sure the server is ready
	// better would be to interlock the goroutine with the listener being ready
	time.Sleep(10 * time.Millisecond)
}
