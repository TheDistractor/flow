// This application exercises the "flow" package via a JSON config file.
// Use the "-v" flag for a list of built-in (i.e. pre-registered) workers.
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"sort"
	"time"

	"github.com/jcw/flow/flow"

	_ "github.com/jcw/flow/database"
	_ "github.com/jcw/flow/decoders"
	_ "github.com/jcw/flow/fbpparse"
	_ "github.com/jcw/flow/javascript"
	_ "github.com/jcw/flow/network"
	_ "github.com/jcw/flow/rfdata"
	_ "github.com/jcw/flow/serial"
	_ "github.com/jcw/flow/workers"
)

var (
	verbose    = flag.Bool("v", false, "show version and registry contents")
	wait       = flag.Bool("w", false, "wait forever, don't exit main")
	configFile = flag.String("c", "config.json", "use configuration file")
	appMain    = flag.String("r", "main", "which registered group to run")
)

func main() {
	defer flow.DontPanic()
	flag.Parse()

	data, err := ioutil.ReadFile(*configFile)
	flow.Check(err)

	var definitions map[string]json.RawMessage
	err = json.Unmarshal(data, &definitions)
	flow.Check(err)

	for name, def := range definitions {
		registerGroup(name, def)
	}

	if *verbose {
		println("Flow " + flow.Version + "\n")
		printRegistry()
		println("\nDocumentation at http://godoc.org/github.com/jcw/flow")
	} else {
		if factory, ok := flow.Registry[*appMain]; ok {
			factory().Run()
			if *wait {
				time.Sleep(1e6 * time.Hour)
			}
		} else {
			panic(*appMain + " not found in: " + *configFile)
		}
	}
}

func registerGroup(name string, def []byte) {
	flow.Registry[name] = func() flow.Worker {
		g := flow.NewGroup()
		err := g.LoadJSON(def)
		flow.Check(err)
		return g
	}
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
