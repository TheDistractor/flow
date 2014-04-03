package flow

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/textproto"
	"os"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

// A transformer processes each message through a supplied function.
func Transformer(fun func(Message) Message) Circuitry {
	return &transformer{fun: fun}
}

type transformer struct {
	Gadget
	In  Input
	Out Output

	fun func(Message) Message
}

func (g *transformer) Run() {
	for m := range g.In {
		// if m, ok := <-g.In; ok {
		g.Out.Send(g.fun(m))
	}
}

// TODO: work-in-progress

// A runner turns a function with I/O channels into a gadget.
func Runner(desc string, fun interface{}) Circuitry {
	h, t := parseDescription(desc)
	fmt.Fprintln(os.Stderr, 333, h)
	fmt.Fprintln(os.Stderr, 444, len(t), t)
	r := runner{desc: desc, fun: fun}
	// r.ins = map[string]*Input{ "In": new(Input) }
	return &r
}

// expects a mime-type header, followed by optional empty line and description
// TODO: title, pins, etc will be in the header, the long desc is in Markdown
func parseDescription(desc string) (header map[string][]string, text string) {
	b := bufio.NewReader(bytes.NewBufferString(desc + "\n\n"))
	header, err := textproto.NewReader(b).ReadMIMEHeader()
	Check(err)
	t, err := ioutil.ReadAll(b)
	Check(err)
	return header, strings.TrimSpace(string(t))
}

type runner struct {
	Gadget

	desc string
	fun  interface{}
	ins  map[string]*Input
	outs map[string]*Output
}

func (g *runner) pinValue(pin string) reflect.Value {
	println("lookupPin: " + pin)
	v := reflect.ValueOf(g.ins[pinPart(pin)])
	if !v.IsValid() {
		glog.Fatalln("pin not defined:", pin)
	}
	println(123)
	return v.Elem()
}

func (g *runner) Run() {
	// TODO: ...
}
