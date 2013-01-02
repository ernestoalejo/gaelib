package db

import (
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

func (m *Model) AddKey(key *datastore.Key) {
	if m.key == nil {
		m.key = key
	}
}
