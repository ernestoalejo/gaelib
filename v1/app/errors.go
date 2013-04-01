package app

import (
	"fmt"

	"appengine"
	"appengine/taskqueue"
)

type HttpError int

func (e HttpError) Error() string {
	return fmt.Sprintf("%d", e)
}

func Forbidden() error {
	return HttpError(403)
}

func NotFound() error {
	return HttpError(404)
}

func NotAllowed() error {
	return HttpError(405)
}

func sendErrorByEmail(c appengine.Context, errorStr string) {
	if appengine.IsDevAppServer() {
		return
	}

	t := NewTask("/tasks/error-mail", map[string]string{
		"Error": errorStr,
	})
	if _, err := taskqueue.Add(c, t, "admin-mails"); err != nil {
		c.Errorf("cannot prepare error mail: %s", err.Error())
	}
}
