package app

import (
	"net/http"
	"net/url"

	"appengine/taskqueue"
)

// Creates a new task and return it
func NewTask(path, name string, values url.Values) *taskqueue.Task {
	// Prepare the headers
	headers := make(http.Header)
	headers.Set("X-AppEngine-FailFast", "1")
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	// Encode the payload
	payload := []byte{}
	if values != nil {
		payload = []byte(values.Encode())
	}

	// Create the task
	return &taskqueue.Task{
		Path:    path,
		Header:  headers,
		Name:    name,
		Payload: payload,
	}
}
