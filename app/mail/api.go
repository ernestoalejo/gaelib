package mail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"conf"

	"appengine"
	"appengine/urlfetch"
)

type Mail struct {
	To, ToName, From, FromName, Subject, Html string
}

// Response from the SendGrid API
type mailAPI struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// Send a mail using the SendGrid API
func SendMail(c appengine.Context, mail *Mail) error {
	client := &http.Client{
		Transport: &urlfetch.Transport{
			Context:  c,
			Deadline: time.Duration(20) * time.Second,
		},
	}

	// Build the data needed for the API call
	data := url.Values{
		"api_user": []string{conf.MAIL_API_USER},
		"api_key":  []string{conf.MAIL_API_KEY},
		"to":       []string{mail.To},
		"toname":   []string{mail.ToName},
		"subject":  []string{mail.Subject},
		"html":     []string{mail.Html},
		"from":     []string{mail.From},
		"fromname": []string{mail.FromName},
	}

	// Request the SendGrid API
	resp, err := client.PostForm(conf.MAIL_SEND_API, data)
	if err != nil {
		return fmt.Errorf("cannot fetch the sendgrid api: %s", err)
	}
	defer resp.Body.Close()

	// Decode the Mail API response
	var r mailAPI
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("cannot decode the api response: %s", err)
	}

	// Test for errors in the api call
	if r.Message != "success" {
		return fmt.Errorf("cannot send the mail: api %s message: %v", r.Message, r.Errors)
	}

	return nil
}
