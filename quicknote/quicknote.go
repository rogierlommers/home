package quicknote

import (
	"fmt"
	"os"
	"path/filepath"
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

// sendMail sends an email
// filename is a string with the filename
// attachment is a []byte
func sendMail(filename string, attachment []byte) error {

	// safe file to tmp location
	tmpFilename := fmt.Sprintf("/tmp/%s", filename)
	err := os.WriteFile(tmpFilename, attachment, 0777)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFilename)

	// initialise mailer. Uses this: https://mailtrap.io/blog/golang-send-email/
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", quickNote.fromEmail)
	mailer.SetHeader("To", quickNote.targetEmail)

	// first determine attachment
	fileExtension := filepath.Ext(filename)
	var subject = ""

	switch fileExtension {
	case ".txt":
		// we need to extract the text out of the text file
		contents, err := os.ReadFile(tmpFilename)
		if err != nil {
			logrus.Error(err)
		}
		subject = fmt.Sprintf("Todo item: %s", contents)
	default:
		subject = fmt.Sprintf("Todo item: %s", filename)
		mailer.Attach(tmpFilename)
	}

	// actual send mail
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", defineBody(subject))

	d := gomail.NewDialer(quickNote.smtpHost, quickNote.smtpPort, quickNote.smtpUsername, quickNote.smtpPassword)
	d.SSL = false
	if err := d.DialAndSend(mailer); err != nil {
		return err
	}

	logrus.Infof("Email succesfully sent")
	return nil
}

func defineBody(s string) string {
	body := fmt.Sprintf("<p><b>Subject:</b><br/>%s</p>", s)
	return body
}
