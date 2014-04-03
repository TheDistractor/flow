package flow

import (
	"reflect"

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
	r := runner{desc: desc, fun: fun}
	// r.ins = map[string]*Input{ "In": new(Input) }
	return &r
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
