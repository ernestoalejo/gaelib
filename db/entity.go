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
