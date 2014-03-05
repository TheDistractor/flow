// Interface to the LevelDB database.
package database

import (
	"encoding/json"
	"strings"

	"github.com/jcw/flow/flow"
	"github.com/syndtr/goleveldb/leveldb"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
)

func init() {
	flow.Registry["LevelDB"] = func() flow.Worker { return &LevelDB{} }
}

type LevelDB struct {
	flow.Work
	Name flow.Input
	Get  flow.Input
	Put  flow.Input
	Keys flow.Input
	Out  flow.Output

	db *leveldb.DB
}

func (w *LevelDB) Run() {
	if name, ok := <-w.Name; ok {
		var err error
		w.db, err = leveldb.OpenFile(name.(string), nil)
		if err != nil {
			panic(err)
		}
		defer w.db.Close()

		active := 3
		for active > 0 {
			select {
			case m, ok := <-w.Get:
				if !ok {
					w.Get = nil
					active--
				} else {
					w.Out.Send(w.get(m.(string)))
				}
			case m, ok := <-w.Put:
				if !ok {
					w.Put = nil
					active--
				} else {
					args := m.([]string)
					if len(args) < 2 {
						w.put(args[0], nil)
					} else {
						w.put(args[0], args[1])
					}
				}
			case m, ok := <-w.Keys:
				if !ok {
					w.Keys = nil
					active--
				} else {
					w.Out.Send(w.keys(m.(string)))
				}
			}
		}
	}
}

func (w *LevelDB) get(key string) (any interface{}) {
	data, err := w.db.Get([]byte(key), nil)
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
		w.db.Put([]byte(key), data, nil)
	} else {
		w.db.Delete([]byte(key), nil)
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

	iter := w.db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		fun(string(iter.Key()), iter.Value())
	}
}
