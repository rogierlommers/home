package quicknote

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

func NewQuicknote(router *gin.Engine, cfg config.AppConfig, m *mailer.Mailer, stats *sqlitedb.DB) {
	router.POST("/api/notes/send", sendMailHandler(m, cfg, stats))
}

func sendMailHandler(m *mailer.Mailer, cfg config.AppConfig, stats *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var (
			title         string
			fileOnDisk    string
			targetEmail   string
			hasAttachment bool
		)

		// read "x-input-type" header
		inputType := c.GetHeader("x-input-type")
		logrus.Debugf("Input type: %s", inputType)

		switch inputType {

		// handle text input
		case "text":

			logrus.Debug("Handling text input")

			var jsonData struct {
				Text string `json:"text"`
			}

			if err := c.BindJSON(&jsonData); err != nil {
				logrus.Errorf("error binding json: %s", err)
				c.JSON(400, gin.H{"msg": "error binding json"})
				return
			}

			title = jsonData.Text
			hasAttachment = false

		// handle file input
		case "file":

			logrus.Info("Handling file input")

			filename := c.GetHeader("X-filename")
			targetEmail = determineTargetEmail(filename)

			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logrus.Errorf("error reading request body: %s", err)
				c.JSON(400, gin.H{"msg": "error reading request body"})
				return
			}

			fileOnDisk = path.Join(cfg.UploadTarget, filename)
			err = os.WriteFile(fileOnDisk, bodyBytes, 0777)
			if err != nil {
				c.JSON(500, gin.H{"msg": "error writing temp file"})
				return
			}

			title = filename
			hasAttachment = true
		}

		// start sending email now
		subject := fmt.Sprintf("Quicknote: %s", title)
		body := fmt.Sprintf("Quicknote received:\n\n%s", subject)

		var (
			statsSource     string
			responseMessage string
		)

		if hasAttachment {
			if err := m.SendMail(subject, targetEmail, body, []string{path.Base(fileOnDisk)}); err != nil {
				logrus.Errorf("sendMail error: %s", err)
				c.JSON(500, gin.H{"msg": fmt.Sprintf("error: mail error: %s", err)})
				return
			}

			statsSource = "quicknotes_with_attachment"
			responseMessage = fmt.Sprintf("(%s) note with attachment %s sent", humanize.Bytes(getSizeInUint64(fileOnDisk)), path.Base(fileOnDisk))

		} else {
			if err := m.SendMail(subject, targetEmail, body, nil); err != nil {
				logrus.Errorf("sendMail error: %s", err)
				c.JSON(500, gin.H{"msg": fmt.Sprintf("error: mail error: %s", err)})
				return
			}

			statsSource = "quicknotes_no_attachment"
			responseMessage = fmt.Sprintf("(%s) note without attachment sent", humanize.Bytes(uint64(len(body))))

		}

		// increment stats
		if stats.IncrementEntry(statsSource) != nil {
			logrus.Errorf("failed to increment quicknotes stat")
		}

		c.JSON(200, gin.H{"msg": responseMessage})
	}
}

// determine target email based on subject. Is subject starts with "w ",
// then use work email, otherwise personal email
func determineTargetEmail(s string) string {
	subject := strings.ToLower(s)

	if len(subject) >= 2 && subject[:2] == "w " {
		return mailer.WorkMail
	}

	return mailer.PrivateMail
}

func getSizeInUint64(filename string) uint64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		logrus.Errorf("error getting file info: %s", err)
		return 0
	}

	return uint64(fileInfo.Size())
}
