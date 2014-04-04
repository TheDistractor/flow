package flow_test

import (
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

func BenchmarkRepeat0(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 0)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepeat1(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 1)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepeat10(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 10)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepeat100(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 100)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe0(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 0)
	g.Connect("p.Out", "s.In", 0)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe1(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 1)
	g.Connect("p.Out", "s.In", 1)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe10(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 10)
	g.Connect("p.Out", "s.In", 10)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe100(b *testing.B) {
	g := flow.NewCircuit()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 100)
	g.Connect("p.Out", "s.In", 100)
	g.Feed("r.In", nil)
	g.Feed("r.Num", b.N)
	g.Run()
}
