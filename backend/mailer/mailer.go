package mailer

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	targetEmail  string
	fromEmail    string
	smtpHost     string
	smtpUsername string
	smtpPassword string
	smtpPort     int
}

func NewMailer() Mailer {

	targetEmail := os.Getenv("QN_TARGET_EMAIL")
	fromEmail := os.Getenv("QN_FROM_EMAIL")
	smtpHost := os.Getenv("QN_SMTP_HOST")
	smtpUsername := os.Getenv("QN_SMTP_USERNAME")
	smtpPassword := os.Getenv("QN_SMTP_PASSWORD")

	smtpPort, err := strconv.Atoi(os.Getenv("QN_SMTP_PORT"))
	if err != nil {
		logrus.Errorf("invalid smtp port provided (%s), default to zero", os.Getenv("QN_SMTP_PORT"))
		smtpPort = 0
	}

	return Mailer{
		targetEmail:  targetEmail,
		fromEmail:    fromEmail,
		smtpHost:     smtpHost,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		smtpPort:     smtpPort,
	}
}

func (m Mailer) SendMail(s string) error {
	// uses this: https://mailtrap.io/blog/golang-send-email/
	mailer := gomail.NewMessage()

	mailer.SetHeader("From", m.fromEmail)
	mailer.SetHeader("To", m.targetEmail)
	mailer.SetHeader("Subject", fmt.Sprintf("note: %s", s))
	mailer.SetBody("text/html", defineBody(s))
	// mailer.Attach("lolcat.jpg")

	d := gomail.NewDialer(m.smtpHost, m.smtpPort, m.smtpUsername, m.smtpPassword)
	d.SSL = false
	if err := d.DialAndSend(mailer); err != nil {
		return err
	}

	return nil
}

func defineBody(s string) string {
	body := fmt.Sprintf("<p><b>Subject:</b><br/>%s</p>", s)
	return body
}
