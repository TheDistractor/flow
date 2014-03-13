package rfdata

import (
	"github.com/jcw/flow/flow"
)

func ExampleJeeBoot() {
	g := flow.NewGroup()
	g.Add("jb", "JeeBoot")
	g.Set("jb.In", []byte{
		224, 0, 2, 212, 17, 190, 240, 6, 48, 3,
		1, 196, 132, 97, 174, 237, 176, 147, 81, 6,
		25, 0, 245,
	})
	g.Set("jb.In", []byte{
		177, 0, 2, 1, 0, 17, 0, 99, 36,
	})
	g.Set("jb.In", []byte{
		177, 1, 0, 0, 0,
	})
	g.Run()
	// Output:
	// JB request 23
	// 11100000 &{0 2 D4 11 F0BE 06300301C48461AEEDB09351061900F5}
	// pair 06300301c48461aeedb09351061900f5 board 0 - no entry
	// JB request 9
	// 10110001 &{0 2 1 11 2463}
	// upgrade &{0 2 0 0 0} hdr 10110001
	// JB reply 0,2,0,0,0,0,0,0,0s
	// Lost *rfdata.upgradeRequest: &{0 2 0 0 0}
	// JB request 5
	// 10110001 &{1 0}
	// len 0 offset 0 64
	// no data at 0..64
	// JB reply 1,0,0s
	// Lost *struct { SwIDXor uint16 }: &{1}
}
