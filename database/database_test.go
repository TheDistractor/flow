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

func ExampleLevelDB() {
	// TODO: clumsy, use external channels

	makeDb := func() flow.Worker {
		g := flow.NewGroup()
		g.Add("d", "LevelDB")
		g.Map("Get", "d.Get")
		g.Map("Put", "d.Put")
		g.Map("Keys", "d.Keys")
		g.Set("d.Name", dbPath)
		return g
	}

	g := flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Put", []string{"a/b", "123"})
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Put", []string{"a/c", "456"})
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Get", "a/b")
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Keys", "a/")
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Put", []string{"a/b"})
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Get", "a/b")
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Keys", "a/")
	g.Run()

	g = flow.NewGroup()
	g.AddWorker("db", makeDb())
	g.Set("db.Put", []string{"a/c"})
	g.Run()
	// Output:
	// Lost string: 123
	// Lost []string: [b c]
	// Lost <nil>: <nil>
	// Lost []string: [c]
}

func TestDatabase(t *testing.T) {
	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", dbPath)
	g.Run()
}
