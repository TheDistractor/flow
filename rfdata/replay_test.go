package rfdata

import (
	"fmt"

	"github.com/jcw/flow/flow"
)

type MockOutput int

func (c *MockOutput) Send(m flow.Memo) {
	if *c < 9 {
		fmt.Printf("%T: %v\n", m, m)
		(*c)++
	}
}

func (c *MockOutput) Close() {}

func ExampleLogReader() {
	lr := new(LogReader)
	name := make(chan flow.Memo, 1)
	lr.Name, lr.Out = name, new(MockOutput)
	name <- "./20121130.txt.gz"
	lr.Run()
	// Output:
	// time.Time: 2012-11-30 00:00:00.062 +0000 UTC
	// string: <usb-A40117UK>
	// string: OK 19 96 12 11 30 1 0 0
	// time.Time: 2012-11-30 00:00:00.101 +0000 UTC
	// string: <usb-A40117UK>
	// string: DF S 3129 63 222769
	// time.Time: 2012-11-30 00:00:02.67 +0000 UTC
	// string: <usb-A40117UK>
	// string: OK 9 14 11 67 235 163 65 28 235 141 166 77 40
}
