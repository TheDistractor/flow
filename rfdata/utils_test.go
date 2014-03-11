package rfdata

import (
	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

func ExampleCalcCrc16() {
	g := flow.NewGroup()
	g.Add("c", "CalcCrc16")
	g.Set("c.In", []byte("abc"))
	g.Run()
	// Output:
	// Lost []uint8: [97 98 99]
	// Lost flow.Tag: {<crc16> 22345}
}

func ExampleReadTextFile() {
	g := flow.NewGroup()
	g.Add("r", "ReadTextFile")
	g.Add("c", "Counter")
	g.Connect("r.Out", "c.In", 0)
	g.Set("r.In", "./blinkAvr1.hex")
	g.Run()
	// Output:
	// Lost flow.Tag: {<open> ./blinkAvr1.hex}
	// Lost flow.Tag: {<close> ./blinkAvr1.hex}
	// Lost int: 47
}

func ExampleIntelHexToBin() {
	g := flow.NewGroup()
	g.Add("r", "ReadTextFile")
	g.Add("b", "IntelHexToBin")
	g.AddWorker("n", flow.Transformer(func(m flow.Memo) flow.Memo {
		if v, ok := m.([]byte); ok {
			m = len(v)
		}
		return m
	}))
	g.Connect("r.Out", "b.In", 0)
	g.Connect("b.Out", "n.In", 0)
	g.Set("r.In", "./blinkAvr1.hex")
	g.Run()
	// Output:
	// Lost flow.Tag: {<open> ./blinkAvr1.hex}
	// Lost flow.Tag: {<addr> 0}
	// Lost int: 726
	// Lost flow.Tag: {<close> ./blinkAvr1.hex}
}

func ExampleBinaryFill() {
	g := flow.NewGroup()
	g.Add("f", "BinaryFill")
	g.Set("f.In", []byte("abcdef"))
	g.Set("f.Len", 5)
	g.Run()
	// Output:
	// Lost []uint8: [97 98 99 100 101 102 255 255 255 255]
}
