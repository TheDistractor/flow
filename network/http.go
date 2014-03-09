package network

import (
	"net"
	"net/http"

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

type FlowHandler struct {
	h http.Handler
	s *HttpServer
}

func (fh *FlowHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fh.s.Out.Send(req.URL)
	fh.h.ServeHTTP(w, req)
}

func (w *HttpServer) Run() {
	mux := http.NewServeMux() // don't use default to allow multiple instances
	for m := range w.Handlers {
		tag := m.(flow.Tag)
		mux.Handle(tag.Tag, &FlowHandler{tag.Val.(http.Handler), w})
	}
	m := <-w.Start
	go func() {
		server := &http.Server{Handler: mux}
		listener, err := net.Listen("tcp", m.(string))
		if err != nil {
			panic(err)
		}
		go server.Serve(listener)
		// stay around until the Start input port is closed
		<-w.Start
		listener.Close()
	}()
}
