// This application can exercise the "flow" package via a JSON config file.
package main

import (
	"os"

	"github.com/jcw/flow/flow"

	_ "github.com/jcw/flow/database"
	_ "github.com/jcw/flow/javascript"
	_ "github.com/jcw/flow/network"
	_ "github.com/jcw/flow/rfdata"
	_ "github.com/jcw/flow/serial"
	_ "github.com/jcw/flow/workers"
)

func main() {
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	g := flow.NewGroup()
	g.LoadFile(configFile)
	g.Run()
}
