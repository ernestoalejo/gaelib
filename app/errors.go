package app

import (
	"bytes"

	"conf"
	"github.com/ernestokarim/gaelib/app/mail"

	"appengine"
)

func LogError(r *Request, err error) {
	r.internalServerError(err.Error())
	r.C.Errorf("%s", err)
	SendErrorByEmail(r.C, err.Error())
}

func SendErrorByEmail(c appengine.Context, errorStr string) {
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
