package quicknote

import (
	"bytes"
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

		// request only contains bytes as attachment
		// pure text will be added as .txt file

		var memoryBuffer = bytes.NewBuffer(nil)

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			logrus.Errorf("error parsing formFile: %s", err)
			c.JSON(400, gin.H{"msg": "error parsing formFile"})
			return
		}
		defer file.Close()

		// read file into buffer
		memoryBuffer = bytes.NewBuffer(nil)
		if _, err := io.Copy(memoryBuffer, file); err != nil {
			logrus.Errorf("error reading file into buffer: %s", err)
			c.JSON(500, gin.H{"msg": "error reading file into buffer"})
			return
		}

		// process text, strip extention in case of .txt file
		// and only send the plain filename as subject
		var (
			subject       string
			hasAttachment bool
			tmpFilename   string
		)

		if header.Filename[len(header.Filename)-4:] == ".txt" {
			subject = header.Filename[:len(header.Filename)-4]
			hasAttachment = false
		} else {
			subject = header.Filename
			hasAttachment = true
			tmpFilename = path.Join(cfg.UploadTarget, header.Filename)

			err := os.WriteFile(tmpFilename, memoryBuffer.Bytes(), 0777)
			if err != nil {
				c.JSON(500, gin.H{"msg": "error writing temp file"})
				return
			}
		}

		targetEmail := determineTargetEmail(subject)

		logrus.Debugf("subject: %s, file: %s, tempFilename: %s", subject, header.Filename, tmpFilename)
		body := fmt.Sprintf("Quicknote received:\n\n%s", subject)

		var (
			statsSource     string
			responseMessage string
		)

		if hasAttachment {
			if err := m.SendMail(subject, targetEmail, body, []string{header.Filename}); err != nil {
				logrus.Errorf("sendMail error: %s", err)
				c.JSON(500, gin.H{"msg": fmt.Sprintf("error: mail error: %s", err)})
				return
			}

			statsSource = "quicknotes_with_attachment"
			responseMessage = fmt.Sprintf("(%s) note with attachment %s sent", humanize.Bytes(uint64(len(memoryBuffer.Bytes()))), header.Filename)

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
