package flow

import (
	"encoding/json"
)

type config struct {
	Workers []struct {
		Type, Name string
	}
	Connections []struct {
		From, To string
		Buf      int
	}
	Requests []struct {
		Data, To string
	}
}

// Load a group from a JSON description in a string.
func (g *Group) LoadJSON(data []byte) error {
	var conf config
	err := json.Unmarshal(data, &conf)
	if err == nil {
		for _, w := range conf.Workers {
			g.Add(w.Name, w.Type)
		}
		for _, c := range conf.Connections {
			g.Connect(c.From, c.To, c.Buf)
		}
		for _, r := range conf.Requests {
			g.Set(r.To, r.Data)
		}
	}
	return err
}
