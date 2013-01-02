package db

import (
	"appengine/datastore"
)

type Model struct {
	// We don't really need this field; but we have it to be
	// a bit more encoding/gob friendly.
	Kind string

	key *datastore.Key
}

func (m *Model) Key() *datastore.Key {
	return m.key
}

func (m *Model) SetKey(key *datastore.Key) {
	m.key = key
	m.Kind = key.Kind()
}

func (m *Model) AddKey(key *datastore.Key) {
	if m.key == nil {
		m.key = key
		m.Kind = key.Kind()
	}
}
