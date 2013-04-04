package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"io"
	"bytes"

	"conf"
	"server/model"

	"appengine"

	"github.com/gorilla/mux"
	gaesessions "code.google.com/p/sadbox/appengine/sessions"
	"github.com/gorilla/sessions"
	"github.com/mjibson/appstats"
	"github.com/mjibson/goon"
)

var dStore = gaesessions.NewDatastoreStore(model.KIND_SESSION, []byte(conf.SESSION_SECRET))

type Handler func(r *Request) error

// Build the router table at init().
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
	r.NotFoundHandler = appstatsWrapper(func(r *Request) error {
		return NotFound()
	})
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
		} else if len(parts[0]) == 0 {
			r.Handle(parts[1], h)
		} else {
			r.Handle(parts[1], h).Methods(parts[0])
		}
	}
}

type responseWriter struct {
	w http.ResponseWriter
	buf *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w: w, buf: bytes.NewBuffer(nil)}
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func (w *responseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

func (w *responseWriter) output() error {
	_, err := io.Copy(w.w, w.buf)
	return err
}

func appstatsWrapper(h Handler) http.Handler {
	f := func(c appengine.Context, w http.ResponseWriter, req *http.Request) {
		// Emit some compatibility & anti-cache headers for IE (you can overwrite
		// them from the handlers)
		w.Header().Set("X-UA-Compatible", "chrome=1")
		w.Header().Set("Cache-Control", "max-age=0,no-cache,no-store,"+
			"post-check=0,pre-check=0")
		w.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

		// Build the request object
		session, _ := dStore.Get(req, conf.SESSION_NAME)
		rw := newResponseWriter(w)
		r := &Request{
			Req: req,
			W:   rw,
			C:   c,
			N:   goon.FromContext(c),
			Session: session,
		}

		// Fatal errors recovery
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic recovered error: %s", rec)
				r.processError(err)
			}
		}()

		// Handle the request
		if err := h(r); err != nil {
			r.processError(err)
		}

		// Save the session & copy the buffered output
		if err := sessions.Save(req, w); err != nil {
			r.processError(err)
		}
		if err := rw.output(); err != nil {
			r.processError(err)
		}
	}
	return appstats.NewHandler(f)
}
