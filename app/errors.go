package app

import (
	"bytes"
	"fmt"
	"runtime/debug"

	"conf"
	"github.com/ernestokarim/gaelib/mail"

	"appengine"
)

type AppError struct {
	CallStack   string
	OriginalErr error
	Code        int
}

func (err *AppError) Error() string {
	return fmt.Sprintf("[status code %d] %s\n\n%s", err.Code, err.OriginalErr, err.CallStack)
}

func (err *AppError) Log(c appengine.Context) {
	c.Errorf("%s", err.Error())
	sendErrorByEmail(c, err.Error())
}

func Error(original error) error {
	return &AppError{
		OriginalErr: original,
		Code:        500,
		CallStack:   fmt.Sprintf("%s", debug.Stack()),
	}
}

func NotFound() error {
	return &AppError{
		Code:      404,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}

func Forbidden() error {
	return &AppError{
		Code:      403,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}

func NotAllowed() error {
	return &AppError{
		Code:      405,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}

func sendErrorByEmail(c appengine.Context, errorStr string) {
	appid := appengine.AppID(c)

	// Try to send an email to the admin if the app is in production
	if !appengine.IsDevAppServer() {
		for _, admin := range conf.ADMIN_EMAILS {
			// Build the template data
			data := map[string]interface{}{
				"Error":    errorStr,
				"UserMail": admin,
				"AppId":    appid,
			}

			// Execute the template
			html := bytes.NewBuffer(nil)
			if err := RawExecuteTemplate(html, []string{"mails/error"}, data); err != nil {
				c.Errorf("cannot prepare an error email to the admin %s: %s", admin, err)
				continue
			}

			// Send the email to the admin
			m := &mail.Mail{
				To:       admin,
				ToName:   "Administrador",
				From:     "errors@" + appid + ".appspotmail.com",
				FromName: "Aviso de Errores",
				Subject:  "Se ha producido un error en la aplicación",
				Html:     string(html.Bytes()),
			}
			if err := mail.SendMail(c, m); err != nil {
				c.Errorf("cannot send an error email to the admin %s: %s", admin, err)
				continue
			}
		}
	}
}
