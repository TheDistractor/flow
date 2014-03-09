// This application can exercise the "flow" package via a JSON config file.
package main

import (
	"io/ioutil"
	"os"

	"github.com/jcw/flow/flow"

	_ "github.com/jcw/flow/database"
	_ "github.com/jcw/flow/decoders"
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
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	err = g.LoadJSON(data)
	if err != nil {
		panic(err)
	}
	g.Run()
}
