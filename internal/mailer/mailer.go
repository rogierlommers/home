package mailer

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/rogierlommers/home/internal/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

const (
	PrivateMail = "private"
	WorkMail    = "work"
)

type Mailer struct {
	targetEmailPrivate string
	targetEmailWork    string
	fromEmail          string
	smtpHost           string
	smtpUsername       string
	smtpPassword       string
	smtpPort           int
	cfg                config.AppConfig
}

func NewMailer(cfg config.AppConfig) *Mailer {
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

	return &Mailer{
		targetEmailPrivate: targetEmailPrivate,
		targetEmailWork:    targetEmailWork,
		fromEmail:          fromEmail,
		smtpHost:           smtpHost,
		smtpUsername:       smtpUsername,
		smtpPassword:       smtpPassword,
		smtpPort:           smtpPort,
		cfg:                cfg,
	}
}

func (m *Mailer) SendMail(subject string, target string, body string, attachments []string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", m.fromEmail)

	// determine recipient based on target
	switch strings.ToLower(target) {
	case PrivateMail:
		mailer.SetHeader("To", m.targetEmailPrivate)
	case WorkMail:
		mailer.SetHeader("To", m.targetEmailWork)
	default:
		mailer.SetHeader("To", m.targetEmailPrivate)
	}

	// add attachments
	for _, v := range attachments {
		incomingFile := path.Join(m.cfg.UploadTarget, v)
		checkIfFileExists(incomingFile)
		mailer.Attach(incomingFile)
	}

	mailer.SetHeader("Subject", fmt.Sprintf("☑️ %s", subject))
	mailer.SetBody("text/html", defineBody(subject, body))

	d := gomail.NewDialer(m.smtpHost, m.smtpPort, m.smtpUsername, m.smtpPassword)
	d.SSL = false
	if err := d.DialAndSend(mailer); err != nil {
		return err
	}

	return nil
}

func defineBody(s string, b string) string {
	logrus.Debugf("incoming subject: %s", s)
	logrus.Debugf("incoming body: %s", b)

	b = strings.ReplaceAll(b, "\n", "<br/>")
	return b
}

func checkIfFileExists(v string) {
	if _, err := os.Stat(v); os.IsNotExist(err) {
		logrus.Errorf("file does not exist: %s", v)
	} else {
		logrus.Debugf("file exists: %s", v)
	}
}
