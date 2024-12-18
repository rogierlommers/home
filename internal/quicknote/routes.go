package quicknote

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// response type to frontend
type response struct {
	Msg string `json:"msg"`
}

func sendMailHandler(c *gin.Context) {
	var buf *bytes.Buffer
	var fileAttached bool

	// read attachment; pure text will be added as .txt file
	file, header, err := c.Request.FormFile("file")
	switch err {

	case nil:
		defer file.Close()
		fileAttached = true

		// read file into buffer
		buf = bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			logrus.Errorf("error reading file into buffer: %s", err)
			return
		}

	case http.ErrMissingFile:
		fileAttached = false

	default:
		logrus.Errorf("error parsing formFile: %s", err)
		return

	}

	// get optional text passed as text header
	optionalText := c.Request.FormValue("text")

	var (
		subjectFilename string
		fileContents    []byte
	)

	if fileAttached {
		logrus.Debugf("file: %s, size: %d", header.Filename, header.Size)
		logrus.Debugf("optional-text: %s", optionalText)
		subjectFilename = header.Filename
		fileContents = buf.Bytes()
	}

	// mail the file
	if err := sendMail(subjectFilename, fileContents, optionalText, fileAttached); err != nil {
		logrus.Errorf("sendMail error: %s", err)
		c.JSON(http.StatusInternalServerError, response{Msg: fmt.Sprintf("error: mail error: %s", err.Error())})
		return
	}

	// write happy flow response
	c.JSON(200, response{Msg: "email succesfully sent"})
}
