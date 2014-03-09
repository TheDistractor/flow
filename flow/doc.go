// This is the core of the dataflow package.
/*
The flow package implements a dataflow mechanism in Go. It was greatly inspired
by Paul Morrison's Flow-based Programming (FBP) and Vladimir Sibirov's "goflow"
implementation - see also https://en.wikipedia.org/wiki/Flow-based_programming.

The flow library is available as import, along with some supporting packages:

    import "github.com/jcw/flow/flow"
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

Run returns once all workers have finished. The output shows up as "lost" since
the output has not been connected:

    Lost int: 3

A group can be used instead as worker, then it acts as a "workgroup". For this,
a mapping from external port names to internal ones is needed to expose them:

    g.Map("MyOut", "c.out")

Once mapped, a group can be used like any other worker:

    g2 := flow.NewGroup()
    g2.AddWorker("g", g)
    g2.Add("p", "Printer")
    g2.Connect("g.MyOut", "p.In", 0)
    g2.Run()

Since the output port has been wired up this time, the output will now be:

    int: 3

Definitions of workers, connections, and initial set requests can be loaded
from a JSON description:

    data, _ := ioutil.ReadFile("config.json")
    g.LoadJSON(data)

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

For this simple case, a Transformer could also have been used:

    ll := flow.Transformer(func(m Memo) Memo) {
        return len(m.(string))
    }
    ...
    g.AddWorker("ll", ll)

This wraps a function into a worker with In and Out ports. It can be used when
there is a one-to-one processing task from incoming to outgoing memos.

To make a worker available by name in the registry, set up a factory method:

    flow.registry["LineLen"] = func() Worker {
        return new(LineLengths)
    }
    ...
    g.Add("ll", "LineLen")

Memo is a synonym for Go's generic "interface{}" type.
*/
package flow
