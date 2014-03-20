// This is the core of the dataflow package.
/*
The flow package implements a dataflow mechanism in Go. It was greatly inspired
by Paul Morrison's Flow-based Programming (FBP) and Vladimir Sibirov's "goflow"
implementation - see also https://en.wikipedia.org/wiki/Flow-based_programming.

The flow library is available as import, along with some supporting packages:

    import "github.com/jcw/flow"
    import _ "github.com/jcw/flow/gadgets"

The "gadgets" package is loaded only for its side-effects here: defining some
basic gadgets in the registry.

To use it, start by creating a "circuit", then add "gadgets" and "wires":

    g := flow.NewCircuit()
    g.Add("r", "Repeater")
    g.Add("c", "Counter")
    g.Connect("r.Out", "c.In", 0)

Then set a few initial values to send and start the whole thing up:

    g.Feed("r.Num", 3)
    g.Feed("r.In", "abc")
    g.Run()

Run returns once all gadgets have finished. Output shows up as "lost" since the
output hasn't been connected:

    Lost int: 3

A circuit can also be used as gadget, collectively called "circuitry". For this,
internal pins must be labeled with external names to expose them:

    g.Label("MyOut", "c.out")

Once pins have been labeled, the circuit can be used inside another circuit:

    g2 := flow.NewCircuit()
    g2.AddCircuitry("g", g)
    g2.Add("p", "Printer")
    g2.Connect("g.MyOut", "p.In", 0)
    g2.Run()

Since the output pin has been wired up this time, the output will now be:

    3

Definitions of gadgets, wires, and initial set requests can be loaded
from a JSON description:

    data, _ := ioutil.ReadFile("config.json")
    g := flow.NewCircuit()
    g.LoadJSON(data)
    g.Run()

Te define your own gadget, create a type which embeds Gadget and defines Run():

    type LineLengths struct {
        flow.Gadget
        In  flow.Input
        Out flow.Output
    }

    func (w *LineLengths) Run() {
        for m := range w.In {
            s := m.(string)     // needs a type assertion
            w.Out.Send(len(s))
        }
    }

    g := flow.NewCircuit()
    g.AddCircuitry("ll", new(LineLengths))
    g.Feed("ll.In", "abc")
    g.Feed("ll.In", "defgh")
    g.Run()

Inputs and outputs become available to the circuit in which this gadget is used.

For this simple case, a Transformer could also have been used:

    ll := flow.Transformer(func(m Message) Message) {
        return len(m.(string))
    }
    ...
    g.AddCircuitry("ll", ll)

This wraps a function into a gadget with In and Out pins. It can be used when
there is a one-to-one processing task from incoming to outgoing messages.

To make a gadget available by name in the registry, set up a factory method:

    flow.registry["LineLen"] = func() GadgetType {
        return new(LineLengths)
    }
    ...
    g.Add("ll", "LineLen")

Message is a synonym for Go's generic "interface{}" type.
*/
package flow
