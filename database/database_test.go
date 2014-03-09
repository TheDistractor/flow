package database

import (
	"os"
	"path"
	"testing"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

var dbPath = path.Join(os.TempDir(), "flow-test-db")

func init() {
	println(dbPath)
}

type dbFeed struct {
	flow.Work
	ToGet, ToPut, ToKeys flow.Output
}

func (w *dbFeed) Run() {
	w.ToPut.Send([]string{"a/b", "123"})
	w.ToPut.Send([]string{"a/c", "456"})
	w.ToGet.Send("a/b")
	w.ToKeys.Send("a/")
	w.ToPut.Send([]string{"a/b"})
	w.ToGet.Send("a/b")
	w.ToKeys.Send("a/")
	w.ToPut.Send([]string{"a/c"})
}

func ExampleLevelDB() {
	// type dbFeed struct {
	//     flow.Work
	//     ToGet, ToPut, ToKeys flow.Output
	// }
	//
	// func (w *dbFeed) Run() {
	//     w.ToPut.Send([]string{"a/b", "123"})
	//     w.ToPut.Send([]string{"a/c", "456"})
	//     w.ToGet.Send("a/b")
	//     w.ToKeys.Send("a/")
	//     w.ToPut.Send([]string{"a/b"})
	//     w.ToGet.Send("a/b")
	//     w.ToKeys.Send("a/")
	//     w.ToPut.Send([]string{"a/c"})
	// }

	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.AddWorker("feed", &dbFeed{})
	g.Connect("feed.ToPut", "db.Put", 0)
	g.Connect("feed.ToGet", "db.Get", 0)
	g.Connect("feed.ToKeys", "db.Keys", 0)
	g.Set("db.Name", dbPath)
	g.Run()
	// Output:
	// Lost flow.Tag: {a/b 123}
	// Lost []string: [b c]
	// Lost flow.Tag: {a/b <nil>}
	// Lost []string: [c]
}

func TestDatabase(t *testing.T) {
	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", dbPath)
	g.Run()
}
