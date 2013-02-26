package app

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Build the router table at init.
//
// Example routes map:
//    map[string]app.Handler{
//      "ERROR::404": stuff.NotFound,
//      "DELETE::/_/example": example.Delete,
//      ....
//    }
//
func Router(routes map[string]Handler) {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)

	for route, handler := range routes {
		parts := strings.Split(route, "::")
		if len(parts) != 2 {
			panic("route not in the method::path format")
		}

		if parts[0] == "ERROR" {
			n, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				panic(err)
			}

			SetErrorHandler(int(n), Handler(handler))

			if n == 404 {
				r.NotFoundHandler = Handler(handler)
			}
		}

		if len(parts[0]) == 0 {
			r.Handle(parts[1], Handler(handler))
		} else {
			r.Handle(parts[1], Handler(handler)).Methods(parts[0])
		}
	}
}
