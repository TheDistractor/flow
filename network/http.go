package network

import (
	"io"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["HttpServer"] = func() flow.Worker { return &HttpServer{} }
}

// HttpServer is a worker which sets up an HTTP server.
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

// Set up the handlers, then start the server and start processing requests.
func (w *HttpServer) Run() {
	mux := http.NewServeMux() // don't use default to allow multiple instances
	for m := range w.Handlers {
		tag := m.(flow.Tag)
		switch v := tag.Val.(type) {
		case string:
			h := createHandler(tag.Tag, v)
			mux.Handle(tag.Tag, &flowHandler{h, w})
		case http.Handler:
			mux.Handle(tag.Tag, &flowHandler{v, w})
		}
	}
	m := <-w.Start
	go func() {
		// will stay running until an error is returned or the app ends
		defer flow.DontPanic()
		panic(http.ListenAndServe(m.(string), mux))
	}()
	// TODO: this is a hack to make sure the server is ready
	// better would be to interlock the goroutine with the listener being ready
	time.Sleep(10 * time.Millisecond)
}

func createHandler(tag, s string) http.Handler {
	// TODO: hook worker in as HTTP handler
	// if _, ok := flow.Registry[s]; ok {
	// 	return http.Handler(reqHandler)
	// }
	if s == "<websocket>" {
		return websocket.Handler(wsHandler)
	}
	if strings.ContainsAny(s, "./") {
		h := http.FileServer(http.Dir(s))
		if s != "/" {
			h = http.StripPrefix(tag, h)
		}
		return h
	}
	panic("cannot create handler for: " + s)
}

func wsHandler(ws *websocket.Conn) {
	defer flow.DontPanic()
	defer ws.Close()

	tag := ws.Request().Header.Get("Sec-Websocket-Protocol")
	if tag == "" {
		tag = "default"
	}

	g := flow.NewGroup()
	g.AddWorker("head", &wsHead{ws: ws})
	g.Add("ws", "WebSocket-"+tag)
	g.AddWorker("tail", &wsTail{ws: ws})
	g.Connect("head.Out", "ws.In", 0)
	g.Connect("ws.Out", "tail.In", 0)
	g.Run()
}

type wsHead struct {
	flow.Work
	Out flow.Output

	ws *websocket.Conn
}

func (w *wsHead) Run() {
	for {
		var msg interface{}
		err := websocket.JSON.Receive(w.ws, &msg)
		if err == io.EOF {
			break
		}
		flow.Check(err)
		w.Out.Send(msg)
	}
}

type wsTail struct {
	flow.Work
	In flow.Input

	ws *websocket.Conn
}

func (w *wsTail) Run() {
	for m := range w.In {
		err := websocket.JSON.Send(w.ws, m)
		flow.Check(err)
	}
}
