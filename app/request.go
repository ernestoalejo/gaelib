package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"code.google.com/p/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()

type Request struct {
	Req *http.Request
	W   http.ResponseWriter
	C   appengine.Context
}

// Load the request data using gorilla schema into a struct
func (r *Request) LoadData(data interface{}) error {
	if err := r.Req.ParseForm(); err != nil {
		return fmt.Errorf("error parsing the request form: %s", err)
	}

	if err := schemaDecoder.Decode(data, r.Req.Form); err != nil {
		return fmt.Errorf("error decoding the schema: %s", err)
	}

	return nil
}

func (r *Request) LoadJsonData(data interface{}) error {
	if err := json.NewDecoder(r.Req.Body).Decode(data); err != nil {
		return fmt.Errorf("error decoding the json: %s", err)
	}

	return nil
}

func (r *Request) EmitJson(data interface{}) error {
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return fmt.Errorf("error encoding the json: %s", err)
	}

	return nil
}

func (r *Request) IsPOST() bool {
	return r.Req.Method == "POST"
}

func (r *Request) Path() string {
	return r.Req.URL.Path + "?" + r.Req.URL.RawQuery
}

// It returns a nil error always for easy of use inside the handlers.
// Example: return r.Redirect("/foo")
func (r *Request) Redirect(path string) error {
	http.Redirect(r.W, r.Req, path, http.StatusFound)
	return nil
}

func (r *Request) NotFound(message string) error {
	http.Error(r.W, message, http.StatusNotFound)
	return nil
}

func (r *Request) Forbidden(message string) error {
	http.Error(r.W, message, http.StatusForbidden)
	return nil
}

func (r *Request) ExecuteTemplate(names []string, data interface{}) error {
	return RawExecuteTemplate(r.W, names, data)
}

// You shouldn't use this function, but directly return
// an error from the handler.
func (r *Request) internalServerError(message string) error {
	http.Error(r.W, message, http.StatusInternalServerError)
	return nil
}

func (r *Request) JsonResponse(data interface{}) error {
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return fmt.Errorf("cannot serialize the response: %s", err)
	}
	return nil
}
