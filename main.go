package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/serial"
	_ "github.com/jcw/flow/workers"
)

type config struct {
	Workers     []struct{ Type, Name string }
	Connections []struct{ From, To string }
	Requests    []struct{ Data, To string }
}

// Load a group from a JSON description in a file.
func LoadGroup(filename string) *flow.Group {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var conf config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		panic(err)
	}

	g := flow.NewGroup()
	for _, w := range conf.Workers {
		g.Add(w.Type, w.Name)
	}
	for _, c := range conf.Connections {
		g.Connect(c.From, c.To, 0)
	}
	for _, r := range conf.Requests {
		g.Request(r.Data, r.To)
	}

	return g
}

func main() {
	g := LoadGroup("config.json")
	g.Run()
}
