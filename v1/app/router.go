package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"appengine"

	"github.com/gorilla/mux"
	"github.com/mjibson/appstats"
	"github.com/mjibson/goon"
)

type Handler func(r *Request) error

// Build the router table at init.
//
// Example routes map:
//    map[string]app.Handler{
//      "ERROR::404": stuff.NotFound,
//      "DELETE::/_/example": example.Delete,
//      ....
//      "::/_/feedback": stuff.Feedback,
//    }
//
func Router(routes map[string]Handler) {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)

	for route, handler := range routes {
		h := appstatsWrapper(handler)

		parts := strings.Split(route, "::")
		if len(parts) != 2 {
			panic("route not in the method::path format")
		} else if parts[0] == "ERROR" {
			n, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				panic(err)
			}

			SetErrorHandler(int(n), handler)

			if n == 404 {
				r.NotFoundHandler = h
			}
		} else if len(parts[0]) == 0 {
			r.Handle(parts[1], h)
		} else {
			r.Handle(parts[1], h).Methods(parts[0])
		}
	}
}

func appstatsWrapper(h Handler) http.Handler {
	f := func(c appengine.Context, w http.ResponseWriter, req *http.Request) {
		// Emit some compatibility & anti-cache headers for IE
		w.Header().Set("X-UA-Compatible", "chrome=1")
		w.Header().Set("Cache-Control", "max-age=0,no-cache,no-store,"+
			"post-check=0,pre-check=0")
		w.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

		r := &Request{
			Req: req,
			W:   w,
			C:   c,
			N:   goon.FromContext(c),
		}

		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic recovered error: %s", rec)
				r.processError(err)
			}
		}()

		if err := h(r); err != nil {
			r.processError(err)
		}
	}
	return appstats.NewHandler(f)
}
