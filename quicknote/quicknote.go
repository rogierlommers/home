package quicknote

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var quickNote = Quicknote{}

type Quicknote struct {
	targetEmail  string
	fromEmail    string
	smtpHost     string
	smtpUsername string
	smtpPassword string
	smtpPort     int
}

func NewQuicknote(router *gin.Engine, cfg config.AppConfig) {

	// get all environment vars
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

	// add endpoints
	router.POST("/api/notes/send", sendMailHandler)

	// initiale package state
	quickNote = Quicknote{
		targetEmail:  targetEmail,
		fromEmail:    fromEmail,
		smtpHost:     smtpHost,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		smtpPort:     smtpPort,
	}

}

func sendMail(s string) error {
	// uses this: https://mailtrap.io/blog/golang-send-email/
	mailer := gomail.NewMessage()

	mailer.SetHeader("From", quickNote.fromEmail)
	mailer.SetHeader("To", quickNote.targetEmail)
	mailer.SetHeader("Subject", fmt.Sprintf("note: %s", s))
	mailer.SetBody("text/html", defineBody(s))
	// mailer.Attach("lolcat.jpg")

	d := gomail.NewDialer(quickNote.smtpHost, quickNote.smtpPort, quickNote.smtpUsername, quickNote.smtpPassword)
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
