package mailer

import "log"

type Mailer struct{}

func NewMailer() Mailer {
	return Mailer{}
}

func (m Mailer) SendMail(s string) error {
	log.Printf("sending message: %s", s)
	return nil
}
