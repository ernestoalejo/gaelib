package app

import (
	"bytes"

	"conf"
	"github.com/ernestokarim/gaelib/apperrors"
	"github.com/ernestokarim/gaelib/mail"

	"appengine"
)

func LogError(c appengine.Context, err error) {
	e, ok := (err).(*apperrors.Error)
	if !ok {
		e = Error(err).(*apperrors.Error)
	}

	c.Errorf("%s", e.Error())
	sendErrorByEmail(c, e.Error())
}

func Error(original error) error {
	return apperrors.New(original)
}

func Errorf(format string, args ...interface{}) error {
	return apperrors.Format(format, args...)
}

func NotFound() error {
	return apperrors.Code(404)
}

func Forbidden() error {
	return apperrors.Code(403)
}

func NotAllowed() error {
	return apperrors.Code(405)
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
			if err := Template(html, []string{"mails/error"}, data); err != nil {
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
