// This is the core of the dataflow package.
/*
The flow package implements a dataflow mechanism in Go. It was greatly inspired
by Paul Morrison's Flow-based Programming (FBP) and Vladimir Sibirov's "goflow"
implementation - see also https://en.wikipedia.org/wiki/Flow-based_programming.

The flow library is available as import, along with some supporting packages:

    import "github.com/jcw/flow"
    import _ "github.com/jcw/flow/workers"

The "workers" package is loaded only for its side-effects here: defining some
basic workers in the registry.

To use it, start by creating a "group", then add "workers" and "connections":

    g := flow.NewGroup()
    g.Add("r", "Repeater")
    g.Add("c", "Counter")
    g.Connect("r.Out", "c.In", 0)

Then set a few initial values to send and start the whole thing up:

    g.Set("r.Num", 3)
    g.Set("r.In", "abc")
    g.Run()

Run will return once all workers have finished. The output will be lost since
the output port is not connected to anything, so it will show up as follows:

    Lost int: 3

A group can also be used as worker, then it acts as a "workgroup". For this,
a mapping from external port names to internal ones is needed to expose them:

    g.Map("MyOut", "c.out")

Once mapped, a group can be used like any other worker:

    g2 := flow.NewGroup()
    g2.AddWorker("g", g)
    g.Add("p", "Printer")
    g.Connect("g.MyOut", "p.In", 0)

Since the output has been wired up this time, the output will now be:

    int: 3

Definitions of workers, connections, and initial set requests can be loaded
from a JSON description. See Group.LoadFile() and Group.LoadString(), e.g.

    g.LoadFile("config.json")

Te define your own worker, create a type which embeds Work and defines Run():

    type LineLengths struct {
        flow.Work
        In  flow.Input
        Out flow.Output
    }

    func (w *LineLengths) Run() {
        for m := range w.In {
            s := m.(string)     // needs a type assertion
            w.Out.Send(len(s))
        }
    }

    g := flow.NewGroup()
    g.AddWorker("ll", new(LineLengths))
    g.Set("ll.In", "abc")
    g.Set("ll.In", "defgh")
    g.Run()

Inputs and outputs become available to the group in which this worker is used.

To make this worker available by name in the registry, set up a factory method:

    flow.registry["LineLengths"] = func() Worker {
        return new(LineLengths)
    }

For this simple case, a Transformer could also have been used:

    ll := flow.Transformer(func(m Memo) Memo) {
        return len(m.(string))
    }
    ...
    g.AddWorker("ll", ll)

This wraps a function into a worker with In and Out ports. It can be used when
there is a one-to-one processing task from incoming to outgoing memos.

Memo's are just a synonym for Go's "interface{}" type.
*/
package flow
