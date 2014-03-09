// Interface to the LevelDB database.
package database

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/jcw/flow/flow"
	"github.com/syndtr/goleveldb/leveldb"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
)

func init() {
	flow.Registry["LevelDB"] = func() flow.Worker { return &LevelDB{} }
}

var (
	dbMutex sync.Mutex
	dbMap   = map[string]*openDb{}
)

type openDb struct {
	name string
	db   *leveldb.DB
	refs int
}

func (odb *openDb) Release() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	odb.refs--
	if odb.refs <= 0 {
		odb.db.Close()
		delete(dbMap, odb.name)
	}
}

func openDatabase(name string) *openDb {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	odb, ok := dbMap[name]
	if !ok {
		db, err := leveldb.OpenFile(name, nil)
		if err != nil {
			panic(err)
		}
		odb = &openDb{name, db, 0}
		dbMap[name] = odb
	}
	odb.refs++
	return odb
}

// LevelDB is a multi-purpose worker to get, put, and scan keys in a database.
// Expects flow.Tag values on its input port. Registers itself as "LevelDB".
type LevelDB struct {
	flow.Work
	Name flow.Input
	In   flow.Input
	Out  flow.Output

	odb *openDb
}

// Open the database and start listening to incoming get/put/keys requests.
func (w *LevelDB) Run() {
	if name, ok := <-w.Name; ok {
		w.odb = openDatabase(name.(string))
		defer w.odb.Release()

		for m := range w.In {
			if tag, ok := m.(flow.Tag); ok {
				switch tag.Tag {
				case "<get>":
					key := tag.Val.(string)
					w.Out.Send(m)
					w.Out.Send(w.get(key))
				case "<keys>":
					prefix := tag.Val.(string)
					w.Out.Send(m)
					w.Out.Send(w.keys(prefix))
				default:
					w.put(tag.Tag, tag.Val)
				}
			} else {
				w.Out.Send(m)
			}
		}
	}
}

func (w *LevelDB) get(key string) (any interface{}) {
	data, err := w.odb.db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &any)
	if err != nil {
		panic(err)
	}
	return
}

func (w *LevelDB) put(key string, value interface{}) {
	if value != nil {
		data, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		w.odb.db.Put([]byte(key), data, nil)
	} else {
		w.odb.db.Delete([]byte(key), nil)
	}
}

func (w *LevelDB) keys(prefix string) (results []string) {
	// TODO: decide whether this key logic is the most useful & least confusing
	// TODO: should use skips and reverse iterators once the db gets larger!
	skip := len(prefix)
	prev := "/" // impossible value, this never matches actual results

	w.iterateOverKeys(prefix, "", func(k string, v []byte) {
		i := strings.IndexRune(k[skip:], '/') + skip
		if i < skip {
			i = len(k)
		}
		if prev != k[skip:i] {
			// need to make a copy of the key, since it's owned by iter
			prev = k[skip:i]
			results = append(results, string(prev))
		}
	})
	return
}

func (w *LevelDB) iterateOverKeys(from, to string, fun func(string, []byte)) {
	slice := &dbutil.Range{[]byte(from), []byte(to)}
	if len(to) == 0 {
		slice.Limit = append(slice.Start, 0xFF)
	}

	iter := w.odb.db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		fun(string(iter.Key()), iter.Value())
	}
}
