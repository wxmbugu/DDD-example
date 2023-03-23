package api

import (
	"github.com/patienttracker/internal/mailer"
)

type SendEmails struct {
	data     any
	mailer   mailer.Mailer
	email    string
	template string
	subject  string
}

func NewSenderMail() SendEmails {
	mail := mailer.NewMailer(25, "", "", "", "") //TODO:Don't send this to upstream with credentials
	return SendEmails{
		mailer: mail,
	}
}
func (s *SendEmails) setdata(data any, subject, template, receiveremail string) SendEmails {
	return SendEmails{
		data:     data,
		mailer:   s.mailer,
		email:    receiveremail,
		template: template,
		subject:  subject,
	}
}
func (s *SendEmails) Background() error {
	err := s.mailer.Send(s.email, s.subject, s.template, s.data)
	if err != nil {
		return err
	}
	return nil
}
