package mailer

import (
	"bytes"
	templates "github.com/patienttracker/template"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dial   *gomail.Dialer
	sender string
	tmp    *templates.Template
}

func NewMailer(port int, sender, host, username, password string) Mailer {
	dialer := gomail.NewDialer(host, port, username, password)
	tmp := templates.New()
	return Mailer{sender: sender, dial: dialer, tmp: tmp}
}

func (mail *Mailer) Send(recipient string, subject, template string, data interface{}) error {
	htmlBody := new(bytes.Buffer)
	err := mail.tmp.Render(htmlBody, template, data)
	if err != nil {
		return err
	}
	m := gomail.NewMessage()
	m.SetHeader("From", mail.sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody.String())
	if err := mail.dial.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
