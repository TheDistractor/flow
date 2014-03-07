package flow

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	Workers     []struct{ Type, Name string }
	Connections []struct{ From, To string }
	Requests    []struct{ Data, To string }
}

// Load a group from a JSON description in a string.
func (g *Group) LoadString(s string) error {
	var conf config
	err := json.Unmarshal([]byte(s), &conf)
	if err == nil {
		for _, w := range conf.Workers {
			g.Add(w.Name, w.Type)
		}
		for _, c := range conf.Connections {
			g.Connect(c.From, c.To, 0)
		}
		for _, r := range conf.Requests {
			g.Set(r.To, r.Data)
		}
	}
	return err
}

// Load a group from a JSON description in a file.
func (g *Group) LoadFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = g.LoadString(string(data))
	}
	return err
}
