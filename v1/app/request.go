package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"appengine"

	"github.com/gorilla/schema"
	"github.com/mjibson/goon"
	"github.com/gorilla/sessions"
)

var (
	schemaDecoder = schema.NewDecoder()
	errorHandlers = map[int]Handler{}
)

type Request struct {
	Req *http.Request
	W   http.ResponseWriter
	C   appengine.Context
	N   *goon.Goon
	Session *sessions.Session
}

// Load the request data using gorilla schema into a struct
func (r *Request) LoadData(data interface{}) error {
	if err := r.Req.ParseForm(); err != nil {
		return fmt.Errorf("parse form failed: %s", err)
	}

	if err := schemaDecoder.Decode(data, r.Req.Form); err != nil {
		e, ok := err.(schema.MultiError)
		if ok {
			// Delete the invalid path errors
			for k, v := range e {
				if strings.Contains(v.Error(), "schema: invalid path") {
					delete(e, k)
				}
			}

			// Return directly if there are no other kind of errors
			if len(e) == 0 {
				return nil
			}
		}

		// Not a MultiError, log it
		if err != nil {
			return fmt.Errorf("schema decode failed: %s", err)
		}
	}

	return nil
}

func (r *Request) LoadJsonData(data interface{}) error {
	if err := json.NewDecoder(r.Req.Body).Decode(data); err != nil {
		if err == io.EOF {
			return nil
		}

		return fmt.Errorf("decode json body failed: %s", err)
	}

	return nil
}

func (r *Request) EmitJson(data interface{}) error {
	// XSSI protection
	fmt.Fprintln(r.W, ")]}',")

	// Encode the output
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return fmt.Errorf("encode json failed: %s", err)
	}

	return nil
}

func (r *Request) IsPOST() bool {
	return r.Req.Method == "POST"
}

func (r *Request) IsDELETE() bool {
	return r.Req.Method == "DELETE"
}

func (r *Request) Path() string {
	u := r.Req.URL.Path
	query := r.Req.URL.RawQuery
	if len(query) > 0 {
		u += "?" + query
	}
	return u
}

// It returns a nil error always for easy of use inside the handlers.
// Example: return r.Redirect("/foo")
func (r *Request) Redirect(path string) error {
	http.Redirect(r.W, r.Req, path, http.StatusFound)
	return nil
}

// It returns a nil error always for easy of use inside the handlers.
// Example: return r.RedirectPermanently("/foo")
func (r *Request) RedirectPermanently(path string) error {
	http.Redirect(r.W, r.Req, path, http.StatusMovedPermanently)
	return nil
}

func (r *Request) Template(names []string, data interface{}) error {
	return Template(r.W, names, data)
}

func (r *Request) TemplateBase(names []string, data interface{}) error {
	dir := "templates"
	if appengine.IsDevAppServer() && r.Req.Header.Get("X-Request-From") == "cb" {
		dir = filepath.Join("client", "app")
	}

	return ExecTemplate(&TemplateConfig{
		LeftDelim:  `{%`,
		RightDelim: `%}`,
		Names:      names,
		W:          r.W,
		Data:       data,
		Dir:        dir,
	})
}

func (r *Request) TemplateDelims(names []string, data interface{}, leftDelim, rightDelim string) error {
	return TemplateDelims(r.W, names, data, leftDelim, rightDelim)
}

func (r *Request) URL() string {
	return r.Req.URL.String()
}

func (r *Request) LogError(err error) {
	r.C.Errorf("%v", err.Error())
	if !strings.Contains(r.URL(), "/tasks/error-mail") && !appengine.IsDevAppServer() {
		sendErrorByEmail(r.C, err.Error())
	}
}

func (r *Request) processError(err error) {
	code := 500
	if e, ok := err.(HttpError); ok {
		code = int(e)
		r.LogError(fmt.Errorf("http status code %s", e))
	} else {
		r.LogError(err)
	}

	h, ok := errorHandlers[code]
	if ok {
		if err := h(r); err == nil {
			return
		}
	}

	http.Error(r.W, "", code)
}

// Sets a new handler function for HTTP errors that returns the code status
func SetErrorHandler(code int, f Handler) {
	errorHandlers[code] = f
}
