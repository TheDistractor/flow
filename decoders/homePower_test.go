package decoders

import (
	"github.com/jcw/flow/flow"
)

func ExampleHomePower() {
	g := flow.NewGroup()
	g.Add("d", "Node-homePower")
	g.Set("d.In", "<Node-homePower>")
	g.Set("d.In", []byte{9, 213, 11, 68, 235, 151, 90, 99, 6, 88, 198, 136, 89})
	g.Set("d.In", "<Node-ookRelay>")
	g.Set("d.In", []byte{19, 115, 49, 78, 27, 37, 189, 117, 0})
	g.Set("d.In", "<Node-homePower>")
	g.Set("d.In", []byte{9, 213, 11, 68, 235, 153, 90, 84, 6, 88, 198, 136, 89})
	g.Run()
	// Output:
	// Lost string: <Node-homePower>
	// Lost map[string]int: map[<reading>:1 c1:3029 p1:78 c2:23191 p2:11009 c3:50776 p3:785]
	// Lost string: <Node-ookRelay>
	// Lost []uint8: [19 115 49 78 27 37 189 117 0]
	// Lost string: <Node-homePower>
	// Lost map[string]int: map[<reading>:1 c2:23193 p2:11111]
}
