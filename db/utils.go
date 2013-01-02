package db

import (
	"strconv"

	"github.com/ernestokarim/gaelib/app"

	"appengine"
	"appengine/datastore"
)

// Return a complete datastore key from a simple string id representing
// an int id. If the id can't be parsed, it returns a nil key.
func KeyFromIntId(c appengine.Context, kind, id string) (*datastore.Key, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		nerr := err.(*strconv.NumError)
		if nerr.Err == strconv.ErrSyntax {
			return nil, nil
		}

		return nil, app.Error(err)
	}

	return datastore.NewKey(c, kind, "", n, nil), nil
}
