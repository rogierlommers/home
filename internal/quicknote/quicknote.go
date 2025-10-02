package quicknote

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/stats"
	"github.com/sirupsen/logrus"
)

func NewQuicknote(router *gin.Engine, cfg config.AppConfig, m *mailer.Mailer, stats *stats.DB) {
	router.POST("/api/notes/send", sendMailHandler(m, cfg, stats))
}

func sendMailHandler(m *mailer.Mailer, cfg config.AppConfig, stats *stats.DB) gin.HandlerFunc {
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

		logrus.Debugf("subject: %s, file: %s, tempFilename: %s", subject, header.Filename, tmpFilename)
		body := fmt.Sprintf("Quicknote received:\n\n%s", subject)

		var (
			statsSource     string
			responseMessage string
		)

		if hasAttachment {
			if err := m.SendMail(subject, mailer.PrivateMail, body, []string{header.Filename}); err != nil {
				logrus.Errorf("sendMail error: %s", err)
				c.JSON(500, gin.H{"msg": fmt.Sprintf("error: mail error: %s", err)})
				return
			}
			statsSource = "quicknotes_with_attachment"
			responseMessage = fmt.Sprintf("(%d bytes) note with attachment %s sent", len(memoryBuffer.Bytes()), header.Filename)
		} else {
			if err := m.SendMail(subject, mailer.PrivateMail, body, nil); err != nil {
				logrus.Errorf("sendMail error: %s", err)
				c.JSON(500, gin.H{"msg": fmt.Sprintf("error: mail error: %s", err)})
				return
			}
			statsSource = "quicknotes_no_attachment"
			responseMessage = fmt.Sprintf("(%d bytes) note without attachment sent", len(body))
		}

		// increment stats
		if stats.IncrementEntry(statsSource) != nil {
			logrus.Errorf("failed to increment quicknotes stat")
		}

		c.JSON(200, gin.H{"msg": responseMessage})
	}
}
