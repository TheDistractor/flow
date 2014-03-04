package database

import (
	"os"
	"testing"

	"github.com/jcw/flow/flow"
	_ "github.com/jcw/flow/workers"
)

const DB_PATH = "./test-db"

func init() {
	os.RemoveAll(DB_PATH)
}

// makeDb := func() flow.Worker {
// 	g := flow.NewGroup()
// 	g.Add("d", "LevelDB")
// 	g.Add("p", "Printer")
// 	g.Connect("d.Out", "p.In", 0)
// 	g.Map("Get", "d.Get")
// 	g.Map("Put", "d.Put")
// 	g.Map("Keys", "d.Keys")
// 	g.Set("db.Name", DB_PATH)
// 	return g
// }
// g.AddWorker("db", makeDb())

func ExampleLevelDB() {
	// TODO: clumsy, use workgroups and/or external channels

	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", DB_PATH)
	g.Set("db.Put", []string{"a/b", "123"})
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", DB_PATH)
	g.Set("db.Put", []string{"a/c", "456"})
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Add("p", "Printer")
	g.Connect("db.Out", "p.In", 0)
	g.Set("db.Name", DB_PATH)
	g.Set("db.Get", "a/b")
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Add("p", "Printer")
	g.Connect("db.Out", "p.In", 0)
	g.Set("db.Name", DB_PATH)
	g.Set("db.Keys", "a/")
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", DB_PATH)
	g.Set("db.Put", []string{"a/b"})
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Add("p", "Printer")
	g.Connect("db.Out", "p.In", 0)
	g.Set("db.Name", DB_PATH)
	g.Set("db.Get", "a/b")
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Add("p", "Printer")
	g.Connect("db.Out", "p.In", 0)
	g.Set("db.Name", DB_PATH)
	g.Set("db.Keys", "a/")
	g.Run()

	g = flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", DB_PATH)
	g.Set("db.Put", []string{"a/c"})
	g.Run()
	// Output:
	// string: 123
	// []string: [b c]
	// <nil>: <nil>
	// []string: [c]
}

func TestDatabase(t *testing.T) {
	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", DB_PATH)
	g.Run()
}
