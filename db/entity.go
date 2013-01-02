package db

import (
	"reflect"

	"github.com/ernestokarim/gaelib/app"

	"appengine"
	"appengine/datastore"
)

type Entity interface {
	Key() *datastore.Key
	SetKey(*datastore.Key)
	AddKey(*datastore.Key)
}

func Kind(kind string) *datastore.Query {
	return datastore.NewQuery(kind)
}

func Query(c appengine.Context, q *datastore.Query, results interface{}) error {
	keys, err := q.GetAll(c, results)
	if err != nil {
		return app.Error(err)
	}

	v := reflect.ValueOf(results).Elem()
	for i, key := range keys {
		v.Index(i).Interface().(Entity).SetKey(key)
	}

	return nil
}

func Get(c appengine.Context, key *datastore.Key, result interface{}) (bool, error) {
	if err := datastore.Get(c, key, result); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, nil
		}
		return false, app.Error(err)
	}

	result.(Entity).SetKey(key)

	return true, nil
}

func Put(c appengine.Context, data interface{}) error {
	ent := data.(Entity)
	key, err := datastore.Put(c, ent.Key(), data)
	if err != nil {
		return app.Error(err)
	}
	ent.SetKey(key)

	return nil
}

func Delete(c appengine.Context, key *datastore.Key) error {
	if err := datastore.Delete(c, key); err != nil {
		return app.Error(err)
	}

	return nil
}
