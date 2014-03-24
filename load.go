package flow

import (
	"encoding/json"
)

type config struct {
	Gadgets []struct {
		Type, Name string
	}
	Wires []struct {
		From, To string
		Capacity int
	}
	Feeds []struct {
		Tag  string
		Data interface{}
		To   string
	}
	Labels []struct {
		External, Internal string
	}
}

// Load a circuit from a JSON description in a string.
func (c *Circuit) LoadJSON(data []byte) error {
	var conf config
	err := json.Unmarshal(data, &conf)
	if err == nil {
		for _, g := range conf.Gadgets {
			c.Add(g.Name, g.Type)
		}
		for _, w := range conf.Wires {
			c.Connect(w.From, w.To, w.Capacity)
		}
		for _, f := range conf.Feeds {
			if f.Tag != "" {
				c.Feed(f.To, Tag{f.Tag, f.Data})
			} else {
				c.Feed(f.To, f.Data)
			}
		}
		for _, l := range conf.Labels {
			c.Label(l.External, l.Internal)
		}
	}
	return err
}
