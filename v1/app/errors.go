package app

import (
	"github.com/ernestokarim/gaelib/v1/errors"

	"appengine"
	"appengine/taskqueue"
)

func NotFound() error {
	return errors.Code(404)
}

func Forbidden() error {
	return errors.Code(403)
}

func NotAllowed() error {
	return errors.Code(405)
}

func sendErrorByEmail(c appengine.Context, errorStr string) {
	t := NewTask("/tasks/error-mail", map[string]string{
		"Error": errorStr,
	})
	if _, err := taskqueue.Add(c, t, "admin-mails"); err != nil {
		c.Errorf("cannot prepare error mail: %s", err.Error())
	}
}
