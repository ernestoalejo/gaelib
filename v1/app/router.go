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
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/mjibson/appstats"
	"github.com/mjibson/goon"
)

var (
	dStore = gaesessions.NewDatastoreStore(model.KIND_SESSION,
			[]byte(conf.SESSION_SECRET))
	xsrfCodecs = securecookie.CodecsFromPairs([]byte(conf.XSRF_SECRET))
)

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
		}

		// Error handlers
		if parts[0] == "ERROR" {
			n, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				panic(err)
			}
			SetErrorHandler(int(n), handler)
			continue
		} 

		// Generalist handlers (no method specified)
		if len(parts[0]) == 0 {
			r.Handle(parts[1], h)
			continue
		}

		// Handlers for a concrete method
		r.Handle(parts[1], h).Methods(parts[0])
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

		// Build the request & session objects
		rw := newResponseWriter(w)
		r := &Request{Req: req, W: rw, C: c, N: goon.FromContext(c)}

		session, token, err := getSession(req, rw)
		if err != nil {
			r.processError(fmt.Errorf("build session failed: %s", err))
			return
		}
		r.Session = session

		// Check XSRF token
		if req.Method != "GET" {
			if ok, err := checkXsrfToken(req, token); err != nil {
				r.processError(fmt.Errorf("check xsrf token failed: %s", err))
				return
			} else if !ok {
				c.Errorf("xsrf token header check failed")
				r.processError(Forbidden())
				return
			}
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

// Return the session, the old XSRF token and an error if needed
func getSession(req *http.Request, w http.ResponseWriter) (*sessions.Session, []uint8, error) {
	session, _ := dStore.Get(req, conf.SESSION_NAME)
	session.Options = &sessions.Options{
		Path: "/",
		MaxAge: 7 * 24 * 60 * 60, // 7 days
	}

	var oldtoken []uint8
	if session.Values["xsrf"] != nil {
		oldtoken = session.Values["xsrf"].([]uint8)
	}
	token := securecookie.GenerateRandomKey(32)
	session.Values["xsrf"] = token

	encoded, err := securecookie.EncodeMulti("XSRF-TOKEN", token, xsrfCodecs...)
	if err != nil {
		return nil, nil, fmt.Errorf("encode token failed: %s", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "XSRF-TOKEN",
		Value: encoded,
		Path: "/",
	})
	return session, oldtoken, nil
}

// Returns true if the XSRF token was correct and an error if needed
func checkXsrfToken(req *http.Request, token []uint8) (bool, error) {
	c := appengine.NewContext(req)

	if token == nil {
		c.Errorf("[xsrf] token is nil")
		return false, nil
	}

	var cookie string
	for _, c := range req.Cookies() {
		if c.Name == "XSRF-TOKEN" {
			cookie = c.Value
		}
	}
	if cookie == "" {
		c.Errorf("[xsrf] no cookie")
		return false, nil
	}

	// Inconsistencies between the script & the cookies
	header := req.Header.Get("X-Xsrf-Token")
	if header != cookie {
		c.Errorf("[xsrf] inconsistency between the header & cookie: %s != %s",
			header, cookie)
		return false, nil
	}

	// Check the token itself
	var unsafeToken []uint8
	err := securecookie.DecodeMulti("XSRF-TOKEN", header, &unsafeToken, xsrfCodecs...)
	if err != nil {
		return false, fmt.Errorf("decode failed: %s", err)
	}
	if len(token) != len(unsafeToken) {
		c.Errorf("[xsrf] length check failed: %d != %d", len(token), len(unsafeToken))
		return false, nil
	}
	for i := range token {
		if token[i] != unsafeToken[i] {
			c.Errorf("[xsrf] character %d is not equal", i)
			return false, nil
		}
	}

	return true, nil
}
