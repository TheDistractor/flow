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
	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", dbPath)
	g.Set("db.In", flow.Tag{"a/b", "123"})
	g.Set("db.In", flow.Tag{"a/c", "456"})
	g.Set("db.In", flow.Tag{"<get>", "a/b"})
	g.Set("db.In", flow.Tag{"<keys>", "a/"})
	g.Set("db.In", flow.Tag{"a/b", nil})
	g.Set("db.In", flow.Tag{"<get>", "a/b"})
	g.Set("db.In", flow.Tag{"<keys>", "a/"})
	g.Set("db.In", flow.Tag{"a/c", nil})
	g.Run()
	// Output:
	// Lost flow.Tag: {<get> a/b}
	// Lost string: 123
	// Lost flow.Tag: {<keys> a/}
	// Lost []string: [b c]
	// Lost flow.Tag: {<get> a/b}
	// Lost <nil>: <nil>
	// Lost flow.Tag: {<keys> a/}
	// Lost []string: [c]
}

func TestDatabase(t *testing.T) {
	g := flow.NewGroup()
	g.Add("db", "LevelDB")
	g.Set("db.Name", dbPath)
	g.Run()
}
