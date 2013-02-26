package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"appengine"

	"github.com/ernestokarim/gaelib/v1/errors"
	"github.com/gorilla/schema"
)

var (
	schemaDecoder = schema.NewDecoder()

	errorHandlers = map[int]Handler{}
)

type Request struct {
	Req *http.Request
	W   http.ResponseWriter
	C   appengine.Context
}

// Load the request data using gorilla schema into a struct
func (r *Request) LoadData(data interface{}) error {
	if err := r.Req.ParseForm(); err != nil {
		return errors.New(err)
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
			return errors.New(err)
		}
	}

	return nil
}

func (r *Request) LoadJsonData(data interface{}) error {
	if err := json.NewDecoder(r.Req.Body).Decode(data); err != nil {
		if err == io.EOF {
			return nil
		}

		return errors.New(err)
	}

	return nil
}

func (r *Request) EmitJson(data interface{}) error {
	// XSSI protection
	fmt.Fprintln(r.W, ")]}',")

	// Encode the output
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return errors.New(err)
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
	if r.Req.Header.Get("X-Request-From") == "cb" {
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

func (r *Request) JsonResponse(data interface{}) error {
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return errors.New(err)
	}
	return nil
}

func (r *Request) processError(err error) {
	e := errors.New(err).(*errors.Error)
	LogError(r.C, e)

	h, ok := errorHandlers[e.Code]
	if ok {
		if err := h(r); err == nil {
			return
		}
	}

	http.Error(r.W, "", e.Code)
}

func SetErrorHandler(code int, f Handler) {
	errorHandlers[code] = f
}
