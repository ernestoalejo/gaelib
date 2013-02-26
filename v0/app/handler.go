package app

import (
	"net/http"

	"appengine"

	"github.com/ernestokarim/gaelib/v0/errors"
)

// All handlers in the app must implement this type
type Handler func(r *Request) error

// Serves a http request
func (fn Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)

	// Emit some compatibility anti-cache headers for IE
	w.Header().Set("X-UA-Compatible", "chrome=1")
	w.Header().Set("Cache-Control", "max-age=0,no-cache,no-store,post-check=0,pre-check=0")
	w.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

	r := &Request{Req: req, W: w, C: c}

	defer func() {
		if rec := recover(); rec != nil {
			err := errors.Format("panic recovered error: %s", rec)
			r.processError(err)
		}
	}()

	if err := fn(r); err != nil {
		r.processError(err)
	}
}
