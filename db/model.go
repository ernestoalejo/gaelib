package db

import (
	"github.com/ernestokarim/gaelib/app"

	"appengine"
	"appengine/datastore"
)

type Model struct {
	key *datastore.Key
}

func (m *Model) Key() *datastore.Key {
	return m.key
}

func (m *Model) SetKey(key *datastore.Key) {
	m.key = key
}

func (m *Model) GetOrDefault(c appengine.Context, result interface{}) error {
	if m.key == nil {
		panic("no key to get the model")
	}

	if err := datastore.Get(c, m.key, result); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil
		}
		return app.Error(err)
	}

	return nil
}

func (m *Model) Save(c appengine.Context, data interface{}) error {
	key, err := datastore.Put(c, m.key, data)
	if err != nil {
		return app.Error(err)
	}

	m.key = key

	return nil
}

func (m *Model) Delete(c appengine.Context) error {
	if err := datastore.Delete(c, m.key); err != nil {
		return app.Error(err)
	}

	return nil
}
