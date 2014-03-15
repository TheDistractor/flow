package flow_test

import (
	"strings"
	"testing"

	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
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
	g.Set("u.In", "def")
	g.Run()
	// Output:
	// Lost string: ABC
	// Lost string: DEF
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
	g.Set("wg.MyIn", "def")

	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: abc
	// Lost string: abc
	// Lost string: def
	// Lost string: def
	// Lost string: def
}

func ExampleNestedGroup() {
	g1 := flow.NewGroup()
	g1.Add("p", "Pipe")
	g1.Map("In", "p.In")
	g1.Map("Out", "p.Out")

	g2 := flow.NewGroup()
	g2.Add("p", "Pipe")
	g2.Map("In", "p.In")
	g2.Map("Out", "p.Out")

	g3 := flow.NewGroup()
	g3.AddWorker("g1", g1)
	g3.AddWorker("g2", g2)
	g3.Connect("g1.Out", "g2.In", 0)
	g3.Map("In", "g1.In")
	g3.Map("Out", "g2.Out")

	g := flow.NewGroup()
	g.Add("p1", "Pipe")
	g.AddWorker("g", g3)
	g.Add("p2", "Pipe")
	g.Connect("p1.Out", "g.In", 0)
	g.Connect("g.Out", "p2.In", 0)
	g.Set("p1.In", "abc")
	g.Set("p1.In", "def")
	g.Set("p1.In", "ghi")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: def
	// Lost string: ghi
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
