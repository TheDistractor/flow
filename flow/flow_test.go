package flow_test

import (
	"strings"
	"testing"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleNewGroup() {
	g := flow.NewGroup()
	g.Run()
}

func ExampleTransformer() {
	upper := flow.Transformer(func(m flow.Memo) flow.Memo {
		return strings.ToUpper(m.(string))
	})

	g := flow.NewGroup()
	g.AddWorker("u", upper)
	g.Set("u.In", "abc")
	g.Run()
	// Output:
	// Lost string: ABC
}

func ExampleGroup_Map() {
	// new workgroup to repeat each incoming memo three times
	wg := flow.NewGroup()
	wg.Add("r", "Repeater")
	wg.Set("r.Num", 3)
	wg.Map("MyIn", "r.In")
	wg.Map("MyOut", "r.Out")

	g := flow.NewGroup()
	g.AddWorker("wg", wg)
	g.Set("wg.MyIn", "abc")

	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: abc
	// Lost string: abc
}

func BenchmarkRepeat0(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 0)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepeat1(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 1)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepeat10(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 10)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepeat100(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("s", "Sink")
	g.Connect("r.Out", "s.In", 100)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe0(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 0)
	g.Connect("p.Out", "s.In", 0)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe1(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 1)
	g.Connect("p.Out", "s.In", 1)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe10(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 10)
	g.Connect("p.Out", "s.In", 10)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}

func BenchmarkRepPipe100(b *testing.B) {
	g := flow.NewGroup()
	g.Add("r", "Repeater")
	g.Add("p", "Pipe")
	g.Add("s", "Sink")
	g.Connect("r.Out", "p.In", 100)
	g.Connect("p.Out", "s.In", 100)
	g.Set("r.In", nil)
	g.Set("r.Num", b.N)
	g.Run()
}
