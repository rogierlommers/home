package quicknote

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var quickNote = Quicknote{}

type Quicknote struct {
	targetEmailPrivate string
	targetEmailWork    string
	fromEmail          string
	smtpHost           string
	smtpUsername       string
	smtpPassword       string
	smtpPort           int
}

func NewQuicknote(router *gin.Engine, cfg config.AppConfig) {

	// get all environment vars
	targetEmailPrivate := os.Getenv("QN_TARGET_EMAIL_PRIVATE")
	targetEmailWork := os.Getenv("QN_TARGET_EMAIL_WORK")
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
		targetEmailPrivate: targetEmailPrivate,
		targetEmailWork:    targetEmailWork,
		fromEmail:          fromEmail,
		smtpHost:           smtpHost,
		smtpUsername:       smtpUsername,
		smtpPassword:       smtpPassword,
		smtpPort:           smtpPort,
	}

}

// sendMail sends an email
func sendMail(filename string, attachment []byte, optionalText string, fileAttached bool) error {

	var target string

	switch fileAttached {

	case true:
		// determine target email based on subjectFilename (attachment)
		if strings.HasPrefix(filename, "w ") {
			target = quickNote.targetEmailWork
		} else {
			target = quickNote.targetEmailPrivate
		}

	case false:
		// determine target email based on text
		if strings.HasPrefix(optionalText, "w ") {
			target = quickNote.targetEmailWork
			optionalText = strings.TrimPrefix(optionalText, "w ")
		} else {
			target = quickNote.targetEmailPrivate
		}
	}

	logrus.Debugf("using address: %s", target)

	// initialise mailer. Uses this: https://mailtrap.io/blog/golang-send-email/
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", quickNote.fromEmail)
	mailer.SetHeader("To", target)

	if fileAttached {
		// safe file to tmp location
		tmpFilename := fmt.Sprintf("/tmp/%s", filename)
		err := os.WriteFile(tmpFilename, attachment, 0777)
		if err != nil {
			return err
		}
		defer os.Remove(tmpFilename)

		// first determine attachment
		fileExtension := filepath.Ext(filename)
		if fileExtension != ".txt" {
			mailer.Attach(tmpFilename)
		}
	}

	// continue sending mail
	var subject, body string

	subject = getFirstLine(optionalText)
	body = optionalText

	// actual send mail
	mailer.SetHeader("Subject", fmt.Sprintf("☑️ %s", subject))
	mailer.SetBody("text/html", defineBody(subject, body))

	d := gomail.NewDialer(quickNote.smtpHost, quickNote.smtpPort, quickNote.smtpUsername, quickNote.smtpPassword)
	d.SSL = false
	if err := d.DialAndSend(mailer); err != nil {
		return err
	}

	logrus.Debugf("Email succesfully sent")
	return nil
}

func defineBody(s string, b string) string {
	b = strings.ReplaceAll(b, "\n", "<br/>")
	body := fmt.Sprintf("<p><b>Subject:</b><br/>%s</p><p><b>Body:</b><br/>%s</p>", s, b)
	return body
}

func getFirstLine(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return "<< empty incoming text >>"
}
