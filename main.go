// This application can exercise the "flow" package via a JSON config file.
// Use the "-v" flag for a list of built-in (i.e. pre-registered) workers.
package main

import (
	"io/ioutil"
	"os"
	"sort"

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
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		println("Flow " + flow.Version + "\n")
		printRegistry()
		println("\nDocumentation at http://godoc.org/github.com/jcw/flow")
		os.Exit(0)
	}

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

func printRegistry() {
	keys := []string{}
	for k := range flow.Registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := " "
	for _, k := range keys {
		if len(s)+len(k) > 78 {
			println(s)
			s = " "
		}
		s += " " + k
	}
	if len(s) > 1 {
		println(s)
	}
}
