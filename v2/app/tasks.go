package app

import (
	"net/http"
	"net/url"

	"appengine/taskqueue"
)

func NewTask(path string, values map[string]string) *taskqueue.Task {
	headers := make(http.Header)
	headers.Set("X-AppEngine-FailFast", "1")
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	payload := []byte{}
	if values != nil {
		vals := url.Values{}
		for k, v := range values {
			vals[k] = []string{v}
		}
		payload = []byte(vals.Encode())
	}

	return &taskqueue.Task{
		Path:    path,
		Header:  headers,
		Payload: payload,
	}
}
