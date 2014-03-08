package rfdata

import (
	"github.com/jcw/flow/flow"
)

func ExampleRF12demo() {
	g := flow.NewGroup()
	g.Add("rf", "Sketch-RF12demo")
	g.Set("rf.In", "[RF12demo.12] _ i31* g5 @ 868 MHz c1 q1")
	g.Set("rf.In", "OK 9 187 176 69 235 249 6 192 234 6 74 190 18 (-66)")
	g.Set("rf.In", "OK 37 2 107 185 0 (-76)")
	g.Set("rf.In", "OK 197 (-60)")
	g.Run()
	// Output:
	// Lost map[string]int: map[<RF12demo>:12 band:868 group:5 id:31]
	// Lost map[string]int: map[<node>:9 rssi:-66]
	// Lost []uint8: [9 187 176 69 235 249 6 192 234 6 74 190 18]
	// Lost map[string]int: map[<node>:5 rssi:-76]
	// Lost []uint8: [37 2 107 185 0]
	// Lost map[string]int: map[<node>:5 rssi:-60]
	// Lost []uint8: [197]
}

func ExampleNodeMap() {
	g := flow.NewGroup()
	g.Add("nm", "NodeMap")
	g.Set("nm.Info", "RFg5i4 roomNode")
	g.Set("nm.In", map[string]int{"<RF12demo>": 1, "group": 5})
	g.Set("nm.In", map[string]int{"<node>": 3})
	g.Set("nm.In", map[string]int{"<node>": 4})
	g.Set("nm.In", map[string]int{"<node>": 5})
	g.Run()
	// Output:
	// Lost map[string]int: map[<RF12demo>:1 group:5]
	// Lost map[string]int: map[<node>:3]
	// Lost map[string]int: map[<node>:4]
	// Lost flow.Tag: {<dispatch> Node-roomNode}
	// Lost map[string]int: map[<node>:5]
}
