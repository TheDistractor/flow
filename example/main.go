// This application exercises the "flow" package via a JSON config file.
// Use the "-i" flag for a list of built-in (i.e. pre-registered) workers.
package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/jcw/flow"
	_ "github.com/jcw/flow/gadgets"
)

var (
	verbose    = flag.Bool("i", false, "show info about version and registry")
	wait       = flag.Bool("k", false, "keep running, don't exit main")
	configFile = flag.String("w", "warmup.json", "specify the warmup file")
	appMain    = flag.String("m", "main", "which registered group to start")
)

func main() {
	flag.Parse()

	err := flow.AddToRegistry(*configFile)
	if err != nil && !*verbose {
		glog.Fatal(err)
	}

	if *verbose {
		println("Flow", flow.Version, "\n")
		flow.PrintRegistry()
		println("\nDocumentation at http://godoc.org/github.com/jcw/flow")
	} else {
		glog.Infof("Flow %s - starting, registry size %d",
			flow.Version, len(flow.Registry))
		if factory, ok := flow.Registry[*appMain]; ok {
			factory().Run()
			if *wait {
				time.Sleep(1e6 * time.Hour)
			}
		} else {
			glog.Fatalln(*appMain, "not found in:", *configFile)
		}
		glog.Infof("Flow %s -, normal exit", flow.Version)
	}
}
