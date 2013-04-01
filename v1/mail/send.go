package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"conf"

	"appengine"
	"appengine/taskqueue"
	"appengine/urlfetch"

	"github.com/ernestokarim/gaelib/v1/app"
)

type Mail struct {
	// Message info
	To, ToName,
	From, FromName,
	Subject string

	// Message body construction
	Templates []string
	Data      interface{}

	// Additional info for templates
	AppId string
}

func (m *Mail) Send(r *app.Request) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(m); err != nil {
		return fmt.Errorf("encode mail failed: %s", err)
	}

	t := app.NewTask("/tasks/mail", map[string]string{
		"Mail": buf.String(),
	})
	if _, err := taskqueue.Add(r.C, t, "mails"); err != nil {
		return fmt.Errorf("enqueue mail failed: %s", err)
	}

	return nil
}

// Response from the SendGrid API
type mailAPI struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// Send a mail using the SendGrid API
func SendGrid(r *app.Request, m *Mail) error {
	m.AppId = appengine.AppID(r.C)
	html := bytes.NewBuffer(nil)
	if err := app.Template(html, m.Templates, m); err != nil {
		return fmt.Errorf("prepare mail template failed: %s", err)
	}

	data := url.Values{
		"api_user": []string{conf.SENDGRID_USER},
		"api_key":  []string{conf.SENDGRID_KEY},
		"to":       []string{m.To},
		"toname":   []string{m.ToName},
		"subject":  []string{m.Subject},
		"html":     []string{html.String()},
		"from":     []string{m.From},
		"fromname": []string{m.FromName},
	}
	client := &http.Client{
		Transport: &urlfetch.Transport{
			Context:  r.C,
			Deadline: time.Duration(40) * time.Second,
		},
	}
	resp, err := client.PostForm(conf.SENDGRID_API, data)
	if err != nil {
		return fmt.Errorf("post mail failed: %s", err)
	}
	defer resp.Body.Close()

	var apiResp mailAPI
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("decode sendgrid response failed: %s", err)
	}
	if apiResp.Message != "success" {
		return fmt.Errorf("cannot send the mail: api %s message: %v",
			apiResp.Message, apiResp.Errors)
	}

	return nil
}
